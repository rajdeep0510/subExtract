package normalizer

import (
	"testing"
	"time"

	"github.com/rajdeepvala/subextract/internal/models"
)

func TestNormalize(t *testing.T) {
	sub := &models.Subtitle{
		Entries: []models.SubtitleEntry{
			{
				Index:     1,
				StartTime: -1 * time.Second, // Negative timing
				EndTime:   2 * time.Second,
				Text:      "  Hello  ", // Trimming
			},
			{
				Index:     5, // Broken index
				StartTime: 1 * time.Second, // Overlap with previous
				EndTime:   3 * time.Second,
				Text:      "World",
			},
			{
				Index:     3,
				StartTime: 5 * time.Second,
				EndTime:   4 * time.Second, // End < Start
				Text:      "Invalid",
			},
			{
				Index:     4,
				StartTime: 6 * time.Second,
				EndTime:   7 * time.Second,
				Text:      "", // Empty text
			},
		},
	}

	Normalize(sub)

	if len(sub.Entries) != 3 {
		t.Errorf("Expected 3 entries after normalization, got %d", len(sub.Entries))
	}

	// Check first entry
	if sub.Entries[0].StartTime != 0 {
		t.Errorf("Expected StartTime 0 for first entry, got %v", sub.Entries[0].StartTime)
	}
	if sub.Entries[0].EndTime != 1*time.Second {
		t.Errorf("Expected EndTime 1s for first entry (trimmed to avoid overlap), got %v", sub.Entries[0].EndTime)
	}
	if sub.Entries[0].Text != "Hello" {
		t.Errorf("Expected text 'Hello', got '%s'", sub.Entries[0].Text)
	}
	if sub.Entries[0].Index != 1 {
		t.Errorf("Expected Index 1, got %d", sub.Entries[0].Index)
	}

	// Check second entry
	if sub.Entries[1].Index != 2 {
		t.Errorf("Expected Index 2, got %d", sub.Entries[1].Index)
	}

	// Check third entry (originally end < start)
	if sub.Entries[2].EndTime <= sub.Entries[2].StartTime {
		t.Errorf("Expected EndTime > StartTime, got %v <= %v", sub.Entries[2].EndTime, sub.Entries[2].StartTime)
	}
}
