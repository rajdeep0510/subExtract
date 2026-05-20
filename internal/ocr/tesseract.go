package ocr

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// IsTesseractInstalled checks if the tesseract command is available on the system.
func IsTesseractInstalled() bool {
	_, err := exec.LookPath("tesseract")
	return err == nil
}

// ExtractText uses the Tesseract CLI to extract text from an image file.
// language can be empty or something like "eng".
func ExtractText(imagePath string, language string) (string, error) {
	args := []string{imagePath, "stdout"}
	if language != "" {
		args = append(args, "-l", language)
	}

	cmd := exec.Command("tesseract", args...)
	
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("tesseract failed: %v (stderr: %s)", err, stderr.String())
	}

	// Clean up the OCR text output (Tesseract often adds extra newlines)
	text := strings.TrimSpace(out.String())
	return text, nil
}
