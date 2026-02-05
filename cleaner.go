package main

import (
	"regexp"
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

// Cleaner processes error messages
type Cleaner struct {
	format string
}

// NewCleaner creates a new error cleaner
func NewCleaner(format string) *Cleaner {
	return &Cleaner{format: format}
}

// Clean processes the error text
func (c *Cleaner) Clean(text string) *CleanedError {
	text = strings.TrimSpace(text)

	// Detect format if auto
	format := c.format
	if format == "auto" {
		format = detectFormat(text)
	}

	// Process based on format
	switch format {
	case "javascript":
		return cleanJavaScript(text)
	case "python":
		return cleanPython(text)
	case "java":
		return cleanJava(text)
	case "go":
		return cleanGo(text)
	case "rust":
		return cleanRust(text)
	default:
		return cleanGeneric(text)
	}
}

// detectFormat attempts to identify the error format
func detectFormat(text string) string {
	lines := strings.Split(text, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Rust compile errors: "error[E0382]:"
		if strings.HasPrefix(line, "error[E") && strings.Contains(line, "]:") {
			return "rust"
		}

		// Rust panic: "thread 'main' panicked at"
		if strings.Contains(line, "panicked at") {
			return "rust"
		}

		// Python: "Traceback (most recent call last):"
		if strings.Contains(line, "Traceback (most recent call last)") {
			return "python"
		}

		// Python syntax/import errors
		if regexp.MustCompile(`^(SyntaxError|IndentationError|TabError|ModuleNotFoundError|ImportError):`).MatchString(line) {
			return "python"
		}

		// TypeScript compile errors: "src/file.ts:42:5 - error TS"
		if strings.Contains(line, " - error TS") {
			return "javascript"
		}

		// JavaScript/Node.js: "Error:" or "at " (stack trace)
		if strings.HasPrefix(line, "at ") ||
			regexp.MustCompile(`(TypeError|ReferenceError|SyntaxError|Error|UnhandledPromiseRejectionWarning):`).MatchString(line) {
			return "javascript"
		}

		// Java: "Exception in thread" or stack with "\tat "
		if strings.Contains(line, "Exception in thread") ||
			strings.HasPrefix(line, "\tat ") {
			return "java"
		}

		// Go build errors: "./main.go:15:2:"
		if regexp.MustCompile(`^\./?\w+\.go:\d+:\d+:`).MatchString(line) {
			return "go"
		}

		// Go test failures: "--- FAIL:"
		if strings.HasPrefix(line, "--- FAIL:") {
			return "go"
		}

		// Go panic: "panic:" or "goroutine" or "fatal error:"
		if strings.HasPrefix(line, "panic:") ||
			strings.HasPrefix(line, "goroutine ") ||
			strings.HasPrefix(line, "fatal error:") {
			return "go"
		}
	}

	return "generic"
}

// Noise removal patterns
var (
	// Timestamps: 2024-01-28T14:10:36, [14:10:36], etc.
	timestampPattern = regexp.MustCompile(`\d{4}-\d{2}-\d{2}[T ]\d{2}:\d{2}:\d{2}(\.\d+)?(Z|[+-]\d{2}:\d{2})?|\[\d{2}:\d{2}:\d{2}\]`)

	// Memory addresses: 0x7f8a9b0c1d2e (12+ hex digits for real addresses)
	memoryPattern = regexp.MustCompile(`0x[0-9a-fA-F]{12,}`)

	// UUIDs: 550e8400-e29b-41d4-a716-446655440000
	uuidPattern = regexp.MustCompile(`[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`)

	// Hex values: 0xdeadbeef
	hexPattern = regexp.MustCompile(`0x[0-9a-fA-F]+`)

	// Absolute paths (simplified - keep relative paths)
	absPathPattern = regexp.MustCompile(`(?:^|[\s(])((?:[A-Z]:\\|/)[^\s:)]+)`)
)

// stripNoise removes common noise from error messages
func stripNoise(text string) string {
	text = timestampPattern.ReplaceAllString(text, "[TIME]")
	text = uuidPattern.ReplaceAllString(text, "[UUID]")
	text = memoryPattern.ReplaceAllString(text, "[ADDR]")
	text = hexPattern.ReplaceAllString(text, "[HEX]")

	// Simplify paths - keep filename only
	text = absPathPattern.ReplaceAllStringFunc(text, func(match string) string {
		parts := strings.Split(match, "/")
		if len(parts) > 0 {
			filename := parts[len(parts)-1]
			// Keep the leading space/paren if present
			prefix := ""
			if len(match) > 0 && (match[0] == ' ' || match[0] == '(') {
				prefix = string(match[0])
			}
			return prefix + filename
		}
		return match
	})

	return text
}

// deduplicateFrames removes consecutive duplicate stack frames
func deduplicateFrames(frames []string) []string {
	if len(frames) == 0 {
		return frames
	}

	result := []string{frames[0]}
	for i := 1; i < len(frames); i++ {
		if frames[i] != frames[i-1] {
			result = append(result, frames[i])
		}
	}
	return result
}
