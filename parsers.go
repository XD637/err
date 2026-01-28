package main

import (
	"regexp"
	"strings"
)

// cleanJavaScript processes JavaScript/Node.js errors
func cleanJavaScript(text string) *CleanedError {
	lines := strings.Split(text, "\n")
	result := &CleanedError{}

	var stackFrames []string

	for i, line := range lines {
		line = strings.TrimSpace(line)

		// Error type and message (first line usually)
		if i == 0 || (result.Type == "" && regexp.MustCompile(`[A-Z]\w*Error:`).MatchString(line)) {
			// Look for error pattern in the line
			errorPattern := regexp.MustCompile(`([A-Z]\w*Error):\s*(.+)`)
			matches := errorPattern.FindStringSubmatch(line)
			if len(matches) >= 3 {
				result.Type = matches[1]
				result.Message = strings.TrimSpace(matches[2])
			} else {
				result.Message = line
			}
			continue
		}

		// Stack frames: "    at functionName (file:line:col)"
		if strings.HasPrefix(line, "at ") {
			frame := stripNoise(line)
			stackFrames = append(stackFrames, frame)
		}
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

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Start of traceback
		if strings.Contains(trimmed, "Traceback (most recent call last)") {
			inTraceback = true
			continue
		}

		// File location: '  File "/path/to/file.py", line 42, in function'
		if inTraceback && strings.HasPrefix(line, "  File ") {
			frame := stripNoise(trimmed)
			stackFrames = append(stackFrames, frame)
			continue
		}

		// Exception type and message
		if inTraceback && regexp.MustCompile(`^[A-Z]\w*(Error|Exception):`).MatchString(trimmed) {
			parts := strings.SplitN(trimmed, ":", 2)
			result.Type = strings.TrimSpace(parts[0])
			if len(parts) == 2 {
				result.Message = strings.TrimSpace(parts[1])
			}
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

// cleanGo processes Go panic traces
func cleanGo(text string) *CleanedError {
	lines := strings.Split(text, "\n")
	result := &CleanedError{}

	var stackFrames []string

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Panic message: "panic: runtime error: invalid memory address"
		if strings.HasPrefix(trimmed, "panic:") {
			result.Type = "panic"
			result.Message = strings.TrimSpace(strings.TrimPrefix(trimmed, "panic:"))
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
			if prevLine != "" && !strings.HasPrefix(prevLine, "goroutine") {
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

// cleanRust processes Rust panic messages
func cleanRust(text string) *CleanedError {
	lines := strings.Split(text, "\n")
	result := &CleanedError{}

	var stackFrames []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Panic message: "thread 'main' panicked at 'message', src/main.rs:42:5"
		if strings.Contains(trimmed, "panicked at") {
			result.Type = "panic"

			// Extract message between quotes
			re := regexp.MustCompile(`panicked at '([^']+)'`)
			matches := re.FindStringSubmatch(trimmed)
			if len(matches) > 1 {
				result.Message = matches[1]
			}
			continue
		}

		// Stack frames (from RUST_BACKTRACE=1)
		// Format varies, but typically: "  42: function_name"
		if regexp.MustCompile(`^\s*\d+:`).MatchString(trimmed) {
			frame := stripNoise(trimmed)
			// Skip std library internals
			if !strings.Contains(frame, "std::") &&
				!strings.Contains(frame, "core::") {
				stackFrames = append(stackFrames, frame)
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
