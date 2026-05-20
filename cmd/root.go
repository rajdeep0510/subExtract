package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"

	"golang.org/x/sync/errgroup"
	"github.com/spf13/cobra"
	"github.com/charmbracelet/lipgloss"
	"github.com/rajdeepvala/subextract/internal/models"
	"github.com/rajdeepvala/subextract/internal/ffmpeg"
	"github.com/rajdeepvala/subextract/internal/extractor"
	"github.com/rajdeepvala/subextract/internal/parser"
	"github.com/rajdeepvala/subextract/internal/detector"
	"github.com/rajdeepvala/subextract/internal/normalizer"
	"github.com/rajdeepvala/subextract/internal/writer"
	"github.com/rajdeepvala/subextract/internal/tui"
)

var (
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Bold(true)
	infoStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("4"))
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Bold(true)
	warnStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	headerStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("5")).Bold(true).Underline(true)
)

var (
	lang      string
	list      bool
	recursive bool
	streamID  int
	jobs      int
	dryRun    bool
	verbose   bool
	fps       float64
)

func isCommandAvailable(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

var rootCmd = &cobra.Command{
	Use:   "subextract [video_file_or_directory]",
	Short: "Extract and normalize subtitles from video files",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check for FFmpeg
		if !isCommandAvailable("ffmpeg") {
			return fmt.Errorf("ffmpeg is not installed or not in PATH; it is required for extraction")
		}
		if !isCommandAvailable("ffprobe") {
			return fmt.Errorf("ffprobe is not installed or not in PATH; it is required for scanning")
		}

		inputPath := args[0]

		info, err := os.Stat(inputPath)
		if err != nil {
			return fmt.Errorf("failed to access input path: %w", err)
		}

		if info.IsDir() {
			return processDirectory(inputPath)
		}

		return processFile(inputPath, false)
	},
}

