package main

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/acgh213/footage/internal/config"
	"github.com/acgh213/footage/internal/export"
	"github.com/acgh213/footage/internal/player"
	"github.com/acgh213/footage/internal/preset"
	"github.com/acgh213/footage/internal/region"
	"github.com/acgh213/footage/internal/session"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App is the main application struct. All exported methods are bound to the
// Wails webview and callable from the frontend.
type App struct {
	ctx     context.Context
	config  *config.Config
	player  *player.Player
	session *session.Session
	toggle  *region.Toggle
}

func newApp() *App {
	a := &App{toggle: region.NewToggle()}
	a.initBackend()
	return a
}

// initBackend performs all initialization that does not require the Wails
// context. It runs in newApp() — before wails.Run() — so that a.config,
// a.player and a.session are guaranteed non-nil by the time the frontend can
// call any bound method.
//
// This used to live in startup() (OnStartup), which races the webview:
// resolveMPVPath() spawns several `cmd.exe /C where` subprocesses and can
// still be running when the frontend calls GetMPVPath. That dereferenced a
// nil a.player and panicked; Wails v2 does not reject the promise on a panic,
// so init() in the frontend hung on its first await and nothing loaded.
func (a *App) initBackend() {
	cfg, err := config.Load()
	if err != nil {
		cfg = config.Default()
	}
	a.config = cfg
	preset.EnsureBuiltins()
	a.player = player.New(a.config.MPVPath)

	// Restore last session or create a new one.
	if s, err := session.LoadLast(); err == nil && s != nil {
		a.session = s
	} else {
		a.session = session.New("default")
		_ = session.Save(a.session)
		_ = session.SavePointer(a.session.ID)
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) shutdown(_ context.Context) {
	if a.player != nil {
		_ = a.player.Stop()
	}
	if a.session != nil {
		_ = session.Save(a.session)
	}
}

// ── file / playback ──────────────────────────────────────────────────────────

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

// OpenFile opens a video file in mpv and adds it to the current session.
func (a *App) OpenFile(path string) error {
	if path == "" {
		return fmt.Errorf("no file path provided")
	}
	if err := a.player.Start(path); err != nil {
		return err
	}
	a.session.AddFile(path)
	for i, f := range a.session.Files {
		if f == path {
			a.session.ActiveIdx = i
			break
		}
	}
	_ = session.Save(a.session)
	_ = session.SavePointer(a.session.ID)
	return nil
}

// GetTimePos returns the current playback position in seconds.
func (a *App) GetTimePos() (float64, error) {
	return a.player.GetTimePos()
}

// StopPlayer stops the mpv process.
func (a *App) StopPlayer() error {
	return a.player.Stop()
}

// Seek seeks mpv by delta seconds (relative if relative=true).
func (a *App) Seek(delta float64, relative bool) error {
	return a.player.Seek(delta, relative)
}

// SetSpeed sets the playback speed multiplier.
func (a *App) SetSpeed(s float64) error {
	return a.player.SetSpeed(s)
}

// Pause pauses playback.
func (a *App) Pause() error {
	return a.player.Pause()
}

// Play resumes playback.
func (a *App) Play() error {
	return a.player.Play()
}

// ── presets ──────────────────────────────────────────────────────────────────

// GetPresets returns all available tag presets.
func (a *App) GetPresets() ([]preset.Preset, error) {
	return preset.ListAll()
}

// ── session ───────────────────────────────────────────────────────────────────

// GetSession returns the current session.
func (a *App) GetSession() *session.Session {
	return a.session
}

// NewSession creates a fresh session and saves a pointer to it.
func (a *App) NewSession(name string) (*session.Session, error) {
	if name == "" {
		name = "untitled"
	}
	s := session.New(name)
	if err := session.Save(s); err != nil {
		return nil, err
	}
	_ = session.SavePointer(s.ID)
	a.session = s
	return s, nil
}

// ListSessions returns all sessions sorted by last update.
func (a *App) ListSessions() ([]*session.Session, error) {
	return session.List()
}

// LoadSession switches to the session with the given ID.
func (a *App) LoadSession(id string) (*session.Session, error) {
	s, err := session.Load(id)
	if err != nil {
		return nil, err
	}
	a.session = s
	_ = session.SavePointer(s.ID)
	return s, nil
}

// AddFileToSession adds a video file to the current session without opening it.
func (a *App) AddFileToSession(path string) (*session.Session, error) {
	if path == "" {
		return nil, fmt.Errorf("no path provided")
	}
	a.session.AddFile(path)
	if err := session.Save(a.session); err != nil {
		return nil, err
	}
	return a.session, nil
}

// RemoveFileFromSession removes a video from the session file list.
func (a *App) RemoveFileFromSession(path string) (*session.Session, error) {
	a.toggle.CancelAll(path)
	a.session.RemoveFile(path)
	if err := session.Save(a.session); err != nil {
		return nil, err
	}
	return a.session, nil
}

// SetActiveFile switches the active file (by index) and opens it in mpv.
func (a *App) SetActiveFile(idx int) error {
	a.session.SetActive(idx)
	_ = session.Save(a.session)
	path := a.session.ActiveFile()
	if path == "" {
		return nil
	}
	return a.player.Start(path)
}

// ── regions ───────────────────────────────────────────────────────────────────

// PressTag handles a hotkey press for a tag. Returns the closed region if one
// was completed, or nil if a new region was opened.
func (a *App) PressTag(tagKey, tagLabel, tagColor string) (*region.Region, error) {
	if a.session == nil {
		return nil, fmt.Errorf("no active session")
	}
	videoPath := a.session.ActiveFile()
	if videoPath == "" {
		return nil, fmt.Errorf("no active video")
	}
	nowSec, err := a.player.GetTimePos()
	if err != nil {
		return nil, fmt.Errorf("get time: %w", err)
	}
	r, closed := a.toggle.Press(a.session.ID, videoPath, tagKey, tagLabel, tagColor, nowSec)
	if closed {
		if err := region.AppendRegion(a.session.ID, r); err != nil {
			return nil, fmt.Errorf("save region: %w", err)
		}
		return r, nil
	}
	return nil, nil
}

// AddBookmark adds a bookmark at the current playback position.
func (a *App) AddBookmark() (*region.Bookmark, error) {
	if a.session == nil {
		return nil, fmt.Errorf("no active session")
	}
	videoPath := a.session.ActiveFile()
	if videoPath == "" {
		return nil, fmt.Errorf("no active video")
	}
	nowSec, err := a.player.GetTimePos()
	if err != nil {
		return nil, fmt.Errorf("get time: %w", err)
	}
	b := &region.Bookmark{
		ID:        region.NewBookmarkID(),
		Kind:      region.KindBookmark,
		VideoPath: videoPath,
		SessionID: a.session.ID,
		TimeSec:   nowSec,
	}
	if err := region.AppendBookmark(a.session.ID, b); err != nil {
		return nil, err
	}
	return b, nil
}

// GetRegions returns all entries for the current session.
func (a *App) GetRegions() ([]region.Entry, error) {
	if a.session == nil {
		return nil, fmt.Errorf("no active session")
	}
	return region.ReadAll(a.session.ID)
}

// GetOpenTags returns the tag keys that have open regions for the active video.
func (a *App) GetOpenTags() []string {
	if a.session == nil {
		return nil
	}
	return a.toggle.OpenTags(a.session.ActiveFile())
}

// UpdateNotes saves notes for a region or bookmark by ID.
func (a *App) UpdateNotes(entryID, notes string) error {
	if a.session == nil {
		return fmt.Errorf("no active session")
	}
	return region.UpdateNotes(a.session.ID, entryID, notes)
}

// DeleteRegion removes a region or bookmark from the manifest.
func (a *App) DeleteRegion(entryID string) error {
	if a.session == nil {
		return fmt.Errorf("no active session")
	}
	return region.DeleteEntry(a.session.ID, entryID)
}

// SeekToEntry seeks mpv to the start of a region (or time of a bookmark).
func (a *App) SeekToEntry(entryID string) error {
	entries, err := region.ReadAll(a.session.ID)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if e.ID() == entryID {
			var t float64
			if e.Region != nil {
				t = e.Region.StartSec
			} else {
				t = e.Bookmark.TimeSec
			}
			return a.player.Seek(t, false)
		}
	}
	return fmt.Errorf("entry %s not found", entryID)
}

