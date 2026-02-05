package rust

import (
	"regexp"
	"strings"

	"github.com/XD637/err/errclean"
	"github.com/XD637/err/registry"
)

// Parser handles Rust compile errors, panics, and backtraces
type Parser struct{}

func init() {
	registry.Register(&Parser{})
}

func (p *Parser) Name() string {
	return "rust"
}

func (p *Parser) Detect(text string) int {
	lines := strings.Split(text, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Rust compile errors: definitive
		if strings.HasPrefix(line, "error[E") && strings.Contains(line, "]:") {
			return 100
		}

		// Rust panic: high confidence
		if strings.Contains(line, "panicked at") {
			return 95
		}

		// Backtrace header
		if strings.Contains(line, "stack backtrace:") {
			return 90
		}

		// Rust file with arrow
		if strings.Contains(line, " --> ") && strings.Contains(line, ".rs:") {
			return 85
		}
	}

	return 0
}

func (p *Parser) Parse(text string) *errclean.CleanedError {
	lines := strings.Split(text, "\n")
	result := &errclean.CleanedError{}

	var stackFrames []string
	inBacktrace := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Compile errors: "error[E0382]: borrow of moved value: `s`"
		if strings.HasPrefix(trimmed, "error[E") {
			compileErrorPattern := regexp.MustCompile(`error\[(E\d+)\]:\s*(.+)`)
			matches := compileErrorPattern.FindStringSubmatch(trimmed)
			if len(matches) >= 3 {
				result.Type = matches[1]
				result.Message = matches[2]
			}
			continue
		}

		// File location for compile errors: "  --> src/main.rs:5:20"
		if (strings.Contains(trimmed, " --> ") || strings.Contains(trimmed, "-->")) && strings.Contains(trimmed, ".rs:") {
			var location string
			if strings.Contains(trimmed, " --> ") {
				parts := strings.Split(trimmed, " --> ")
				if len(parts) >= 2 {
					location = parts[1]
				}
			} else if strings.Contains(trimmed, "-->") {
				parts := strings.Split(trimmed, "-->")
				if len(parts) >= 2 {
					location = strings.TrimSpace(parts[1])
				}
			}
			if location != "" {
				stackFrames = append(stackFrames, errclean.StripNoise(location))
			}
			continue
		}

		// Panic message: "thread 'main' panicked at 'message', src/main.rs:42:5"
		if strings.Contains(trimmed, "panicked at") {
			if result.Type == "" {
				result.Type = "panic"
			}

			// Extract message between quotes
			re := regexp.MustCompile(`panicked at '([^']+)'`)
			matches := re.FindStringSubmatch(trimmed)
			if len(matches) > 1 {
				result.Message = matches[1]
			}

			// Extract file location
			filePattern := regexp.MustCompile(`([^,]+\.rs:\d+:\d+)`)
			fileMatches := filePattern.FindStringSubmatch(trimmed)
			if len(fileMatches) > 1 {
				stackFrames = append(stackFrames, errclean.StripNoise(fileMatches[1]))
			}
			continue
		}

		// Backtrace header
		if strings.Contains(trimmed, "stack backtrace:") {
			inBacktrace = true
			continue
		}

		// Stack frames (from RUST_BACKTRACE=1)
		// Format: "  42: function_name" or "   0: rust_begin_unwind"
		if inBacktrace && regexp.MustCompile(`^\d+:`).MatchString(trimmed) {
			frame := errclean.StripNoise(trimmed)
			// Skip std library internals unless they're the only frames
			if (!strings.Contains(frame, "std::") &&
				!strings.Contains(frame, "core::") &&
				!strings.Contains(frame, "rust_begin_unwind")) || len(stackFrames) == 0 {
				stackFrames = append(stackFrames, frame)
			}
			continue
		}

		// File location in backtrace: "             at /path/to/file.rs:42:5"
		if inBacktrace && strings.Contains(trimmed, "at ") && strings.Contains(trimmed, ".rs:") {
			parts := strings.Split(trimmed, "at ")
			if len(parts) >= 2 {
				location := errclean.StripNoise(parts[1])
				// Add to last frame if it doesn't have a location
				if len(stackFrames) > 0 && !strings.Contains(stackFrames[len(stackFrames)-1], ".rs:") {
					stackFrames[len(stackFrames)-1] += " " + location
				}
			}
		}
	}

	result.Stack = errclean.DeduplicateFrames(stackFrames)
	result.Message = errclean.StripNoise(result.Message)
	return result
}
