package main

import (
	"strings"
	"testing"
)

func TestCleaner(t *testing.T) {
	tests := []struct {
		name         string
		format       string
		input        string
		expectedType string
	}{
		{
			name:         "Auto-detect JavaScript",
			format:       "auto",
			input:        "TypeError: Cannot read property 'foo' of undefined",
			expectedType: "TypeError",
		},
		{
			name:         "Auto-detect Python",
			format:       "auto",
			input:        "Traceback (most recent call last):\nValueError: test",
			expectedType: "ValueError",
		},
		{
			name:         "Auto-detect Go",
			format:       "auto",
			input:        "panic: runtime error",
			expectedType: "panic",
		},
		{
			name:         "Auto-detect Rust",
			format:       "auto",
			input:        "error[E0382]: borrow of moved value",
			expectedType: "E0382",
		},
		{
			name:         "Explicit format",
			format:       "python",
			input:        "ValueError: test error",
			expectedType: "ValueError",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleaner := NewCleaner(tt.format)
			result := cleaner.Clean(tt.input)

			if !strings.Contains(result.Type, tt.expectedType) {
				t.Errorf("Type = %v, want to contain %v", result.Type, tt.expectedType)
			}
		})
	}
}
