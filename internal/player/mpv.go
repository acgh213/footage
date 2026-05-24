//go:build windows

package player

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/Microsoft/go-winio"
)

const pipeName = `\\.\pipe\footage-mpv`

// Player manages the mpv subprocess and its IPC connection.
type Player struct {
	mu      sync.Mutex
	cmd     *exec.Cmd
	ipc     *ipc
	mpvPath string
}

// New creates a Player. mpvPath is optional; if empty, the path is resolved
// from PATH and well-known bundled locations.
func New(mpvPath string) *Player {
	return &Player{mpvPath: resolveMPVPath(mpvPath)}
}

// Start launches mpv with the given video file and connects to its IPC pipe.
// If mpv is already running, it is stopped first.
func (p *Player) Start(videoPath string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.stopLocked()

	if p.mpvPath == "" {
		return fmt.Errorf(
			"mpv not found — install mpv and add it to PATH, " +
				"or place mpv.exe in assets/mpv/ and run scripts/get-mpv.ps1",
		)
	}

	args := []string{
		"--input-ipc-server=" + pipeName,
		"--idle=yes",
		"--keep-open=yes",
		videoPath,
	}
	p.cmd = exec.Command(p.mpvPath, args...)
	// Detach from our console so mpv opens its own window.
	p.cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	}

	if err := p.cmd.Start(); err != nil {
		p.cmd = nil
		return fmt.Errorf("start mpv: %w", err)
	}

	conn, err := pollForPipe(pipeName, 5*time.Second)
	if err != nil {
		_ = p.cmd.Process.Kill()
		p.cmd = nil
		return fmt.Errorf("mpv IPC not ready: %w", err)
	}

	// Brief pause so mpv finishes its startup sequence before we start
	// sending commands. Without this, early queries can return stale state.
	time.Sleep(300 * time.Millisecond)

	p.ipc = newIPC(conn)
	return nil
}

// Stop sends quit to mpv and waits for the process to exit.
func (p *Player) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.stopLocked()
}

func (p *Player) stopLocked() error {
	if p.ipc != nil {
		_, _ = p.ipc.send("quit")
		p.ipc.close()
		p.ipc = nil
	}
	if p.cmd != nil && p.cmd.Process != nil {
		done := make(chan error, 1)
		go func() { done <- p.cmd.Wait() }()
		select {
		case <-done:
		case <-time.After(2 * time.Second):
			_ = p.cmd.Process.Kill()
			<-done
		}
		p.cmd = nil
	}
	return nil
}

// ResolvedPath returns the mpv binary path the player resolved to, or "" if
// none was found. Used by the frontend for diagnostics.
func (p *Player) ResolvedPath() string {
	return p.mpvPath
}

// IsRunning reports whether mpv is currently running.
func (p *Player) IsRunning() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.ipc != nil
}

// GetTimePos returns the current playback position in seconds.
// Tries time-pos first (universally supported), then time-position (alias).
// Returns 0, nil when mpv is loaded but position is not yet available.
func (p *Player) GetTimePos() (float64, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.ipc == nil {
		return 0, fmt.Errorf("mpv is not running")
	}
	v, err := p.ipc.getFloat("get_property", "time-pos")
	if err != nil {
		// Fall back to the alternate property name.
		v, err = p.ipc.getFloat("get_property", "time-position")
	}
	if err != nil {
		// "property unavailable" just means no file is loaded / seeking.
		// Treat it as 0 rather than a hard error.
		errStr := err.Error()
		if strings.Contains(errStr, "unavailable") || strings.Contains(errStr, "not found") {
			return 0, nil
		}
		return 0, err
	}
	return v, nil
}

// GetProperty queries any named mpv property and returns it as a string.
// Useful for diagnostics.
func (p *Player) GetProperty(name string) (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.ipc == nil {
		return "", fmt.Errorf("mpv is not running")
	}
	data, err := p.ipc.send("get_property", name)
	if err != nil {
		return "", err
	}
	// Strip surrounding quotes if mpv returned a JSON string.
	s := strings.Trim(string(data), `"`)
	return s, nil
}

// Seek seeks relative or absolute by delta seconds.
func (p *Player) Seek(delta float64, relative bool) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.ipc == nil {
		return fmt.Errorf("mpv is not running")
	}
	mode := "absolute"
	if relative {
		mode = "relative"
	}
	_, err := p.ipc.send("seek", delta, mode)
	return err
}

// SetSpeed sets the playback speed multiplier.
func (p *Player) SetSpeed(s float64) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.ipc == nil {
		return fmt.Errorf("mpv is not running")
	}
	_, err := p.ipc.send("set_property", "speed", s)
	return err
}

