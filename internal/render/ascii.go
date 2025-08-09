package render

import (
	"image"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"gocv.io/x/gocv"
)

// getTerminalSize returns the current terminal dimensions
func getTerminalSize() (width, height int) {
	// Try to get from environment variables first
	if cols := os.Getenv("COLUMNS"); cols != "" {
		if w, err := strconv.Atoi(cols); err == nil {
			width = w
		}
	}
	if lines := os.Getenv("LINES"); lines != "" {
		if h, err := strconv.Atoi(lines); err == nil {
			height = h
		}
	}

	// If environment variables are not set, try to get from stty
	if width == 0 || height == 0 {
		cmd := exec.Command("stty", "size")
		cmd.Stdin = os.Stdin
		output, err := cmd.Output()
		if err == nil {
			parts := strings.Fields(string(output))
			if len(parts) >= 2 {
				if h, err := strconv.Atoi(parts[0]); err == nil {
					height = h
				}
				if w, err := strconv.Atoi(parts[1]); err == nil {
					width = w
				}
			}
		}
	}

	// Fallback default values if still not set
	if width == 0 {
		width = 80
	}
	if height == 0 {
		height = 24
	}

	// Reserve some space for status messages and ensure minimum size
	height -= 3
	if height < 10 {
		height = 10
	}
	if width < 40 {
		width = 40
	}

	return width, height
}

// convertToASCII converts a grayscale pixel value to an ASCII character
func convertToASCII(pixelValue byte) string {
	// Basic ASCII characters from dark to light
	asciiChars := []string{" ", ".", ":", "-", "=", "+", "*", "#", "%", "@"}

	// Map pixel value (0-255) to ASCII character index (0-9)
	charIndex := int(pixelValue) * (len(asciiChars) - 1) / 255
	if charIndex >= len(asciiChars) {
		charIndex = len(asciiChars) - 1
	}

	return asciiChars[charIndex]
}

// ConvertFrameToASCII converts a frame to ASCII art with proper scaling
func ConvertFrameToASCII(frame gocv.Mat) string {
	// Convert to grayscale for better ASCII representation
	gray := gocv.NewMat()
	defer gray.Close()
	gocv.CvtColor(frame, &gray, gocv.ColorBGRToGray)

	// Get terminal dimensions
	termWidth, termHeight := getTerminalSize()

	// Calculate scaling factors to fit the image in terminal
	// Terminal characters are typically 2:1 aspect ratio (height:width)
	charAspectRatio := 2.0

	// Adjust target size based on terminal size for better quality
	targetWidth := termWidth
	targetHeight := termHeight

	// For larger terminals, we can afford higher resolution
	if termWidth > 100 && termHeight > 30 {
		targetWidth = termWidth - 2 // Leave some margin
		targetHeight = termHeight - 2
	} else if termWidth > 60 && termHeight > 20 {
		targetWidth = termWidth - 1
		targetHeight = termHeight - 1
	}

	scaleX := float64(gray.Cols()) / float64(targetWidth)
	scaleY := float64(gray.Rows()) / (float64(targetHeight) * charAspectRatio)

	// Use the larger scale to maintain aspect ratio
	scale := scaleX
	if scaleY > scaleX {
		scale = scaleY
	}

	// Ensure minimum scale for quality
	if scale < 1.0 {
		scale = 1.0
	}

	// Calculate new dimensions
	newWidth := int(float64(gray.Cols()) / scale)
	newHeight := int(float64(gray.Rows()) / scale)

	// Resize the image to fit terminal
	resized := gocv.NewMat()
	defer resized.Close()
	gocv.Resize(gray, &resized, image.Point{X: newWidth, Y: newHeight}, 0, 0, gocv.InterpolationLinear)

	var result strings.Builder

	// Convert frame to ASCII
	for y := 0; y < resized.Rows(); y++ {
		for x := 0; x < resized.Cols(); x++ {
			// Get pixel value
			pixelValue := resized.GetUCharAt(y, x)

			// Convert to ASCII
			asciiChar := convertToASCII(pixelValue)
			result.WriteString(asciiChar)
		}
		result.WriteString("\n")
	}

	return result.String()
}
