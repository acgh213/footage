package region

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// rawEntry is the union shape stored in JSONL. One of region/bookmark fields
// is populated based on "kind".
type rawEntry struct {
	ID        string          `json:"id"`
	Kind      Kind            `json:"kind"`
	VideoPath string          `json:"video_path"`
	SessionID string          `json:"session_id"`
	// region fields
	TagKey    string          `json:"tag_key,omitempty"`
	TagLabel  string          `json:"tag_label,omitempty"`
	TagColor  string          `json:"tag_color,omitempty"`
	StartSec  float64         `json:"start_sec,omitempty"`
	EndSec    float64         `json:"end_sec,omitempty"`
	// bookmark fields
	TimeSec   float64         `json:"time_sec,omitempty"`
	Notes     string          `json:"notes,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
	// tombstone
	Deleted   bool            `json:"deleted,omitempty"`
	RawExtra  json.RawMessage `json:"-"`
}

func toEntry(r rawEntry) Entry {
	if r.Kind == KindBookmark {
		return Entry{Bookmark: &Bookmark{
			ID: r.ID, Kind: KindBookmark,
			VideoPath: r.VideoPath, SessionID: r.SessionID,
			TimeSec: r.TimeSec, Notes: r.Notes,
			CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
		}}
	}
	return Entry{Region: &Region{
		ID: r.ID, Kind: KindRegion,
		VideoPath: r.VideoPath, SessionID: r.SessionID,
		TagKey: r.TagKey, TagLabel: r.TagLabel, TagColor: r.TagColor,
		StartSec: r.StartSec, EndSec: r.EndSec,
		Notes: r.Notes, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
	}}
}

func fromRegion(r *Region) rawEntry {
	return rawEntry{
		ID: r.ID, Kind: KindRegion,
		VideoPath: r.VideoPath, SessionID: r.SessionID,
		TagKey: r.TagKey, TagLabel: r.TagLabel, TagColor: r.TagColor,
		StartSec: r.StartSec, EndSec: r.EndSec,
		Notes: r.Notes, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
	}
}

func fromBookmark(b *Bookmark) rawEntry {
	return rawEntry{
		ID: b.ID, Kind: KindBookmark,
		VideoPath: b.VideoPath, SessionID: b.SessionID,
		TimeSec: b.TimeSec, Notes: b.Notes,
		CreatedAt: b.CreatedAt, UpdatedAt: b.UpdatedAt,
	}
}

// ManifestPath returns the path to the manifest file for a session.
// Stored alongside the session's JSON file.
func ManifestPath(sessionID string) (string, error) {
	cfg, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(cfg, "Footage", "sessions")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return filepath.Join(dir, sessionID+".jsonl"), nil
}

// AppendRegion appends a region entry to the manifest. Fast path for logging.
func AppendRegion(sessionID string, r *Region) error {
	return appendRaw(sessionID, fromRegion(r))
}

// AppendBookmark appends a bookmark entry to the manifest.
func AppendBookmark(sessionID string, b *Bookmark) error {
	return appendRaw(sessionID, fromBookmark(b))
}

func appendRaw(sessionID string, r rawEntry) error {
	path, err := ManifestPath(sessionID)
	if err != nil {
		return err
	}
	line, err := json.Marshal(r)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = fmt.Fprintf(f, "%s\n", line)
	return err
}

// ReadAll parses all entries from a manifest file.
// Blank lines are skipped; malformed lines return an error.
func ReadAll(sessionID string) ([]Entry, error) {
	path, err := ManifestPath(sessionID)
	if err != nil {
		return nil, err
	}
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	var entries []Entry
	scanner := bufio.NewScanner(f)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var r rawEntry
		if err := json.Unmarshal([]byte(line), &r); err != nil {
			return nil, fmt.Errorf("manifest line %d: %w", lineNum, err)
		}
		if r.Deleted {
			continue
		}
		entries = append(entries, toEntry(r))
	}
	return entries, scanner.Err()
}

// Rewrite atomically replaces the manifest with the given entries.
func Rewrite(sessionID string, entries []Entry) error {
	path, err := ManifestPath(sessionID)
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(f)
	for _, e := range entries {
		var raw rawEntry
		if e.Region != nil {
			raw = fromRegion(e.Region)
		} else {
			raw = fromBookmark(e.Bookmark)
		}
		if err := enc.Encode(raw); err != nil {
			f.Close()
			os.Remove(tmp)
			return err
		}
	}
	f.Close()
	return os.Rename(tmp, path)
}

// UpdateNotes rewrites the manifest with updated notes for a given entry ID.
func UpdateNotes(sessionID, entryID, notes string) error {
	entries, err := ReadAll(sessionID)
	if err != nil {
		return err
	}
	now := time.Now()
	for i, e := range entries {
		if e.ID() == entryID {
			if e.Region != nil {
				entries[i].Region.Notes = notes
				entries[i].Region.UpdatedAt = now
			} else {
				entries[i].Bookmark.Notes = notes
				entries[i].Bookmark.UpdatedAt = now
			}
			break
		}
	}
	return Rewrite(sessionID, entries)
}

// DeleteEntry rewrites the manifest without the given entry ID.
func DeleteEntry(sessionID, entryID string) error {
	entries, err := ReadAll(sessionID)
	if err != nil {
		return err
	}
	filtered := entries[:0]
	for _, e := range entries {
		if e.ID() != entryID {
			filtered = append(filtered, e)
		}
	}
	return Rewrite(sessionID, filtered)
}