// Pause pauses playback.
func (p *Player) Pause() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.ipc == nil {
		return fmt.Errorf("mpv is not running")
	}
	_, err := p.ipc.send("set_property", "pause", true)
	return err
}

// Play resumes playback.
func (p *Player) Play() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.ipc == nil {
		return fmt.Errorf("mpv is not running")
	}
	_, err := p.ipc.send("set_property", "pause", false)
	return err
}

// pollForPipe polls until the named pipe is available or the timeout elapses.
func pollForPipe(name string, timeout time.Duration) (net.Conn, error) {
	deadline := time.Now().Add(timeout)
	dialTimeout := 500 * time.Millisecond
	for time.Now().Before(deadline) {
		conn, err := winio.DialPipe(name, &dialTimeout)
		if err == nil {
			return conn, nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return nil, fmt.Errorf("timed out waiting for mpv named pipe %s", name)
}

// resolveMPVPath finds the mpv binary. Search order:
//  1. user-configured path from config.json
//  2. LookPath for common executable names (mpv, mpvnet, mpv.exe, mpvnet.exe)
//  3. cmd.exe /C where — catches cases where GUI process PATH != console PATH
//  4. WinGet packages directory scan (covers mpv.net WinGet installs)
//  5. Common fixed install locations
//  6. assets/mpv/mpv.exe relative to the running executable (bundled)
func resolveMPVPath(configured string) string {
	if configured != "" {
		if abs, err := filepath.Abs(configured); err == nil {
			if _, err := os.Stat(abs); err == nil {
				return abs
			}
		}
	}

	// Try LookPath with all common names for mpv on Windows.
	for _, name := range []string{"mpv", "mpv.exe", "mpvnet", "mpvnet.exe"} {
		if path, err := exec.LookPath(name); err == nil {
			return path
		}
	}

	// cmd.exe /C where — the Windows shell can find things the Go process misses
	// when PATH was updated after the process launched (e.g. WinGet install).
	for _, name := range []string{"mpv", "mpvnet"} {
		out, err := exec.Command("cmd.exe", "/C", "where", name).Output()
		if err == nil {
			if line := strings.SplitN(strings.TrimSpace(string(out)), "\n", 2)[0]; line != "" {
				line = strings.TrimSpace(line)
				if _, err := os.Stat(line); err == nil {
					return line
				}
			}
		}
	}

	// Scan the WinGet packages directory for any mpv.exe or mpvnet.exe.
	// WinGet installs to %LOCALAPPDATA%\Microsoft\WinGet\Packages\{id}_{source}\{version}\
	if localAppData := os.Getenv("LOCALAPPDATA"); localAppData != "" {
		wingetPkgs := filepath.Join(localAppData, "Microsoft", "WinGet", "Packages")
		for _, name := range []string{"mpv.exe", "mpvnet.exe"} {
			// Walk one level: Packages/<pkg>/<version>/<exe>
			pattern := filepath.Join(wingetPkgs, "*", "*", name)
			if matches, err := filepath.Glob(pattern); err == nil && len(matches) > 0 {
				return matches[0]
			}
			// Some packages land without a version subdirectory.
			pattern = filepath.Join(wingetPkgs, "*", name)
			if matches, err := filepath.Glob(pattern); err == nil && len(matches) > 0 {
				return matches[0]
			}
		}
		// Also check WinGet links directory (shimmed executables).
		wingetLinks := filepath.Join(localAppData, "Microsoft", "WinGet", "Links")
		for _, name := range []string{"mpv.exe", "mpvnet.exe"} {
			p := filepath.Join(wingetLinks, name)
			if _, err := os.Stat(p); err == nil {
				return p
			}
		}
	}

	// Common fixed install paths.
	for _, p := range []string{
		`C:\Program Files\mpv\mpv.exe`,
		`C:\Program Files\mpv.net\mpv.exe`,
		`C:\Program Files\mpv.net\mpvnet.exe`,
		`C:\mpv\mpv.exe`,
	} {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}

	// Bundled copy relative to the running executable.
	exe, err := os.Executable()
	if err != nil {
		return ""
	}
	exeDir := filepath.Dir(exe)
	for _, rel := range []string{
		filepath.Join("assets", "mpv", "mpv.exe"),
		filepath.Join("..", "..", "assets", "mpv", "mpv.exe"),
		filepath.Join("..", "..", "..", "assets", "mpv", "mpv.exe"),
	} {
		p := filepath.Join(exeDir, rel)
		if abs, err := filepath.Abs(p); err == nil {
			if _, err := os.Stat(abs); err == nil {
				return abs
			}
		}
	}
	return ""
}
