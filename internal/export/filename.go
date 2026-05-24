package export

import (
	"fmt"
	"path/filepath"
	"strings"
)

// Filename builds a descriptive output filename for an exported clip.
// Format: <video_stem>_HH-MM-SS_<tag>.mp4
func Filename(videoPath string, startSec float64, tagKey string) string {
	stem := strings.TrimSuffix(filepath.Base(videoPath), filepath.Ext(videoPath))
	ts := formatHMS(startSec)
	tag := sanitize(tagKey)
	return fmt.Sprintf("%s_%s_%s.mp4", stem, ts, tag)
}

func formatHMS(s float64) string {
	total := int(s)
	h := total / 3600
	m := (total % 3600) / 60
	sec := total % 60
	return fmt.Sprintf("%02d-%02d-%02d", h, m, sec)
}

// sanitize strips characters that are unsafe in filenames.
func sanitize(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-' || r == '_' {
			b.WriteRune(r)
		} else {
			b.WriteRune('_')
		}
	}
	return b.String()
}
