package parser

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/rajdeepvala/subextract/internal/models"
)

// MicroDVD format: {start_frame}{end_frame}Text
var microDVDRegex = regexp.MustCompile(`\{(\d+)\}\{(\d+)\}(.*)`)

// ParseMicroDVD parses a MicroDVD .sub file into the internal Subtitle model.
// fps is required for converting frame numbers to time durations.
func ParseMicroDVD(filename string, fps float64) (*models.Subtitle, error) {
	if fps <= 0 {
		return nil, fmt.Errorf("invalid FPS: %f (FPS is required for MicroDVD conversion)", fps)
	}

	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var subtitle models.Subtitle
	subtitle.Filename = filename
	subtitle.Format = "sub"

	scanner := bufio.NewScanner(file)
	index := 1
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		matches := microDVDRegex.FindStringSubmatch(line)
		if len(matches) != 4 {
			// Some files might have metadata lines or slightly different formats
			continue
		}

		startFrame, _ := strconv.ParseInt(matches[1], 10, 64)
		endFrame, _ := strconv.ParseInt(matches[2], 10, 64)
		text := strings.ReplaceAll(matches[3], "|", "\n")

		startTime := time.Duration(float64(startFrame)/fps*1000) * time.Millisecond
		endTime := time.Duration(float64(endFrame)/fps*1000) * time.Millisecond

		subtitle.Entries = append(subtitle.Entries, models.SubtitleEntry{
			Index:     index,
			StartTime: startTime,
			EndTime:   endTime,
			Text:      text,
		})
		index++
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return &subtitle, nil
}
