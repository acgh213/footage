package region

import "strings"

// Query filters regions by tag, video path, text content, and duration.
type Query struct {
	Tags        []string
	Video       string
	ContainsText string
	MinDuration  float64
	MaxDuration  float64
}

// Filter returns only the regions (not bookmarks) matching q.
// Zero values in q are treated as "no constraint".
func Filter(entries []Entry, q Query) []Entry {
	tagSet := make(map[string]bool, len(q.Tags))
	for _, t := range q.Tags {
		tagSet[strings.ToLower(t)] = true
	}

	var out []Entry
	for _, e := range entries {
		if e.Region == nil {
			continue
		}
		r := e.Region
		if len(q.Tags) > 0 && !tagSet[strings.ToLower(r.TagKey)] && !tagSet[strings.ToLower(r.TagLabel)] {
			continue
		}
		if q.Video != "" && !strings.EqualFold(r.VideoPath, q.Video) {
			continue
		}
		if q.ContainsText != "" {
			hay := strings.ToLower(r.Notes + " " + r.TagLabel)
			if !strings.Contains(hay, strings.ToLower(q.ContainsText)) {
				continue
			}
		}
		dur := r.EndSec - r.StartSec
		if q.MinDuration > 0 && dur < q.MinDuration {
			continue
		}
		if q.MaxDuration > 0 && dur > q.MaxDuration {
			continue
		}
		out = append(out, e)
	}
	return out
}