// ── export ────────────────────────────────────────────────────────────────────

// ExportRegion remuxes a single region to the exports/ directory beside the video.
func (a *App) ExportRegion(entryID string) (string, error) {
	if a.session == nil {
		return "", fmt.Errorf("no active session")
	}
	entries, err := region.ReadAll(a.session.ID)
	if err != nil {
		return "", err
	}
	for _, e := range entries {
		if e.ID() != entryID || e.Region == nil {
			continue
		}
		r := e.Region
		outDir := filepath.Join(filepath.Dir(r.VideoPath), "exports")
		outFile := filepath.Join(outDir, export.Filename(r.VideoPath, r.StartSec, r.TagKey))
		opt := export.Options{
			VideoPath:  r.VideoPath,
			StartSec:   r.StartSec,
			EndSec:     r.EndSec,
			OutputPath: outFile,
			FFmpegPath: a.config.FFmpegPath,
		}
		if err := export.Remux(context.Background(), opt); err != nil {
			return "", err
		}
		return outFile, nil
	}
	return "", fmt.Errorf("region %s not found", entryID)
}

// GetInProgressRegions returns regions currently being recorded for the active video.
// Used by the frontend to show live in-progress entries before they are closed.
func (a *App) GetInProgressRegions() []region.InProgressRegion {
	if a.session == nil {
		return nil
	}
	videoPath := a.session.ActiveFile()
	if videoPath == "" {
		return nil
	}
	infos := a.toggle.InProgress(videoPath)
	result := make([]region.InProgressRegion, len(infos))
	for i, info := range infos {
		result[i] = region.InProgressRegion{
			TagKey:   info.TagKey,
			TagLabel: info.TagLabel,
			TagColor: info.TagColor,
			StartSec: info.StartSec,
		}
	}
	return result
}

