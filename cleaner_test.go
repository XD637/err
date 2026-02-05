package main

import (
	"strings"
	"testing"
)

func TestDetectFormat(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "JavaScript error",
			input: `TypeError: Cannot read property 'foo' of undefined
    at Object.<anonymous> (/app/index.js:42:5)`,
			expected: "javascript",
		},
		{
			name: "Python traceback",
			input: `Traceback (most recent call last):
  File "main.py", line 10, in <module>
    raise ValueError("test")
ValueError: test`,
			expected: "python",
		},
		{
			name: "Java exception",
			input: `Exception in thread "main" java.lang.NullPointerException
	at com.example.Main.main(Main.java:42)`,
			expected: "java",
		},
		{
			name: "Go panic",
			input: `panic: runtime error: invalid memory address or nil pointer dereference
goroutine 1 [running]:
main.main()
	/app/main.go:42 +0x123`,
			expected: "go",
		},
		{
			name:     "Rust panic",
			input:    `thread 'main' panicked at 'index out of bounds', src/main.rs:42:5`,
			expected: "rust",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detectFormat(tt.input)
			if result != tt.expected {
				t.Errorf("detectFormat() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestStripNoise(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		contains    []string
		notContains []string
	}{
		{
			name:        "timestamps",
			input:       "2024-01-28T14:10:36Z Error occurred",
			contains:    []string{"[TIME]", "Error occurred"},
			notContains: []string{"2024-01-28"},
		},
		{
			name:        "memory addresses",
			input:       "Segfault at 0x7f8a9b0c1d2e",
			contains:    []string{"[ADDR]"},
			notContains: []string{"0x7f8a9b0c1d2e"},
		},
		{
			name:        "UUIDs",
			input:       "Request 550e8400-e29b-41d4-a716-446655440000 failed",
			contains:    []string{"[UUID]", "failed"},
			notContains: []string{"550e8400"},
		},
		{
			name:        "hex values",
			input:       "Error code: 0xdeadbeef",
			contains:    []string{"[HEX]"},
			notContains: []string{"0xdeadbeef"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stripNoise(tt.input)

			for _, want := range tt.contains {
				if !strings.Contains(result, want) {
					t.Errorf("stripNoise() result should contain %q, got %q", want, result)
				}
			}

			for _, notWant := range tt.notContains {
				if strings.Contains(result, notWant) {
					t.Errorf("stripNoise() result should not contain %q, got %q", notWant, result)
				}
			}
		})
	}
}

func TestDeduplicateFrames(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "no duplicates",
			input:    []string{"frame1", "frame2", "frame3"},
			expected: []string{"frame1", "frame2", "frame3"},
		},
		{
			name:     "consecutive duplicates",
			input:    []string{"frame1", "frame1", "frame2", "frame2", "frame2"},
			expected: []string{"frame1", "frame2"},
		},
		{
			name:     "empty",
			input:    []string{},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := deduplicateFrames(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("deduplicateFrames() length = %v, want %v", len(result), len(tt.expected))
				return
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("deduplicateFrames()[%d] = %v, want %v", i, result[i], tt.expected[i])
				}
			}
		})
	}
}

func TestCleanJavaScript(t *testing.T) {
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
			result := cleanJavaScript(tt.input)

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

func TestCleanPython(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectedType string
		expectedMsg  string
		expectStack  bool
	}{
		{
			name: "Standard ValueError",
			input: `Traceback (most recent call last):
  File "/home/user/app/main.py", line 42, in <module>
    raise ValueError("invalid input")
ValueError: invalid input`,
			expectedType: "ValueError",
			expectedMsg:  "invalid input",
			expectStack:  true,
		},
		{
			name: "Syntax Error",
			input: `  File "/home/user/projects/app/script.py", line 15
    if x == 5
            ^
SyntaxError: invalid syntax`,
			expectedType: "SyntaxError",
			expectedMsg:  "invalid syntax",
			expectStack:  true,
		},
		{
			name: "Import Error",
			input: `Traceback (most recent call last):
  File "/home/user/projects/app/main.py", line 3, in <module>
    from mymodule import something
  File "/home/user/projects/app/mymodule.py", line 10, in <module>
    import nonexistent_package
ModuleNotFoundError: No module named 'nonexistent_package'`,
			expectedType: "ModuleNotFoundError",
			expectedMsg:  "No module named 'nonexistent_package'",
			expectStack:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanPython(tt.input)

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

func TestCleanGo(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectedType string
		expectedMsg  string
		expectStack  bool
	}{
		{
			name: "Standard panic",
			input: `panic: runtime error: invalid memory address or nil pointer dereference
[signal SIGSEGV: segmentation violation code=0x1 addr=0x0 pc=0x123]

goroutine 1 [running]:
main.processData(...)
	/home/user/app/main.go:42 +0x123
main.main()
	/home/user/app/main.go:10 +0x45`,
			expectedType: "panic",
			expectedMsg:  "runtime error: invalid memory address or nil pointer dereference",
			expectStack:  true,
		},
		{
			name: "Build error",
			input: `# github.com/user/myapp
./main.go:15:2: undefined: fmt.Printl
./main.go:20:15: cannot use "hello" (type untyped string) as type int in assignment`,
			expectedType: "build error",
			expectedMsg:  "undefined: fmt.Printl",
			expectStack:  true,
		},
		{
			name: "Test failure",
			input: `--- FAIL: TestCalculate (0.00s)
    calculator_test.go:25: Expected 10, got 5
    calculator_test.go:30: Expected true, got false
FAIL`,
			expectedType: "test failure",
			expectedMsg:  "Expected 10, got 5",
			expectStack:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanGo(tt.input)

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

func TestCleanedErrorFormat(t *testing.T) {
	err := &CleanedError{
		Type:    "TypeError",
		Message: "Cannot read property 'foo' of undefined",
		Stack:   []string{"at main.js:42", "at app.js:10"},
	}

	result := err.Format()

	if !strings.Contains(result, "TypeError") {
		t.Error("Format should contain type")
	}

	if !strings.Contains(result, "Cannot read property") {
		t.Error("Format should contain message")
	}

	if !strings.Contains(result, "main.js:42") {
		t.Error("Format should contain stack frames")
	}
}
