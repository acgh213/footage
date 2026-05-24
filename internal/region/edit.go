package region

import (
	"fmt"
	"time"
)

const mergeTolerance = 0.5 // seconds

// Nudge adjusts the start or end of a region by delta seconds and rewrites.
// field must be "start" or "end".
func Nudge(sessionID, id, field string, delta float64) error {
	return set(sessionID, id, field, func(current float64) float64 {
		return current + delta
	})
}

// Set assigns an absolute value to the start or end of a region.
// field must be "start" or "end".
func Set(sessionID, id, field string, value float64) error {
	return set(sessionID, id, field, func(_ float64) float64 {
		return value
	})
}

func set(sessionID, id, field string, fn func(float64) float64) error {
	entries, err := ReadAll(sessionID)
	if err != nil {
		return err
	}
	now := time.Now()
	found := false
	for i, e := range entries {
		if e.ID() != id || e.Region == nil {
			continue
		}
		r := e.Region
		switch field {
		case "start":
			v := fn(r.StartSec)
			if v < 0 {
				v = 0
			}
			if r.EndSec > 0 && v >= r.EndSec {
				return fmt.Errorf("start cannot be >= end")
			}
			r.StartSec = v
		case "end":
			v := fn(r.EndSec)
			if v <= r.StartSec {
				return fmt.Errorf("end cannot be <= start")
			}
			r.EndSec = v
		default:
			return fmt.Errorf("field must be start or end, got %q", field)
		}
		r.UpdatedAt = now
		entries[i].Region = r
		found = true
		break
	}
	if !found {
		return fmt.Errorf("region %s not found", id)
	}
	return Rewrite(sessionID, entries)
}

// Delete removes a region or bookmark from the manifest.
func Delete(sessionID, id string) error {
	entries, err := ReadAll(sessionID)
	if err != nil {
		return err
	}
	filtered := entries[:0]
	for _, e := range entries {
		if e.ID() != id {
			filtered = append(filtered, e)
		}
	}
	if len(filtered) == len(entries) {
		return fmt.Errorf("entry %s not found", id)
	}
	return Rewrite(sessionID, filtered)
}

// Merge combines two regions into one. Both must be the same tag and same
// video. The gap between them must be ≤ mergeTolerance seconds. The result
// has the earlier start and later end.
func Merge(sessionID, idA, idB string) error {
	entries, err := ReadAll(sessionID)
	if err != nil {
		return err
	}

	var a, b *Region
	var aIdx, bIdx int
	for i, e := range entries {
		if e.Region == nil {
			continue
		}
		if e.ID() == idA {
			a = e.Region
			aIdx = i
		} else if e.ID() == idB {
			b = e.Region
			bIdx = i
		}
	}
	if a == nil {
		return fmt.Errorf("region %s not found", idA)
	}
	if b == nil {
		return fmt.Errorf("region %s not found", idB)
	}
	if a.TagKey != b.TagKey {
		return fmt.Errorf("cannot merge regions with different tags (%s, %s)", a.TagKey, b.TagKey)
	}
	if a.VideoPath != b.VideoPath {
		return fmt.Errorf("cannot merge regions from different videos")
	}

	// Ensure a is before b.
	if a.StartSec > b.StartSec {
		a, b = b, a
		aIdx, bIdx = bIdx, aIdx
	}

	gap := b.StartSec - a.EndSec
	if gap > mergeTolerance {
		return fmt.Errorf("gap between regions is %.2fs (max %.2fs)", gap, mergeTolerance)
	}

	merged := &Region{
		ID:        NewRegionID(),
		Kind:      KindRegion,
		VideoPath: a.VideoPath,
		SessionID: a.SessionID,
		TagKey:    a.TagKey,
		TagLabel:  a.TagLabel,
		TagColor:  a.TagColor,
		StartSec:  a.StartSec,
		EndSec:    b.EndSec,
		Notes:     joinNotes(a.Notes, b.Notes),
		CreatedAt: a.CreatedAt,
		UpdatedAt: time.Now(),
	}

	// Build new entries list: replace the first with merged, drop the second.
	result := make([]Entry, 0, len(entries)-1)
	for i, e := range entries {
		if i == aIdx {
			result = append(result, Entry{Region: merged})
		} else if i == bIdx {
			continue
		} else {
			result = append(result, e)
		}
	}
	return Rewrite(sessionID, result)
}

func joinNotes(a, b string) string {
	if a == "" {
		return b
	}
	if b == "" {
		return a
	}
	return a + " / " + b
}
