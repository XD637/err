package javascript

import (
	"strings"
	"testing"
)

func TestJavaScriptParser(t *testing.T) {
	parser := &Parser{}

	tests := []struct {
		name         string
		input        string
		expectedType string
		expectedMsg  string
		expectStack  bool
	}{
		{
			name: "Standard TypeError",
			input: `TypeError: Cannot read property 'foo' of undefined
    at Object.<anonymous> (/Users/dev/app/index.js:42:5)
    at Module._compile (internal/modules/cjs/loader.js:1063:30)`,
			expectedType: "TypeError",
			expectedMsg:  "Cannot read property 'foo' of undefined",
			expectStack:  true,
		},
		{
			name: "Unhandled Promise Rejection",
			input: `2024-01-28T14:10:36.123Z UnhandledPromiseRejectionWarning: Error: Database connection failed
    at Database.connect (/home/user/projects/myapp/src/database.js:42:15)`,
			expectedType: "Error",
			expectedMsg:  "Database connection failed",
			expectStack:  true,
		},
		{
			name: "TypeScript compile error",
			input: `src/index.ts:42:5 - error TS2322: Type 'string' is not assignable to type 'number'.

42     const count: number = "hello";
       ~~~~~`,
			expectedType: "TS2322",
			expectedMsg:  "Type 'string' is not assignable to type 'number'.",
			expectStack:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.Parse(tt.input)

			if result.Type != tt.expectedType {
				t.Errorf("Type = %v, want %v", result.Type, tt.expectedType)
			}

			if !strings.Contains(result.Message, tt.expectedMsg) {
				t.Errorf("Message should contain %q, got %v", tt.expectedMsg, result.Message)
			}

			if tt.expectStack && len(result.Stack) == 0 {
				t.Error("Stack should not be empty")
			}
		})
	}
}

func TestJavaScriptDetect(t *testing.T) {
	parser := &Parser{}

	tests := []struct {
		name     string
		input    string
		minScore int
	}{
		{
			name:     "TypeScript error",
			input:    "src/index.ts:42:5 - error TS2322: Type error",
			minScore: 100,
		},
		{
			name:     "Promise rejection",
			input:    "UnhandledPromiseRejectionWarning: Error: failed",
			minScore: 95,
		},
		{
			name:     "Standard error",
			input:    "TypeError: Cannot read property",
			minScore: 80,
		},
		{
			name:     "Not JavaScript",
			input:    "Traceback (most recent call last):",
			minScore: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := parser.Detect(tt.input)
			if score < tt.minScore {
				t.Errorf("Detect() = %v, want >= %v", score, tt.minScore)
			}
		})
	}
}
