package writer

import (
	"fmt"
	"os"
	"time"

	"github.com/rajdeepvala/subextract/internal/models"
)

// WriteSRT writes the Subtitle model to an SRT file.
func WriteSRT(sub *models.Subtitle, outputFile string) error {
	file, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer file.Close()

	for i, entry := range sub.Entries {
		if i > 0 {
			fmt.Fprintln(file)
		}
		fmt.Fprintf(file, "%d\n", entry.Index)
		fmt.Fprintf(file, "%s --> %s\n", formatSRTTime(entry.StartTime), formatSRTTime(entry.EndTime))
		fmt.Fprintf(file, "%s\n", entry.Text)
	}

	return nil
}

func formatSRTTime(d time.Duration) string {
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second
	d -= s * time.Second
	ms := d / time.Millisecond

	return fmt.Sprintf("%02d:%02d:%02d,%03d", h, m, s, ms)
}
