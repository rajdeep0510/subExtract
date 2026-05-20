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

var srtTimeRegex = regexp.MustCompile(`(\d{2}):(\d{2}):(\d{2}),(\d{3})`)

// ParseSRT parses an SRT file into the internal Subtitle model.
func ParseSRT(filename string) (*models.Subtitle, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var subtitle models.Subtitle
	subtitle.Filename = filename
	subtitle.Format = "srt"

	scanner := bufio.NewScanner(file)
	var currentEntry *models.SubtitleEntry
	
	state := 0 // 0: index, 1: timing, 2: text
	
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			if currentEntry != nil {
				subtitle.Entries = append(subtitle.Entries, *currentEntry)
				currentEntry = nil
			}
			state = 0
			continue
		}

		switch state {
		case 0:
			index, err := strconv.Atoi(line)
			if err == nil {
				currentEntry = &models.SubtitleEntry{Index: index}
				state = 1
			}
		case 1:
			times := strings.Split(line, " --> ")
			if len(times) == 2 {
				start, err1 := parseSRTTime(times[0])
				end, err2 := parseSRTTime(times[1])
				if err1 == nil && err2 == nil {
					currentEntry.StartTime = start
					currentEntry.EndTime = end
					state = 2
				}
			}
		case 2:
			if currentEntry.Text != "" {
				currentEntry.Text += "\n"
			}
			currentEntry.Text += line
		}
	}

	if currentEntry != nil {
		subtitle.Entries = append(subtitle.Entries, *currentEntry)
	}

	return &subtitle, nil
}

func parseSRTTime(s string) (time.Duration, error) {
	matches := srtTimeRegex.FindStringSubmatch(s)
	if len(matches) != 5 {
		return 0, fmt.Errorf("invalid SRT time format: %s", s)
	}

	hours, _ := strconv.Atoi(matches[1])
	mins, _ := strconv.Atoi(matches[2])
	secs, _ := strconv.Atoi(matches[3])
	millis, _ := strconv.Atoi(matches[4])

	return time.Duration(hours)*time.Hour +
		time.Duration(mins)*time.Minute +
		time.Duration(secs)*time.Second +
		time.Duration(millis)*time.Millisecond, nil
}
