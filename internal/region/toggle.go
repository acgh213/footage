package region

import (
	"sync"
	"time"
)

// openRegion tracks an in-progress region that hasn't been closed yet.
type openRegion struct {
	id        string
	tagKey    string
	tagLabel  string
	tagColor  string
	videoPath string
	startSec  float64
	openedAt  time.Time
}

// Toggle manages the one-open-region-per-tag-per-video invariant.
// Safe for concurrent use.
type Toggle struct {
	mu   sync.Mutex
	open map[string]*openRegion // key: videoPath+"\x00"+tagKey
}

func NewToggle() *Toggle {
	return &Toggle{open: make(map[string]*openRegion)}
}

func key(videoPath, tagKey string) string {
	return videoPath + "\x00" + tagKey
}

// Press handles a tag hotkey press for the given video and time position.
// If no region is open for this tag+video, it opens one and returns nil, "".
// If one is already open, it closes it and returns the completed Region plus
// the region ID (so the caller can append it to the manifest).
func (t *Toggle) Press(
	sessionID, videoPath, tagKey, tagLabel, tagColor string,
	nowSec float64,
) (*Region, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()

	k := key(videoPath, tagKey)
	if existing, ok := t.open[k]; ok {
		// Close the open region.
		delete(t.open, k)
		now := time.Now()
		r := &Region{
			ID:        NewRegionID(),
			Kind:      KindRegion,
			VideoPath: videoPath,
			SessionID: sessionID,
			TagKey:    existing.tagKey,
			TagLabel:  existing.tagLabel,
			TagColor:  existing.tagColor,
			StartSec:  existing.startSec,
			EndSec:    nowSec,
			CreatedAt: existing.openedAt,
			UpdatedAt: now,
		}
		return r, true
	}

	// Open a new region.
	t.open[k] = &openRegion{
		id:        NewRegionID(),
		tagKey:    tagKey,
		tagLabel:  tagLabel,
		tagColor:  tagColor,
		videoPath: videoPath,
		startSec:  nowSec,
		openedAt:  time.Now(),
	}
	return nil, false
}

// OpenTags returns the tag keys that currently have open regions for a video.
func (t *Toggle) OpenTags(videoPath string) []string {
	t.mu.Lock()
	defer t.mu.Unlock()
	var tags []string
	for k, o := range t.open {
		if o.videoPath == videoPath {
			_ = k
			tags = append(tags, o.tagKey)
		}
	}
	return tags
}

// CancelAll discards all open regions for the given video without closing them.
// Called when switching videos mid-session.
func (t *Toggle) CancelAll(videoPath string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	for k, o := range t.open {
		if o.videoPath == videoPath {
			delete(t.open, k)
		}
	}
}

// IsOpen reports whether a region is currently open for this tag+video.
func (t *Toggle) IsOpen(videoPath, tagKey string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	_, ok := t.open[key(videoPath, tagKey)]
	return ok
}
