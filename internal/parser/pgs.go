package parser

import (
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mbiamont/go-pgs-parser/displaySet"
	"github.com/mbiamont/go-pgs-parser/pgs"
	"github.com/rajdeepvala/subextract/internal/models"
	"github.com/rajdeepvala/subextract/internal/ocr"
	"github.com/rajdeepvala/subextract/internal/detector"
)

// ParsePGS parses a .sup file, extracts images, and uses Tesseract OCR to read the text.
// The lang parameter is passed to Tesseract (e.g., "eng", "fra").
func ParsePGS(filename string, lang string) (*models.Subtitle, error) {
	if !ocr.IsTesseractInstalled() {
		return nil, fmt.Errorf("tesseract OCR is not installed or not in PATH; it is required for PGS processing")
	}

	var subtitle models.Subtitle
	subtitle.Filename = filename
	subtitle.Format = "pgs"
	subtitle.Language = lang

	pgsParser := pgs.NewPgsParser()
	
	var currentEntry *models.SubtitleEntry
	index := 1
	
	// Prepare a temporary directory for images
	tempDir, err := os.MkdirTemp("", "subextract_pgs_ocr_*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir for OCR: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Smart detection state
	isSmartDetect := lang == "" || lang == "und"
	if isSmartDetect && lang == "" {
		lang = "eng" // Default for detection phase
	}
	detectedCount := 0

	err = pgsParser.ParseDisplaySets(filename, func(data displaySet.DisplaySet, startTime time.Duration) error {
		img, err := data.ToImageData()
		if err != nil {
			return err
		}

		if img == nil {
			// A DisplaySet with no image marks the end of the current subtitle.
			if currentEntry != nil {
				currentEntry.EndTime = startTime
				subtitle.Entries = append(subtitle.Entries, *currentEntry)
				currentEntry = nil
			}
		} else {
			// If we get an image while one is already active, close the previous one
			if currentEntry != nil {
				currentEntry.EndTime = startTime
				subtitle.Entries = append(subtitle.Entries, *currentEntry)
			}

			// Create a temporary PNG file for the image
			tempFile := filepath.Join(tempDir, fmt.Sprintf("img_%d.png", index))
			f, err := os.Create(tempFile)
			if err != nil {
				return fmt.Errorf("failed to create temp image: %w", err)
			}
			
			if err := png.Encode(f, img.Image); err != nil {
				f.Close()
				return fmt.Errorf("failed to encode PNG: %w", err)
			}
			f.Close()

			// Perform OCR on the image
			text, err := ocr.ExtractText(tempFile, lang)
			if err != nil {
				return fmt.Errorf("OCR failed on index %d: %w", index, err)
			}

			// Clean up extra whitespace/garbage from Tesseract
			text = strings.TrimSpace(text)
			
			// Even if text is empty, we record it (the normalizer will drop it if necessary)
			currentEntry = &models.SubtitleEntry{
				Index:     index,
				StartTime: startTime,
				Text:      text,
			}
			index++

			// Perform smart detection if needed
			if isSmartDetect && text != "" {
				detectedCount++
				if detectedCount == 15 { // Wait for 15 entries with text
					// Temporary subtitle for detection
					tempSub := &models.Subtitle{Entries: subtitle.Entries}
					detected := detector.DetectLanguage(tempSub)
					if detected != "und" && detected != lang {
						lang = detected
						subtitle.Language = detected
						// We don't re-run OCR for previous frames to save time, 
						// but future frames will use the better language.
					}
					isSmartDetect = false // Detection finished
				}
			}
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse PGS file: %w", err)
	}

	// Handle the edge case where the file ends without an empty DisplaySet
	if currentEntry != nil {
		currentEntry.EndTime = currentEntry.StartTime + (2 * time.Second) // 2-second fallback
		subtitle.Entries = append(subtitle.Entries, *currentEntry)
	}

	return &subtitle, nil
}
