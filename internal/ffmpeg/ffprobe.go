package ffmpeg

import (
	"encoding/json"
	"fmt"
	"os/exec"
)

type FFProbeOutput struct {
	Streams []Stream `json:"streams"`
}

type Stream struct {
	Index            int               `json:"index"`
	CodecName        string            `json:"codec_name"`
	CodecType        string            `json:"codec_type"`
	AvgFrameRate     string            `json:"avg_frame_rate"`
	Language         string            `json:"tags,omitempty"`
	Tags             map[string]string `json:"tags"`
	Disposition      map[string]int    `json:"disposition"`
	DetectedLanguage string            `json:"-"`
}

func (s *Stream) GetLanguage() string {
	if s.DetectedLanguage != "" {
		return s.DetectedLanguage
	}
	if lang, ok := s.Tags["language"]; ok {
		return lang
	}
	return "Unknown"
}

// GetVideoFPS parses the average frame rate from a video stream.
func (s *Stream) GetVideoFPS() (float64, error) {
	if s.AvgFrameRate == "" || s.AvgFrameRate == "0/0" {
		return 0, fmt.Errorf("no frame rate information")
	}

	var num, den float64
	_, err := fmt.Sscanf(s.AvgFrameRate, "%f/%f", &num, &den)
	if err != nil {
		return 0, fmt.Errorf("invalid frame rate format: %s", s.AvgFrameRate)
	}

	if den == 0 {
		return 0, fmt.Errorf("invalid frame rate: denominator is zero")
	}

	return num / den, nil
}

// GetStreams uses ffprobe to list all streams in a media file.
func GetStreams(inputFile string) ([]Stream, error) {
	cmd := exec.Command("ffprobe", 
		"-v", "error", 
		"-show_entries", "stream=index,codec_name,codec_type,disposition,avg_frame_rate:stream_tags=language", 
		"-of", "json", 
		inputFile,
	)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run ffprobe: %w", err)
	}

	var ffOut FFProbeOutput
	if err := json.Unmarshal(output, &ffOut); err != nil {
		return nil, fmt.Errorf("failed to parse ffprobe output: %w", err)
	}

	return ffOut.Streams, nil
}

// GetSubtitleStreams filters the streams for subtitle types.
func GetSubtitleStreams(inputFile string) ([]Stream, error) {
	streams, err := GetStreams(inputFile)
	if err != nil {
		return nil, err
	}

	var subStreams []Stream
	for _, s := range streams {
		if s.CodecType == "subtitle" {
			subStreams = append(subStreams, s)
		}
	}

	return subStreams, nil
}
