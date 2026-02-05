package errclean

import (
	"strings"
)

// CleanedError represents a normalized error
type CleanedError struct {
	Type    string
	Message string
	Stack   []string
}

// ANSI color codes
const (
	colorReset = "\033[0m"
	colorRed   = "\033[31m"
	colorGray  = "\033[90m"
	colorBold  = "\033[1m"
)

// Format returns a human-readable representation with colors
func (e *CleanedError) Format() string {
	var sb strings.Builder

	if e.Type != "" {
		sb.WriteString(colorRed)
		sb.WriteString(colorBold)
		sb.WriteString(e.Type)
		sb.WriteString(colorReset)
		if e.Message != "" {
			sb.WriteString(": ")
		}
	}

	if e.Message != "" {
		sb.WriteString(colorRed)
		sb.WriteString(e.Message)
		sb.WriteString(colorReset)
	}

	if len(e.Stack) > 0 {
		sb.WriteString("\n")
		for _, frame := range e.Stack {
			sb.WriteString(colorGray)
			sb.WriteString("  ")
			sb.WriteString(frame)
			sb.WriteString(colorReset)
			sb.WriteString("\n")
		}
	}

	return sb.String()
}
