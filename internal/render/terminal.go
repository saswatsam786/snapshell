package render

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
)

// ClearTerminal clears the terminal screen
func ClearTerminal() {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "cls")
	} else {
		cmd = exec.Command("clear")
	}
	cmd.Stdout = os.Stdout
	cmd.Run()
}

// MoveCursorToTop moves the cursor to the top of the terminal
func MoveCursorToTop() {
	fmt.Print("\033[H")
}

// HideCursor hides the terminal cursor
func HideCursor() {
	fmt.Print("\033[?25l")
}

// ShowCursor shows the terminal cursor
func ShowCursor() {
	fmt.Print("\033[?25h")
}

// GetTerminalSize gets the current terminal size
func GetTerminalSize() (width, height int) {
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

	// Default values if environment variables are not set
	if width == 0 {
		width = 120
	}
	if height == 0 {
		height = 30
	}

	return width, height
}