// ── editing ───────────────────────────────────────────────────────────────────

// NudgeRegion adjusts the start or end of a region by delta seconds.
// field must be "start" or "end".
func (a *App) NudgeRegion(entryID, field string, delta float64) error {
	if a.session == nil {
		return fmt.Errorf("no active session")
	}
	return region.Nudge(a.session.ID, entryID, field, delta)
}

// SetRegionTime sets the start or end of a region to an absolute value.
func (a *App) SetRegionTime(entryID, field string, value float64) error {
	if a.session == nil {
		return fmt.Errorf("no active session")
	}
	return region.Set(a.session.ID, entryID, field, value)
}

// MergeRegions merges two adjacent same-tag regions into one.
func (a *App) MergeRegions(idA, idB string) error {
	if a.session == nil {
		return fmt.Errorf("no active session")
	}
	return region.Merge(a.session.ID, idA, idB)
}

// SearchRegions filters the current session's regions by the given query.
func (a *App) SearchRegions(tags []string, video, text string, minDur, maxDur float64) ([]region.Entry, error) {
	if a.session == nil {
		return nil, fmt.Errorf("no active session")
	}
	entries, err := region.ReadAll(a.session.ID)
	if err != nil {
		return nil, err
	}
	q := region.Query{
		Tags:         tags,
		Video:        video,
		ContainsText: text,
		MinDuration:  minDur,
		MaxDuration:  maxDur,
	}
	return region.Filter(entries, q), nil
}

// ── batch export ──────────────────────────────────────────────────────────────

// BatchExportProgress is a progress update sent during a batch export.
type BatchExportProgress struct {
	RegionID string `json:"region_id"`
	OutPath  string `json:"out_path,omitempty"`
	ErrMsg   string `json:"err,omitempty"`
	Done     bool   `json:"done"`
}

// BatchExport exports a list of region IDs to the given output directory.
// Progress events are emitted on the "batch-progress" Wails event channel.
// Returns the number of successfully exported clips.
func (a *App) BatchExport(regionIDs []string, outDir string) (int, error) {
	if a.session == nil {
		return 0, fmt.Errorf("no active session")
	}
	allEntries, err := region.ReadAll(a.session.ID)
	if err != nil {
		return 0, err
	}

	idSet := make(map[string]bool, len(regionIDs))
	for _, id := range regionIDs {
		idSet[id] = true
	}

	var regions []region.Region
	for _, e := range allEntries {
		if e.Region != nil && idSet[e.Region.ID] {
			regions = append(regions, *e.Region)
		}
	}

	if outDir == "" {
		// Default to exports/ beside the first region's video.
		if len(regions) > 0 {
			outDir = filepath.Join(filepath.Dir(regions[0].VideoPath), "exports")
		} else {
			return 0, fmt.Errorf("no matching regions found")
		}
	}

	succeeded := 0
	export.RunBatch(context.Background(), regions, outDir, a.config.FFmpegPath,
		func(r export.BatchResult) {
			msg := ""
			if r.Err != nil {
				msg = r.Err.Error()
			} else {
				succeeded++
			}
			runtime.EventsEmit(a.ctx, "batch-progress", BatchExportProgress{
				RegionID: r.RegionID,
				OutPath:  r.OutPath,
				ErrMsg:   msg,
				Done:     r.Err == nil,
			})
		},
	)
	return succeeded, nil
}

// BrowseForExportDir opens a folder picker for choosing a batch export directory.
func (a *App) BrowseForExportDir() (string, error) {
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "choose export folder",
	})
}

// ── diagnostics ──────────────────────────────────────────────────────────────

// GetMPVStatus returns true if mpv is currently running.
func (a *App) GetMPVStatus() bool {
	return a.player.IsRunning()
}

// GetMPVPath returns the resolved mpv binary path, or an empty string if mpv
// was not found.
func (a *App) GetMPVPath() string {
	return a.player.ResolvedPath()
}

// TestIPC queries mpv-version over IPC and returns the raw string.
func (a *App) TestIPC() (string, error) {
	return a.player.GetProperty("mpv-version")
}

// BrowseForMPV opens a file picker so the user can manually locate mpv.exe.
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
