package main

import (
	"regexp"
	"strings"
)

// cleanJavaScript processes JavaScript/Node.js/TypeScript errors
func cleanJavaScript(text string) *CleanedError {
	lines := strings.Split(text, "\n")
	result := &CleanedError{}

	var stackFrames []string
	foundError := false

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
				// Match "Error", "TypeError", "ReferenceError", etc.
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
			frame := stripNoise(line)
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

	result.Stack = deduplicateFrames(stackFrames)
	result.Message = stripNoise(result.Message)
	return result
}

// cleanPython processes Python tracebacks
func cleanPython(text string) *CleanedError {
	lines := strings.Split(text, "\n")
	result := &CleanedError{}

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
						stackFrames = append(stackFrames, stripNoise(prevLine))
						break
					}
				}
			}
			break
		}

		// File location: '  File "/path/to/file.py", line 42, in function'
		if inTraceback && strings.HasPrefix(line, "  File ") {
			frame := stripNoise(trimmed)
			stackFrames = append(stackFrames, frame)
			continue
		}

		// Exception type and message - various formats
		exceptionPattern := regexp.MustCompile(`^([A-Z]\w*(?:Error|Exception|Warning)):\s*(.*)`)
		if inTraceback && exceptionPattern.MatchString(trimmed) {
			matches := exceptionPattern.FindStringSubmatch(trimmed)
			if len(matches) >= 2 {
				result.Type = strings.TrimSpace(matches[1])
				if len(matches) >= 3 {
					result.Message = strings.TrimSpace(matches[2])
				}
			}
			break
		}

		// Handle cases where exception has no message
		if inTraceback && regexp.MustCompile(`^[A-Z]\w*(?:Error|Exception|Warning)$`).MatchString(trimmed) {
			result.Type = trimmed
			break
		}
	}

	result.Stack = deduplicateFrames(stackFrames)
	result.Message = stripNoise(result.Message)
	return result
}

// cleanJava processes Java stack traces
func cleanJava(text string) *CleanedError {
	lines := strings.Split(text, "\n")
	result := &CleanedError{}

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
			frame := stripNoise(trimmed)
			// Skip common framework noise
			if !strings.Contains(frame, "java.lang.reflect") &&
				!strings.Contains(frame, "sun.reflect") {
				stackFrames = append(stackFrames, frame)
			}
		}
	}

	result.Stack = deduplicateFrames(stackFrames)
	result.Message = stripNoise(result.Message)
	return result
}

// cleanGo processes Go panic traces, build errors, and test failures
func cleanGo(text string) *CleanedError {
	lines := strings.Split(text, "\n")
	result := &CleanedError{}

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
				stackFrames = append(stackFrames, stripNoise(parts[0]))
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
				stackFrames = append(stackFrames, stripNoise(parts[0]))
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
			frame := stripNoise(trimmed)

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

	result.Stack = deduplicateFrames(stackFrames)
	result.Message = stripNoise(result.Message)
	return result
}

// cleanRust processes Rust panic messages, compile errors, and backtraces
func cleanRust(text string) *CleanedError {
	lines := strings.Split(text, "\n")
	result := &CleanedError{}

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
				stackFrames = append(stackFrames, stripNoise(location))
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
				stackFrames = append(stackFrames, stripNoise(fileMatches[1]))
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
			frame := stripNoise(trimmed)
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
				location := stripNoise(parts[1])
				// Add to last frame if it doesn't have a location
				if len(stackFrames) > 0 && !strings.Contains(stackFrames[len(stackFrames)-1], ".rs:") {
					stackFrames[len(stackFrames)-1] += " " + location
				}
			}
		}
	}

	result.Stack = deduplicateFrames(stackFrames)
	result.Message = stripNoise(result.Message)
	return result
}

// cleanGeneric handles unknown error formats
func cleanGeneric(text string) *CleanedError {
	lines := strings.Split(text, "\n")
	result := &CleanedError{}

	// Try to extract first meaningful line as message
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			result.Message = stripNoise(trimmed)
			break
		}
	}

	// Look for anything that looks like a stack frame
	var stackFrames []string
	for _, line := range lines[1:] {
		trimmed := strings.TrimSpace(line)

		// Common stack frame indicators
		if strings.Contains(trimmed, "at ") ||
			strings.Contains(trimmed, "File ") ||
			regexp.MustCompile(`\.\w+:\d+`).MatchString(trimmed) {
			frame := stripNoise(trimmed)
			stackFrames = append(stackFrames, frame)
		}
	}

	result.Stack = deduplicateFrames(stackFrames)
	result.Type = "error"

	return result
}
