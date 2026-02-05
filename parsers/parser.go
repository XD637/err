package parsers

import "github.com/XD637/err/errclean"

// Parser defines the interface that all language-specific parsers must implement
type Parser interface {
	// Name returns the unique identifier for this parser (e.g., "javascript", "python")
	Name() string

	// Detect returns a confidence score (0-100) indicating how likely this parser
	// can handle the given error text. Higher scores indicate better matches.
	// Return 100 for definitive matches, 0 for no match.
	Detect(text string) int

	// Parse processes the error text and returns a cleaned error structure
	Parse(text string) *errclean.CleanedError
}
