package extractor

import (
	"fmt"
	"os/exec"
)

// ExtractStream extracts a specific subtitle stream from a video file using FFmpeg.
// It converts the stream to SRT format during extraction.
func ExtractStream(inputFile string, streamIndex int, outputFile string) error {
	// Map the stream index to the FFmpeg-compatible format (s:<i>)
	// Usually, streamIndex from ffprobe is absolute.
	// We want to map the absolute stream index.
	cmd := exec.Command("ffmpeg",
		"-v", "error",
		"-i", inputFile,
		"-map", fmt.Sprintf("0:%d", streamIndex),
		"-f", "srt",
		"-y", // overwrite
		outputFile,
	)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to extract stream %d: %w", streamIndex, err)
	}

	return nil
}

// ExtractRawStream extracts a subtitle stream to a raw file without conversion (e.g. for .sup)
func ExtractRawStream(inputFile string, streamIndex int, outputFile string) error {
	cmd := exec.Command("ffmpeg",
		"-v", "error",
		"-i", inputFile,
		"-map", fmt.Sprintf("0:%d", streamIndex),
		"-c:s", "copy",
		"-y", // overwrite
		outputFile,
	)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to extract raw stream %d: %w", streamIndex, err)
	}

	return nil
}
