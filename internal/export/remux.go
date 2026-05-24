package export

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// Options controls a single remux job.
type Options struct {
	VideoPath  string
	StartSec   float64
	EndSec     float64
	OutputPath string
	FFmpegPath string // optional; falls back to PATH
}

// Remux runs ffmpeg stream copy for the given region. It blocks until ffmpeg
// exits or ctx is cancelled. Returns an error if ffmpeg exits non-zero.
func Remux(ctx context.Context, opt Options) error {
	ffmpeg := opt.FFmpegPath
	if ffmpeg == "" {
		var err error
		ffmpeg, err = exec.LookPath("ffmpeg")
		if err != nil {
			return fmt.Errorf("ffmpeg not found in PATH — install ffmpeg or set ffmpeg_path in settings")
		}
	}

	if err := os.MkdirAll(filepath.Dir(opt.OutputPath), 0755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}

	args := []string{
		"-y",
		"-ss", fmt.Sprintf("%.3f", opt.StartSec),
		"-to", fmt.Sprintf("%.3f", opt.EndSec),
		"-i", opt.VideoPath,
		"-c", "copy",
		"-avoid_negative_ts", "make_zero",
		opt.OutputPath,
	}

	cmd := exec.CommandContext(ctx, ffmpeg, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg: %w\n%s", err, string(out))
	}
	return nil
}
