package errclean

import (
	"regexp"
	"strings"
)

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

// StripNoise removes common noise from error messages
func StripNoise(text string) string {
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

// DeduplicateFrames removes consecutive duplicate stack frames
func DeduplicateFrames(frames []string) []string {
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
