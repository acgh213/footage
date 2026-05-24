package region

import "time"

// Kind distinguishes log entries.
type Kind string

const (
	KindRegion   Kind = "region"
	KindBookmark Kind = "bookmark"
)

// Region is a tagged time span within a video.
type Region struct {
	ID           string    `json:"id"`
	Kind         Kind      `json:"kind"`
	VideoPath    string    `json:"video_path"`
	SessionID    string    `json:"session_id"`
	TagKey       string    `json:"tag_key,omitempty"`
	TagLabel     string    `json:"tag_label,omitempty"`
	TagColor     string    `json:"tag_color,omitempty"`
	StartSec     float64   `json:"start_sec"`
	EndSec       float64   `json:"end_sec,omitempty"`
	Notes        string    `json:"notes,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Bookmark is a single-point annotation (no end time).
type Bookmark struct {
	ID        string    `json:"id"`
	Kind      Kind      `json:"kind"` // always "bookmark"
	VideoPath string    `json:"video_path"`
	SessionID string    `json:"session_id"`
	TimeSec   float64   `json:"time_sec"`
	Notes     string    `json:"notes,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// InProgressRegion is a region that has been opened but not yet closed.
// Returned by the toggle for display in the UI before the region is committed.
type InProgressRegion struct {
	TagKey   string  `json:"tag_key"`
	TagLabel string  `json:"tag_label"`
	TagColor string  `json:"tag_color"`
	StartSec float64 `json:"start_sec"`
}

// Entry is either a Region or Bookmark stored in the manifest.
// Exactly one of Region/Bookmark is set; the other is nil.
type Entry struct {
	Region   *Region   `json:"region,omitempty"`
	Bookmark *Bookmark `json:"bookmark,omitempty"`
}

// ID returns the entry's stable ULID-based ID.
func (e Entry) ID() string {
	if e.Region != nil {
		return e.Region.ID
	}
	return e.Bookmark.ID
}

// Kind returns the entry's kind.
func (e Entry) EntryKind() Kind {
	if e.Region != nil {
		return e.Region.Kind
	}
	return e.Bookmark.Kind
}

// VideoPath returns the video path for the entry.
func (e Entry) VideoPath() string {
	if e.Region != nil {
		return e.Region.VideoPath
	}
	return e.Bookmark.VideoPath
}
