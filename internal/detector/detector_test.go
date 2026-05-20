package detector

import (
	"testing"

	"github.com/rajdeepvala/subextract/internal/models"
)

func TestDetectLanguage(t *testing.T) {
	tests := []struct {
		name     string
		entries  []models.SubtitleEntry
		expected string
	}{
		{
			name: "English",
			entries: []models.SubtitleEntry{
				{Text: "Hello, how are you today?"},
				{Text: "I am doing great, thank you."},
				{Text: "This is a subtitle in English language."},
			},
			expected: "eng",
		},
		{
			name: "French",
			entries: []models.SubtitleEntry{
				{Text: "Bonjour, comment allez-vous?"},
				{Text: "Je vais bien, merci."},
				{Text: "Ceci est un sous-titre en français."},
			},
			expected: "fra",
		},
		{
			name: "Spanish",
			entries: []models.SubtitleEntry{
				{Text: "Hola, ¿cómo estás?"},
				{Text: "Estoy muy bien, gracias."},
				{Text: "Este es un subtítulo en español."},
			},
			expected: "spa",
		},
		{
			name:     "Empty",
			entries:  []models.SubtitleEntry{},
			expected: "und",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sub := &models.Subtitle{Entries: tt.entries}
			got := DetectLanguage(sub)
			if got != tt.expected {
				t.Errorf("DetectLanguage() = %v, want %v", got, tt.expected)
			}
		})
	}
}
