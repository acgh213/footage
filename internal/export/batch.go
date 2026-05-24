package export

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/acgh213/footage/internal/region"
)

// BatchJob holds state for a single region in a batch export run.
type BatchJob struct {
	Region     region.Region
	OutputPath string
	Err        error
	Done       bool
}

// BatchResult is sent to the progress callback for each completed job.
type BatchResult struct {
	RegionID string
	OutPath  string
	Err      error
}

// RunBatch exports a slice of regions one at a time. onResult is called after
// each job completes (or fails). Blocks until all jobs finish or ctx is
// cancelled. outDir is the base directory; subdirectories are NOT created per
// video — all clips land in outDir directly.
func RunBatch(
	ctx context.Context,
	regions []region.Region,
	outDir string,
	ffmpegPath string,
	onResult func(BatchResult),
) error {
	if len(regions) == 0 {
		return nil
	}

	var mu sync.Mutex
	var firstErr error

	for _, r := range regions {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		outFile := filepath.Join(outDir, Filename(r.VideoPath, r.StartSec, r.TagKey))
		opt := Options{
			VideoPath:  r.VideoPath,
			StartSec:   r.StartSec,
			EndSec:     r.EndSec,
			OutputPath: outFile,
			FFmpegPath: ffmpegPath,
		}

		err := Remux(ctx, opt)
		res := BatchResult{RegionID: r.ID, OutPath: outFile, Err: err}
		if err != nil {
			mu.Lock()
			if firstErr == nil {
				firstErr = fmt.Errorf("region %s: %w", r.ID, err)
			}
			mu.Unlock()
		}
		if onResult != nil {
			onResult(res)
		}
	}
	return firstErr
}
