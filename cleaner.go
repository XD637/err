package main

import (
	"github.com/XD637/err/errclean"
	"github.com/XD637/err/registry"

	// Import all parsers to register them
	_ "github.com/XD637/err/parsers/golang"
	_ "github.com/XD637/err/parsers/java"
	_ "github.com/XD637/err/parsers/javascript"
	_ "github.com/XD637/err/parsers/python"
	_ "github.com/XD637/err/parsers/rust"
)

// Cleaner processes error messages using registered parsers
type Cleaner struct {
	format string
}

// NewCleaner creates a new error cleaner
func NewCleaner(format string) *Cleaner {
	return &Cleaner{format: format}
}

// Clean processes the error text using the appropriate parser
func (c *Cleaner) Clean(text string) *errclean.CleanedError {
	var parser interface {
		Parse(string) *errclean.CleanedError
	}

	if c.format == "auto" {
		// Auto-detect the best parser
		parser = registry.DetectParser(text)
		if parser == nil {
			// Fallback to generic parsing
			return &errclean.CleanedError{
				Type:    "error",
				Message: errclean.StripNoise(text),
			}
		}
	} else {
		// Use specified parser
		parser = registry.GetParser(c.format)
		if parser == nil {
			// Fallback to generic parsing
			return &errclean.CleanedError{
				Type:    "error",
				Message: errclean.StripNoise(text),
			}
		}
	}

	return parser.Parse(text)
}
