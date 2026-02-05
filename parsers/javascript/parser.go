package javascript

import (
	"regexp"
	"strings"

	"github.com/XD637/err/errclean"
	"github.com/XD637/err/registry"
)

// Parser handles JavaScript, Node.js, and TypeScript errors
type Parser struct{}

// Register this parser on package import
func init() {
	registry.Register(&Parser{})
}

// Name returns the parser identifier
func (p *Parser) Name() string {
	return "javascript"
}

// Detect returns confidence score for JavaScript/TypeScript errors
func (p *Parser) Detect(text string) int {
	lines := strings.Split(text, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// npm errors: high confidence
		if strings.HasPrefix(line, "npm ERR!") {
			return 100
		}

		// TypeScript compile errors: high confidence
		if strings.Contains(line, " - error TS") {
			return 100
		}

		// Unhandled promise rejections: high confidence
		if strings.Contains(line, "UnhandledPromiseRejectionWarning:") {
			return 95
		}

		// Standard JavaScript errors: medium-high confidence
		if regexp.MustCompile(`(TypeError|ReferenceError|SyntaxError|Error):`).MatchString(line) {
			return 80
		}

		// Stack trace format: medium confidence (could be other languages)
		if strings.HasPrefix(line, "at ") {
			return 60
		}
	}

	return 0
}

// Parse processes JavaScript/TypeScript error text
func (p *Parser) Parse(text string) *errclean.CleanedError {
	lines := strings.Split(text, "\n")
	result := &errclean.CleanedError{}

	var stackFrames []string
	foundError := false

	// Check for npm errors first
	if strings.Contains(text, "npm ERR!") {
		return parseNpmError(lines)
	}

	for i, line := range lines {
		line = strings.TrimSpace(line)

		// TypeScript compile errors: "src/file.ts:42:5 - error TS2322: message"
		if strings.Contains(line, " - error TS") {
			tsPattern := regexp.MustCompile(`error (TS\d+):\s*(.+)`)
			matches := tsPattern.FindStringSubmatch(line)
			if len(matches) >= 3 {
				result.Type = matches[1]
				result.Message = strings.TrimSpace(matches[2])
				foundError = true
			}
			continue
		}

		// Unhandled promise rejections
		if strings.Contains(line, "UnhandledPromiseRejectionWarning:") {
			parts := strings.SplitN(line, "UnhandledPromiseRejectionWarning:", 2)
			if len(parts) == 2 {
				msg := strings.TrimSpace(parts[1])
				// Extract error type if present - look for "Error: message" pattern
				// Use a simpler pattern: any word followed by colon
				errorPattern := regexp.MustCompile(`^([A-Z]\w+):\s*(.+)`)
				matches := errorPattern.FindStringSubmatch(msg)
				if len(matches) >= 3 && (strings.HasSuffix(matches[1], "Error") ||
					strings.HasSuffix(matches[1], "Exception") ||
					strings.HasSuffix(matches[1], "Warning")) {
					result.Type = matches[1]
					result.Message = strings.TrimSpace(matches[2])
				} else {
					result.Type = "UnhandledPromiseRejection"
					result.Message = msg
				}
				foundError = true
			}
			continue
		}

		// Standard error pattern: "TypeError: message"
		if !foundError && (i == 0 || regexp.MustCompile(`^[A-Z]\w*Error:`).MatchString(line)) {
			errorPattern := regexp.MustCompile(`^([A-Z]\w*(?:Error|Exception|Warning)):\s*(.+)`)
			matches := errorPattern.FindStringSubmatch(line)
			if len(matches) >= 3 {
				result.Type = matches[1]
				result.Message = strings.TrimSpace(matches[2])
				foundError = true
			} else if i == 0 && !strings.HasPrefix(line, "at ") {
				result.Message = line
				foundError = true
			}
			continue
		}

		// Stack frames: "    at functionName (file:line:col)"
		if strings.HasPrefix(line, "at ") {
			frame := errclean.StripNoise(line)
			// Skip Node.js internal frames unless they're the only ones
			if !strings.Contains(frame, "internal/") || len(stackFrames) == 0 {
				stackFrames = append(stackFrames, frame)
			}
		}
	}

	// If no error type found, default to Error
	if result.Type == "" && result.Message != "" {
		result.Type = "Error"
	}

	result.Stack = errclean.DeduplicateFrames(stackFrames)
	result.Message = errclean.StripNoise(result.Message)
	return result
}

// parseNpmError handles npm-specific error formats
func parseNpmError(lines []string) *errclean.CleanedError {
	result := &errclean.CleanedError{
		Type: "npm",
	}

	var errorCode string
	var mainMessage string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Remove "npm ERR!" prefix
		if strings.HasPrefix(trimmed, "npm ERR!") {
			content := strings.TrimSpace(strings.TrimPrefix(trimmed, "npm ERR!"))

			// Extract error code: "code ENOENT"
			if strings.HasPrefix(content, "code ") {
				errorCode = strings.TrimPrefix(content, "code ")
				continue
			}

			// Extract main error message (usually starts with error code)
			if errorCode != "" && strings.HasPrefix(content, strings.ToLower(errorCode)) {
				mainMessage = content
				break
			}

			// Fallback: capture first substantial message
			if mainMessage == "" && len(content) > 10 && !strings.HasPrefix(content, "syscall") &&
				!strings.HasPrefix(content, "path") && !strings.HasPrefix(content, "errno") {
				mainMessage = content
			}
		}
	}

	if errorCode != "" {
		result.Type = "npm " + errorCode
	}

	if mainMessage != "" {
		result.Message = errclean.StripNoise(mainMessage)
	} else if errorCode != "" {
		result.Message = errorCode
	}

	return result
}
