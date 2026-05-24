//go:build windows

package player

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
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

// IsRunning reports whether mpv is currently running.
func (p *Player) IsRunning() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.ipc != nil
}

// GetTimePos returns the current playback position in seconds.
func (p *Player) GetTimePos() (float64, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.ipc == nil {
		return 0, fmt.Errorf("mpv is not running")
	}
	return p.ipc.getFloat("get_property", "time-position")
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

// resolveMPVPath finds the mpv binary in priority order:
//  1. user-configured path (from config.json)
//  2. PATH
//  3. assets/mpv/mpv.exe relative to this executable
//  4. assets/mpv/mpv.exe two levels up (wails dev mode)
func resolveMPVPath(configured string) string {
	if configured != "" {
		if _, err := os.Stat(configured); err == nil {
			return configured
		}
	}
	if path, err := exec.LookPath("mpv"); err == nil {
		return path
	}
	exe, err := os.Executable()
	if err != nil {
		return ""
	}
	exeDir := filepath.Dir(exe)
	candidates := []string{
		filepath.Join(exeDir, "assets", "mpv", "mpv.exe"),
		filepath.Join(exeDir, "..", "..", "assets", "mpv", "mpv.exe"),
		filepath.Join(exeDir, "..", "..", "..", "assets", "mpv", "mpv.exe"),
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	return ""
}
