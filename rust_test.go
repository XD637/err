package main

import (
	"strings"
	"testing"
)

func TestCleanRust(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectedType string
		expectedMsg  string
		expectStack  bool
	}{
		{
			name: "Simple panic",
			input: `thread 'main' panicked at 'index out of bounds: the len is 3 but the index is 5', src/main.rs:42:5
note: run with ` + "`RUST_BACKTRACE=1`" + ` environment variable to display a backtrace`,
			expectedType: "panic",
			expectedMsg:  "index out of bounds",
			expectStack:  true,
		},
		{
			name: "Compile error",
			input: `error[E0382]: borrow of moved value: ` + "`s`" + `
  --> src/main.rs:5:20
   |
3  |     let s = String::from("hello");
   |         - move occurs because ` + "`s`" + ` has type ` + "`String`" + `, which does not implement the ` + "`Copy`" + ` trait
4  |     let s2 = s;
   |              - value moved here
5  |     println!("{}", s);
   |                    ^ value borrowed here after move`,
			expectedType: "E0382",
			expectedMsg:  "borrow of moved value",
			expectStack:  true,
		},
		{
			name: "Panic with backtrace",
			input: `thread 'main' panicked at 'assertion failed: x == y', src/lib.rs:10:5
stack backtrace:
   0: rust_begin_unwind
             at /rustc/d5a82bbd26e1ad8b7401f6a718a9c57c96905483/library/std/src/panicking.rs:575:5
   1: core::panicking::panic_fmt
             at /rustc/d5a82bbd26e1ad8b7401f6a718a9c57c96905483/library/core/src/panicking.rs:64:14
   2: myapp::calculate
             at ./src/lib.rs:10:5
   3: myapp::main
             at ./src/main.rs:5:5`,
			expectedType: "panic",
			expectedMsg:  "assertion failed",
			expectStack:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanRust(tt.input)

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
