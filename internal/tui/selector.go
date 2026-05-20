package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/rajdeepvala/subextract/internal/ffmpeg"
)

// SelectStream provides an interactive TUI to select a subtitle stream.
func SelectStream(streams []ffmpeg.Stream) (*ffmpeg.Stream, error) {
	if len(streams) == 0 {
		return nil, fmt.Errorf("no streams provided to selector")
	}

	if len(streams) == 1 {
		return &streams[0], nil
	}

	var options []huh.Option[int]
	for _, s := range streams {
		lang := s.GetLanguage()
		
		// Map common ISO codes to full names if ffprobe returns short codes
		// Most files return "eng" or "en", we want "English"
		fullLang := getFullLanguageName(lang)
		
		status := ""
		if s.DetectedLanguage != "" {
			status = " (detected)"
		}

		label := fmt.Sprintf("[%d] %s - %s%s", s.Index, s.CodecName, fullLang, status)
		options = append(options, huh.NewOption(label, s.Index))
	}

	var selectedIndex int
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[int]().
				Title("Select Subtitle Stream").
				Description("Multiple subtitle tracks found. Choose one to extract:").
				Options(options...).
				Value(&selectedIndex),
		),
	)

	err := form.Run()
	if err != nil {
		return nil, err
	}

	for _, s := range streams {
		if s.Index == selectedIndex {
			return &s, nil
		}
	}

	return nil, fmt.Errorf("stream with index %d not found", selectedIndex)
}

type loadingModel struct {
	spinner  spinner.Model
	title    string
	action   func()
	done     chan bool
	quitting bool
}

func (m loadingModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, func() tea.Msg {
		m.action()
		m.done <- true
		return tea.Quit()
	})
}

func (m loadingModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			m.quitting = true
			return m, tea.Quit
		}
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m loadingModel) View() string {
	if m.quitting {
		return ""
	}
	return fmt.Sprintf("\n %s %s...\n", m.spinner.View(), m.title)
}

// ShowLoading runs an action while showing a spinner TUI.
func ShowLoading(title string, action func()) error {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	done := make(chan bool, 1)
	m := loadingModel{
		spinner: s,
		title:   title,
		action:  action,
		done:    done,
	}

	p := tea.NewProgram(m)
	_, err := p.Run()
	return err
}

// getFullLanguageName converts ISO codes to human-readable names
func getFullLanguageName(lang string) string {
	l := strings.ToLower(lang)
	
	// If it's already a full name (from whatlanggo), return it
	if len(l) > 3 {
		return strings.Title(l)
	}

	// Simple map for common ffprobe outputs
	isoMap := map[string]string{
		"eng": "English",
		"en":  "English",
		"fra": "French",
		"fr":  "French",
		"deu": "German",
		"ger": "German",
		"de":  "German",
		"spa": "Spanish",
		"es":  "Spanish",
		"ita": "Italian",
		"it":  "Italian",
		"jpn": "Japanese",
		"ja":  "Japanese",
		"chi": "Chinese",
		"zho": "Chinese",
		"zh":  "Chinese",
		"rus": "Russian",
		"ru":  "Russian",
		"hun": "Hungarian",
		"hu":  "Hungarian",
		"und": "Unknown",
	}

	if full, ok := isoMap[l]; ok {
		return full
	}

	return strings.Title(l)
}
