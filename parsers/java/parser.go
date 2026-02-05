package java

import (
	"regexp"
	"strings"

	"github.com/XD637/err/errclean"
	"github.com/XD637/err/registry"
)

// Parser handles Java exceptions and stack traces
type Parser struct{}

func init() {
	registry.Register(&Parser{})
}

func (p *Parser) Name() string {
	return "java"
}

func (p *Parser) Detect(text string) int {
	lines := strings.Split(text, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Java exception header: definitive
		if strings.Contains(line, "Exception in thread") {
			return 100
		}

		// Java stack trace format: high confidence
		if strings.HasPrefix(line, "\tat ") {
			return 90
		}

		// Caused by: high confidence
		if strings.HasPrefix(line, "Caused by:") {
			return 85
		}

		// Java exception pattern
		if regexp.MustCompile(`^[a-z]+(\.[a-z]+)+\.[A-Z]\w*(Exception|Error):`).MatchString(line) {
			return 80
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

		// Exception type and message (first line or "Caused by:")
		if i == 0 || strings.HasPrefix(trimmed, "Caused by:") {
			// Format: "java.lang.NullPointerException: message"
			parts := strings.SplitN(trimmed, ":", 2)
			if len(parts) >= 1 {
				exType := strings.TrimPrefix(parts[0], "Caused by: ")
				// Extract simple class name
				typeParts := strings.Split(exType, ".")
				result.Type = typeParts[len(typeParts)-1]

				if len(parts) == 2 {
					result.Message = strings.TrimSpace(parts[1])
				}
			}
			continue
		}

		// Stack frames: "\tat com.example.Class.method(File.java:42)"
		if strings.HasPrefix(trimmed, "at ") {
			frame := errclean.StripNoise(trimmed)
			// Skip common framework noise
			if !strings.Contains(frame, "java.lang.reflect") &&
				!strings.Contains(frame, "sun.reflect") {
				stackFrames = append(stackFrames, frame)
			}
		}
	}

	result.Stack = errclean.DeduplicateFrames(stackFrames)
	result.Message = errclean.StripNoise(result.Message)
	return result
}