func processDirectory(dirPath string) error {
	if verbose {
		fmt.Println(infoStyle.Render(fmt.Sprintf("Scanning directory: %s", dirPath)))
	}

	var files []string
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			if path != dirPath && !recursive {
				return filepath.SkipDir
			}
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		switch ext {
		case ".mkv", ".mp4", ".avi", ".srt", ".ass", ".ssa", ".vtt", ".sub":
			files = append(files, path)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("error walking directory: %w", err)
	}

	if len(files) == 0 {
		fmt.Println(warnStyle.Render("No supported files found."))
		return nil
	}

	fmt.Println(headerStyle.Render(fmt.Sprintf("Found %d files to process. Using %d parallel jobs.", len(files), jobs)))

	g, _ := errgroup.WithContext(context.Background())
	g.SetLimit(jobs)

	processedCount := int32(0)
	for _, f := range files {
		file := f
		g.Go(func() error {
			if verbose {
				fmt.Printf("Processing %s\n", file)
			}
			err := processFile(file, true)
			if err != nil {
				fmt.Fprintln(os.Stderr, errorStyle.Render(fmt.Sprintf("Error processing %s: %v", file, err)))
				return nil // Don't stop the whole batch for one file error
			}
			atomic.AddInt32(&processedCount, 1)
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return err
	}

	summary := fmt.Sprintf("\nFinished batch processing. Successfully processed %d/%d files.", processedCount, len(files))
	if processedCount == int32(len(files)) {
		fmt.Println(successStyle.Render(summary))
	} else {
		fmt.Println(warnStyle.Render(summary))
	}
	return nil
}

func processFile(inputFile string, isBatch bool) error {
	ext := strings.ToLower(filepath.Ext(inputFile))

	var sub *models.Subtitle
	var err error

	// Check if it's already a subtitle file
	isSubtitle := false
	switch ext {
	case ".srt", ".ass", ".ssa", ".vtt", ".sub":
		isSubtitle = true
	}

	if isSubtitle {
		if verbose {
			fmt.Println(infoStyle.Render(fmt.Sprintf("Processing subtitle file: %s", inputFile)))
		}
		if dryRun {
			fmt.Printf("[Dry-run] Would process subtitle file: %s\n", inputFile)
			return nil
		}

		if ext == ".sub" {
			// MicroDVD requires FPS
			currentFPS := fps
			if currentFPS <= 0 {
				// Try to detect from a video file with the same name
				base := strings.TrimSuffix(inputFile, ext)
				videoExtensions := []string{".mkv", ".mp4", ".avi"}
				for _, vExt := range videoExtensions {
					videoFile := base + vExt
					if _, err := os.Stat(videoFile); err == nil {
						if verbose {
							fmt.Printf("Attempting to detect FPS from %s\n", videoFile)
						}
						streams, err := ffmpeg.GetStreams(videoFile)
						if err == nil {
							for _, s := range streams {
								if s.CodecType == "video" {
									vFPS, err := s.GetVideoFPS()
									if err == nil {
										currentFPS = vFPS
										if verbose {
											fmt.Printf("Detected FPS: %.3f\n", currentFPS)
										}
										break
									}
								}
							}
						}
					}
					if currentFPS > 0 {
						break
					}
				}
			}

			if currentFPS <= 0 {
				return fmt.Errorf("FPS missing for MicroDVD subtitle conversion; use --fps flag or ensure video file is next to it")
			}
			sub, err = parser.ParseMicroDVD(inputFile, currentFPS)
		} else {
			// For now we only have SRT parser, but we can use ffmpeg to convert any to SRT first
			tmpFile := filepath.Join(os.TempDir(), fmt.Sprintf("subextract_tmp_%d_%x.srt", os.Getpid(), filepath.Base(inputFile)))
			defer os.Remove(tmpFile)

			// Convert to SRT using FFmpeg (handles ASS/SSA/VTT to SRT conversion easily)
			convCmd := exec.Command("ffmpeg", "-v", "error", "-i", inputFile, "-y", tmpFile)
			if err := convCmd.Run(); err != nil {
				return fmt.Errorf("failed to convert %s to srt: %w", inputFile, err)
			}

			sub, err = parser.ParseSRT(tmpFile)
		}
		if err != nil {
			return err
		}
	} else {
		// Handle as video file
		streams, err := ffmpeg.GetSubtitleStreams(inputFile)
		if err != nil {
			return err
		}

		if len(streams) == 0 {
			if verbose {
				fmt.Printf("No subtitle streams found in %s\n", inputFile)
			}
			return nil // Or return error if explicitly requested for a single file
		}

		if list {
			fmt.Println(headerStyle.Render(fmt.Sprintf("Subtitle streams in %s:", inputFile)))
			for _, s := range streams {
				fmt.Printf("[%d] codec: %s, lang: %s\n", s.Index, s.CodecName, s.GetLanguage())
			}
			return nil
		}

		// Selection logic
		var selectedStream *ffmpeg.Stream
		
		if streamID != -1 {
			for _, s := range streams {
				if s.Index == streamID {
					selectedStream = &s
					break
				}
			}
			if selectedStream == nil {
				return fmt.Errorf("stream index %d not found in %s", streamID, inputFile)
			}
		} else if lang != "" {
			for _, s := range streams {
				if strings.EqualFold(s.GetLanguage(), lang) {
					selectedStream = &s
					break
				}
			}
			if selectedStream == nil {
				if isBatch {
					if verbose {
						fmt.Printf("Skipping %s: no stream with language %s\n", inputFile, lang)
					}
					return nil
				}
				return fmt.Errorf("no subtitle stream found with language: %s", lang)
			}
		} else if len(streams) > 1 {
			// Pre-select language detection for unknown tracks
			if !isBatch {
				_ = tui.ShowLoading("Probing subtitle tracks for language detection", func() {
					g, _ := errgroup.WithContext(context.Background())
					g.SetLimit(jobs)
					for i := range streams {
						idx := i
						g.Go(func() error {
							s := &streams[idx]
							if s.GetLanguage() == "Unknown" {
								detected := probeStreamLanguage(inputFile, s)
								if detected != "Unknown" {
									s.DetectedLanguage = detected
								}
							}
							return nil
						})
					}
					_ = g.Wait()
				})
			}

			if isBatch {
				// In batch mode, we pick the first one.
				selectedStream = &streams[0]
			} else {
				// Multiple streams found, use TUI selector
				var err error
				selectedStream, err = tui.SelectStream(streams)
				if err != nil {
					return err
				}
			}
		} else {
			// Default to the first stream
			selectedStream = &streams[0]
		}

		// Safeguard: Check if the codec is image-based
		isImageBased := false
		switch selectedStream.CodecName {
		case "hdmv_pgs_subtitle":
			isImageBased = true
		case "dvd_subtitle", "dvb_subtitle", "xsub":
			return fmt.Errorf("detected image-based subtitles (%s); only hdmv_pgs_subtitle is supported for OCR currently", selectedStream.CodecName)
		}

		if dryRun {
			fmt.Printf("[Dry-run] Would extract stream %d (%s) from %s\n", selectedStream.Index, selectedStream.GetLanguage(), inputFile)
			return nil
		}

		if isImageBased {
			// Extract to .sup and parse with OCR
			tmpFile := filepath.Join(os.TempDir(), fmt.Sprintf("subextract_tmp_%d_%d.sup", os.Getpid(), selectedStream.Index))
			defer os.Remove(tmpFile)

			loadingTitle := fmt.Sprintf("Extracting and running OCR on stream %d (%s)", selectedStream.Index, selectedStream.GetLanguage())
			_ = tui.ShowLoading(loadingTitle, func() {
				err = extractor.ExtractRawStream(inputFile, selectedStream.Index, tmpFile)
				if err != nil {
					return
				}

				ocrLang := lang
				if ocrLang == "" {
					ocrLang = selectedStream.GetLanguage()
				}
				if ocrLang == "Unknown" || ocrLang == "und" {
					ocrLang = "eng" // Fallback to English
				}

				sub, err = parser.ParsePGS(tmpFile, ocrLang)
			})
			
			if err != nil {
				return fmt.Errorf("extraction or OCR failed: %w", err)
			}
		} else {
			// Create a temporary file for extraction
			tmpFile := filepath.Join(os.TempDir(), fmt.Sprintf("subextract_tmp_%d_%d.srt", os.Getpid(), selectedStream.Index))
			defer os.Remove(tmpFile)

			loadingTitle := fmt.Sprintf("Extracting stream %d (%s)", selectedStream.Index, selectedStream.GetLanguage())
			_ = tui.ShowLoading(loadingTitle, func() {
				err = extractor.ExtractStream(inputFile, selectedStream.Index, tmpFile)
				if err != nil {
					return
				}

				// Parse the extracted SRT
				sub, err = parser.ParseSRT(tmpFile)
			})
			
			if err != nil {
				return fmt.Errorf("failed to extract or parse subtitles: %w", err)
			}
		}
	}

	// Normalize
	if verbose {
		fmt.Println("Normalizing subtitles...")
	}
	normalizer.Normalize(sub)

	// Smart Language Detection
	if sub.Language == "" || sub.Language == "Unknown" {
		detected := detector.DetectLanguage(sub)
		if detected != "Unknown" {
			sub.Language = detected
			if verbose {
				fmt.Printf("Smart-detected language: %s\n", sub.Language)
			}
		}
	}

	// Determine output filename
	var outputFile string
	if isSubtitle {
		outputFile = strings.TrimSuffix(inputFile, ext) + ".normalized.srt"
	} else {
		outputFile = strings.TrimSuffix(inputFile, ext) + ".srt"
	}

	// Write the normalized SRT
	err = writer.WriteSRT(sub, outputFile)
	if err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	if verbose || !isBatch {
		fmt.Println(successStyle.Render(fmt.Sprintf("Successfully saved to %s", outputFile)))
	}
	return nil
}

func Execute() {
	if err := GetRoot().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, errorStyle.Render("\nError: ")+err.Error())
		os.Exit(1)
	}
}

