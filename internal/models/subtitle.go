package models

import "time"

// SubtitleEntry represents a single subtitle block with timing and text.
type SubtitleEntry struct {
	Index       int           `json:"index"`
	StartTime   time.Duration `json:"start_time"`
	EndTime     time.Duration `json:"end_time"`
	Text        string        `json:"text"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// Subtitle represents a collection of subtitle entries and associated metadata.
type Subtitle struct {
	Entries  []SubtitleEntry
	Format   string
	Language string
	Filename string
}
