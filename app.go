package main

import (
	"context"
	"fmt"

	"github.com/acgh213/footage/internal/config"
	"github.com/acgh213/footage/internal/player"
	"github.com/acgh213/footage/internal/preset"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App is the main application struct. All exported methods are bound to the
// Wails webview and callable from the frontend.
type App struct {
	ctx    context.Context
	config *config.Config
	player *player.Player
}

func newApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	cfg, err := config.Load()
	if err != nil {
		cfg = config.Default()
	}
	a.config = cfg
	preset.EnsureBuiltins()
	a.player = player.New(a.config.MPVPath)
}

func (a *App) shutdown(_ context.Context) {
	if a.player != nil {
		_ = a.player.Stop()
	}
}

// BrowseForFile opens the native file picker and returns the selected path.
func (a *App) BrowseForFile() (string, error) {
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "select video file",
		Filters: []runtime.FileFilter{
			{DisplayName: "video files", Pattern: "*.mp4;*.mkv;*.avi;*.webm;*.mov;*.wmv;*.m4v"},
			{DisplayName: "all files", Pattern: "*"},
		},
	})
	return path, err
}

// OpenFile opens a video file in mpv.
func (a *App) OpenFile(path string) error {
	if path == "" {
		return fmt.Errorf("no file path provided")
	}
	return a.player.Start(path)
}

// GetTimePos returns the current playback position in seconds.
func (a *App) GetTimePos() (float64, error) {
	return a.player.GetTimePos()
}

// StopPlayer stops the mpv process.
func (a *App) StopPlayer() error {
	return a.player.Stop()
}

// GetPresets returns all available tag presets.
func (a *App) GetPresets() ([]preset.Preset, error) {
	return preset.ListAll()
}

// GetMPVStatus returns true if mpv is currently running.
func (a *App) GetMPVStatus() bool {
	return a.player.IsRunning()
}

// GetMPVPath returns the resolved mpv binary path, or an empty string if mpv
// was not found. Used by the UI to show a diagnostic.
func (a *App) GetMPVPath() string {
	return a.player.ResolvedPath()
}

// BrowseForMPV opens a file picker so the user can manually locate mpv.exe.
// The selected path is saved to config and the player is re-initialized.
func (a *App) BrowseForMPV() (string, error) {
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "locate mpv.exe",
		Filters: []runtime.FileFilter{
			{DisplayName: "executables", Pattern: "*.exe"},
			{DisplayName: "all files", Pattern: "*"},
		},
	})
	if err != nil || path == "" {
		return "", err
	}
	a.config.MPVPath = path
	_ = config.Save(a.config)
	a.player = player.New(path)
	return path, nil
}
