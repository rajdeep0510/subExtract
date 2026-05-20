package normalizer

import (
	"sort"
	"strings"
	"time"

	"github.com/rajdeepvala/subextract/internal/models"
)

const MinDuration = 100 * time.Millisecond

// Normalize applies the normalization pipeline to the subtitle model.
func Normalize(sub *models.Subtitle) {
	if len(sub.Entries) == 0 {
		return
	}

	// 1. Sort by start time
	sort.Slice(sub.Entries, func(i, j int) bool {
		if sub.Entries[i].StartTime == sub.Entries[j].StartTime {
			return sub.Entries[i].EndTime < sub.Entries[j].EndTime
		}
		return sub.Entries[i].StartTime < sub.Entries[j].StartTime
	})

	var normalized []models.SubtitleEntry

	for i := 0; i < len(sub.Entries); i++ {
		entry := sub.Entries[i]

		// 2. Remove empty text
		entry.Text = strings.TrimSpace(entry.Text)
		if entry.Text == "" {
			continue
		}

		// 3. Fix negative timings
		if entry.StartTime < 0 {
			entry.StartTime = 0
		}
		if entry.EndTime < 0 {
			entry.EndTime = 0
		}

		// 4. Ensure end > start
		if entry.EndTime <= entry.StartTime {
			entry.EndTime = entry.StartTime + MinDuration
		}

		// 5. Normalize text (UTF-8 assumed as Go strings are UTF-8)
		// and normalize line endings
		entry.Text = strings.ReplaceAll(entry.Text, "\r\n", "\n")
		entry.Text = strings.ReplaceAll(entry.Text, "\r", "\n")

		normalized = append(normalized, entry)
	}

	// 6. Fix overlaps
	for i := 0; i < len(normalized)-1; i++ {
		if normalized[i].EndTime > normalized[i+1].StartTime {
			// Trim the current one to not overlap with the next one
			normalized[i].EndTime = normalized[i+1].StartTime
			
			// If trimming made it too short, ensure minimum duration
			if normalized[i].EndTime <= normalized[i].StartTime {
				normalized[i].EndTime = normalized[i].StartTime + MinDuration
				// If it still overlaps, we might have a problem, but this is a simple fix for now
			}
		}
	}

	// 7. Rebuild indexes
	for i := range normalized {
		normalized[i].Index = i + 1
	}

	sub.Entries = normalized
}
