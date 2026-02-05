package golang

import (
	"regexp"
	"strings"

	"github.com/XD637/err/errclean"
	"github.com/XD637/err/registry"
)

// Parser handles Go errors, panics, build errors, and test failures
type Parser struct{}

func init() {
	registry.Register(&Parser{})
}

func (p *Parser) Name() string {
	return "go"
}

func (p *Parser) Detect(text string) int {
	lines := strings.Split(text, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Go build errors: definitive
		if regexp.MustCompile(`^\./?[\w/]+\.go:\d+:\d+:`).MatchString(line) {
			return 100
		}

		// Go test failures: definitive
		if strings.HasPrefix(line, "--- FAIL:") {
			return 100
		}

		// Panic: high confidence
		if strings.HasPrefix(line, "panic:") {
			return 95
		}

		// Fatal error: high confidence
		if strings.HasPrefix(line, "fatal error:") {
			return 95
		}

		// Goroutine: medium-high confidence
		if strings.HasPrefix(line, "goroutine ") {
			return 85
		}

		// Go file in stack trace
		if strings.Contains(line, ".go:") {
			return 70
		}
	}

	return 0
}

func (p *Parser) Parse(text string) *errclean.CleanedError {
	lines := strings.Split(text, "\n")
	result := &errclean.CleanedError{}

	var stackFrames []string

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Build errors: "./main.go:15:2: undefined: fmt.Printl"
		if regexp.MustCompile(`^\./?[\w/]+\.go:\d+:\d+:`).MatchString(trimmed) {
			// Only set the message if this is the first error
			if result.Type != "build error" {
				result.Type = "build error"
				// Extract the error message
				parts := strings.SplitN(trimmed, ": ", 2)
				if len(parts) >= 2 {
					result.Message = parts[1]
				}
			}
			// Keep all file locations
			parts := strings.SplitN(trimmed, ": ", 2)
			if len(parts) >= 1 {
				stackFrames = append(stackFrames, errclean.StripNoise(parts[0]))
			}
			continue
		}

		// Test failures: "--- FAIL: TestName (0.00s)"
		if strings.HasPrefix(trimmed, "--- FAIL:") {
			result.Type = "test failure"
			// Don't set message here, wait for actual error details
			continue
		}

		// Test error details: "    calculator_test.go:25: Expected 10, got 5"
		if result.Type == "test failure" && regexp.MustCompile(`^\w+_test\.go:\d+:`).MatchString(trimmed) {
			parts := strings.SplitN(trimmed, ": ", 2)
			if len(parts) >= 2 {
				// Set the first actual error message
				if result.Message == "" {
					result.Message = parts[1]
				}
				stackFrames = append(stackFrames, errclean.StripNoise(parts[0]))
			}
			continue
		}

		// Panic message: "panic: runtime error: invalid memory address"
		if strings.HasPrefix(trimmed, "panic:") {
			result.Type = "panic"
			result.Message = strings.TrimSpace(strings.TrimPrefix(trimmed, "panic:"))
			continue
		}

		// Fatal errors: "fatal error: concurrent map writes"
		if strings.HasPrefix(trimmed, "fatal error:") {
			result.Type = "fatal error"
			result.Message = strings.TrimSpace(strings.TrimPrefix(trimmed, "fatal error:"))
			continue
		}

		// Stack frames come in pairs:
		// functionName(args)
		//     /path/to/file.go:42 +0x123
		if i > 0 && strings.Contains(line, ".go:") {
			// This is the file:line part
			frame := errclean.StripNoise(trimmed)

			// Get the function name from previous line
			prevLine := strings.TrimSpace(lines[i-1])
			if prevLine != "" && !strings.HasPrefix(prevLine, "goroutine") &&
				!strings.HasPrefix(prevLine, "panic:") && !strings.HasPrefix(prevLine, "fatal error:") {
				// Combine function and location
				funcName := strings.Split(prevLine, "(")[0]
				stackFrames = append(stackFrames, funcName+" "+frame)
			}
		}
	}

	result.Stack = errclean.DeduplicateFrames(stackFrames)
	result.Message = errclean.StripNoise(result.Message)
	return result
}
