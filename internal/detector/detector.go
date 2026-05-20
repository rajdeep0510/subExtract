package detector

import (
	"strings"

	"github.com/abadojack/whatlanggo"
	"github.com/rajdeepvala/subextract/internal/models"
)

// DetectLanguage attempts to detect the language of the subtitle content.
// It samples entries from different parts of the file to get a better representation.
func DetectLanguage(sub *models.Subtitle) string {
	if sub == nil || len(sub.Entries) == 0 {
		return "Unknown"
	}

	// Concatenate text from a wider sample (up to 40 entries, spread out)
	var sb strings.Builder
	step := len(sub.Entries) / 40
	if step == 0 {
		step = 1
	}

	count := 0
	for i := 0; i < len(sub.Entries) && count < 40; i += step {
		text := strings.TrimSpace(sub.Entries[i].Text)
		if text != "" {
			sb.WriteString(text)
			sb.WriteString(" ")
			count++
		}
	}

	if sb.Len() == 0 {
		return "Unknown"
	}

	info := whatlanggo.Detect(sb.String())
	
	// If confidence is too low, we might want to stay as Unknown
	if info.Confidence < 0.2 {
		return "Unknown"
	}

	// whatlanggo's Lang type has a String() method that returns the full name (e.g. "English")
	return info.Lang.String()
}

// GetISO6393 returns the 3-letter code for a given text sample.
func GetISO6393(text string) string {
	info := whatlanggo.Detect(text)
	if info.Lang == whatlanggo.Eng && info.Confidence < 0.1 {
		return "und"
	}
	return info.Lang.Iso6393()
}
