package registry

import (
	"sort"
	"sync"

	"github.com/XD637/err/parsers"
)

// Registry manages all registered error parsers
type Registry struct {
	mu      sync.RWMutex
	parsers []parsers.Parser
	byName  map[string]parsers.Parser
}

// Global registry instance
var global = &Registry{
	byName: make(map[string]parsers.Parser),
}

// Register adds a parser to the global registry
func Register(p parsers.Parser) {
	global.Register(p)
}

// Register adds a parser to this registry
func (r *Registry) Register(p parsers.Parser) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.parsers = append(r.parsers, p)
	r.byName[p.Name()] = p
}

// DetectParser returns the best matching parser for the given text
// based on confidence scores from each parser's Detect method
func (r *Registry) DetectParser(text string) parsers.Parser {
	r.mu.RLock()
	defer r.mu.RUnlock()

	type result struct {
		parser     parsers.Parser
		confidence int
	}

	results := make([]result, 0, len(r.parsers))
	for _, p := range r.parsers {
		confidence := p.Detect(text)
		if confidence > 0 {
			results = append(results, result{p, confidence})
		}
	}

	if len(results) == 0 {
		return nil
	}

	// Sort by confidence (highest first)
	sort.Slice(results, func(i, j int) bool {
		return results[i].confidence > results[j].confidence
	})

	return results[0].parser
}

// GetParser returns a parser by name, or nil if not found
func (r *Registry) GetParser(name string) parsers.Parser {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.byName[name]
}

// DetectParser returns the best matching parser from the global registry
func DetectParser(text string) parsers.Parser {
	return global.DetectParser(text)
}

// GetParser returns a parser by name from the global registry
func GetParser(name string) parsers.Parser {
	return global.GetParser(name)
}

// AllParsers returns all registered parsers
func AllParsers() []parsers.Parser {
	global.mu.RLock()
	defer global.mu.RUnlock()

	result := make([]parsers.Parser, len(global.parsers))
	copy(result, global.parsers)
	return result
}