func GetRoot() *cobra.Command {
	return rootCmd
}

func probeStreamLanguage(inputFile string, s *ffmpeg.Stream) string {
	// Create a temporary file for the snippet
	tmpFile := filepath.Join(os.TempDir(), fmt.Sprintf("probe_%d_%d.srt", os.Getpid(), s.Index))
	defer os.Remove(tmpFile)

	// Extract just 60 seconds of the subtitle track
	// For image-based tracks, we skip probing for now as OCR is slow for TUI
	if s.CodecName == "hdmv_pgs_subtitle" || s.CodecName == "dvd_subtitle" {
		return "Unknown"
	}

	cmd := exec.Command("ffmpeg",
		"-v", "error",
		"-i", inputFile,
		"-map", fmt.Sprintf("0:%d", s.Index),
		"-t", "60", // Only first 60 seconds
		"-f", "srt",
		"-y",
		tmpFile,
	)

	if err := cmd.Run(); err != nil {
		return "Unknown"
	}

	sub, err := parser.ParseSRT(tmpFile)
	if err != nil {
		return "Unknown"
	}

	return detector.DetectLanguage(sub)
}

func init() {
	rootCmd.Flags().StringVarP(&lang, "lang", "l", "", "Language of the subtitle stream to extract")
	rootCmd.Flags().BoolVarP(&list, "list", "s", false, "List subtitle streams available in the file")
	rootCmd.Flags().BoolVarP(&recursive, "recursive", "r", false, "Recursively process directories")
	rootCmd.Flags().IntVarP(&streamID, "stream", "i", -1, "Stream index to extract (overrides --lang)")
	rootCmd.Flags().IntVarP(&jobs, "jobs", "j", 1, "Number of parallel jobs for batch processing")
	rootCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Don't actually extract or write files")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose logging")
	rootCmd.Flags().Float64Var(&fps, "fps", 0, "FPS for MicroDVD (.sub) conversion (auto-detected if possible)")
}
