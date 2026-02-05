package python

import (
	"regexp"
	"strings"

	"github.com/XD637/err/errclean"
	"github.com/XD637/err/registry"
)

// Parser handles Python errors and tracebacks
type Parser struct{}

func init() {
	registry.Register(&Parser{})
}

func (p *Parser) Name() string {
	return "python"
}

func (p *Parser) Detect(text string) int {
	lines := strings.Split(text, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Traceback: definitive Python
		if strings.Contains(line, "Traceback (most recent call last)") {
			return 100
		}

		// Python-specific syntax errors
		if regexp.MustCompile(`^(SyntaxError|IndentationError|TabError|ModuleNotFoundError|ImportError):`).MatchString(line) {
			return 95
		}

		// Python exception pattern
		if regexp.MustCompile(`^[A-Z]\w*(Error|Exception|Warning):`).MatchString(line) {
			return 70
		}

		// File pattern in traceback
		if strings.HasPrefix(line, "File ") && strings.Contains(line, ".py") {
			return 80
		}
	}

	return 0
}

func (p *Parser) Parse(text string) *errclean.CleanedError {
	lines := strings.Split(text, "\n")
	result := &errclean.CleanedError{}

	var stackFrames []string
	inTraceback := false

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Start of traceback
		if strings.Contains(trimmed, "Traceback (most recent call last)") {
			inTraceback = true
			continue
		}

		// Syntax errors have a different format
		if strings.HasPrefix(trimmed, "SyntaxError:") ||
			strings.HasPrefix(trimmed, "IndentationError:") ||
			strings.HasPrefix(trimmed, "TabError:") {
			parts := strings.SplitN(trimmed, ":", 2)
			result.Type = strings.TrimSpace(parts[0])
			if len(parts) == 2 {
				result.Message = strings.TrimSpace(parts[1])
			}
			// Look for file location in previous lines
			if i > 0 {
				for j := i - 1; j >= 0 && j >= i-3; j-- {
					prevLine := strings.TrimSpace(lines[j])
					if strings.HasPrefix(prevLine, "File ") {
						stackFrames = append(stackFrames, errclean.StripNoise(prevLine))
						break
					}
				}
			}
			break
		}

		// File location: '  File "/path/to/file.py", line 42, in function'
		if inTraceback && strings.HasPrefix(line, "  File ") {
			frame := errclean.StripNoise(trimmed)
			stackFrames = append(stackFrames, frame)
			continue
		}

		// Exception type and message - various formats
		exceptionPattern := regexp.MustCompile(`^([A-Z]\w*(?:Error|Exception|Warning)):\s*(.*)`)
		if exceptionPattern.MatchString(trimmed) {
			matches := exceptionPattern.FindStringSubmatch(trimmed)
			if len(matches) >= 2 {
				result.Type = strings.TrimSpace(matches[1])
				if len(matches) >= 3 {
					result.Message = strings.TrimSpace(matches[2])
				}
			}
			// If we're in a traceback, this is the final exception
			if inTraceback {
				break
			}
			// Otherwise, continue looking for more context
		}

		// Handle cases where exception has no message
		if regexp.MustCompile(`^[A-Z]\w*(?:Error|Exception|Warning)$`).MatchString(trimmed) {
			result.Type = trimmed
			if inTraceback {
				break
			}
		}
	}

	result.Stack = errclean.DeduplicateFrames(stackFrames)
	result.Message = errclean.StripNoise(result.Message)
	return result
}
