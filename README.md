# subExtract

`subExtract` is a high-performance CLI tool written in Go, designed to extract, normalize, and convert subtitles from video files. It leverages the power of FFmpeg for media handling and Tesseract OCR for processing image-based (PGS) subtitles, providing a seamless experience for managing your media library's subtitles.

## Features

- **Multi-Format Extraction**: Extract subtitles from `.mkv`, `.mp4`, `.avi`, and more.
- **OCR Support**: Built-in support for image-based subtitles (HDMV PGS) using Tesseract OCR.
- **Smart Language Detection**: Automatically detects subtitle language using `whatlanggo` when metadata is missing.
- **Normalization**: Automatically cleans up and normalizes subtitle formatting and timing.
- **Batch Processing**: Process entire directories with parallel job support for maximum efficiency.
- **Interactive TUI**: Beautiful terminal user interface for selecting tracks when multiple options are available.
- **Auto-conversion**: Converts specialized formats like MicroDVD (.sub), ASS/SSA, and VTT to standard SRT.
- **MicroDVD FPS Detection**: Automatically attempts to detect the correct FPS for MicroDVD conversion from accompanying video files.

## Prerequisites

`subExtract` relies on a few powerful system tools that must be installed and available in your `PATH`:

1.  **FFmpeg & FFprobe**: Required for media stream analysis and extraction.
2.  **Tesseract OCR**: Required for converting image-based subtitles (PGS) into text.
    *   *Note: Ensure you have the necessary language data files (e.g., `tessdata/eng.traineddata`) for the languages you wish to OCR.*

## Installation

### From Source

```bash
go install github.com/rajdeepvala/subextract@latest
```

*Alternatively, clone the repository and build manually:*

```bash
git clone https://github.com/rajdeepvala/subextract.git
cd subextract
go build -o subextract main.go
```

## Usage

### Basic Usage

Extract subtitles from a single video file:
```bash
subextract movie.mkv
```

### Advanced Commands

**List available streams without extracting:**
```bash
subextract movie.mkv --list
```

**Extract a specific language:**
```bash
subextract movie.mkv --lang eng
```

**Extract a specific stream by index:**
```bash
subextract movie.mkv --stream 2
```

**Batch process a directory (Recursive):**
```bash
subextract ./movies --recursive --jobs 4
```

**Convert MicroDVD (.sub) with specific FPS:**
```bash
subextract subtitles.sub --fps 23.976
```

### Available Flags

| Flag | Shorthand | Description |
| :--- | :--- | :--- |
| `--lang` | `-l` | Language of the subtitle stream to extract |
| `--list` | `-s` | List subtitle streams available in the file |
| `--recursive`| `-r` | Recursively process directories |
| `--stream` | `-i` | Stream index to extract (overrides --lang) |
| `--jobs` | `-j` | Number of parallel jobs for batch processing (default 1) |
| `--fps` | | FPS for MicroDVD (.sub) conversion (auto-detected if possible) |
| `--dry-run` | | Show what would be done without making changes |
| `--verbose` | `-v` | Enable verbose logging |
| `--help` | `-h` | Show help for subextract |

## Project Structure

- `cmd/`: CLI entry point and command definitions (Cobra).
- `internal/`: Private library code.
    - `extractor/`: FFmpeg-based stream extraction logic.
    - `ocr/`: Tesseract OCR integration.
    - `parser/`: Parsers for various subtitle formats (SRT, PGS, MicroDVD).
    - `detector/`: Language detection logic.
    - `normalizer/`: Subtitle cleaning and formatting.
    - `tui/`: Interactive terminal components (LipGloss/Huh).
- `docs/`: Documentation and man pages.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3. Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the Branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## License

Distributed under the MIT License. See `LICENSE` for more information.
