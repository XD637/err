# err

Clean and normalize error messages from **Rust, Go, Python, Node.js/JavaScript, TypeScript, and Java**.

**Copy any error. Pipe to `err`. Get clean output.**

Strips timestamps, paths, memory addresses, and noise. Shows you what matters.

## Features

-  **Multi-language support**: Rust, Go, Python, JavaScript/Node.js, TypeScript, Java
-  **Auto-detection**: Automatically identifies error format
-  **Comprehensive error types**: Handles compile errors, runtime errors, panics, test failures, and more
-  **Smart filtering**: Removes noise while preserving critical information
-  **Stack trace cleaning**: Deduplicates frames and removes irrelevant internals

## Installation

### Brew (Recommended)

```bash
brew install XD637/err/err
```

### Go 

```bash
go install github.com/XD637/err@latest
```

### Manual Download

Download the latest binary from [releases page](https://github.com/XD637/err/releases) (when available).

### Build from Source

```bash
git clone https://github.com/XD637/err.git
cd err
go build -o err
# Move to your PATH
mv err /usr/local/bin/  # Linux/Mac
```

## Usage

```bash
# Pipe from commands (primary usage)
npm test 2>&1 | err
python script.py 2>&1 | err
go run main.go 2>&1 | err
cargo build 2>&1 | err
tsc 2>&1 | err

# From file
err error.log

# Save and process
npm test 2>&1 | tee error.log | err

# Verbose output
err -v error.log

# Specific format
err -format python < traceback.txt
```

## Supported Languages & Error Types

### JavaScript/Node.js/TypeScript

**Supported error types:**
-  Standard errors (TypeError, ReferenceError, SyntaxError, etc.)
-  Unhandled promise rejections
-  TypeScript compile errors (TS2322, TS2339, etc.)
-  Async/await errors
-  Stack traces with source maps

**Example:**
```bash
npm test 2>&1 | err
# TypeError: Cannot read property 'foo' of undefined
#   at Object.<anonymous> (index.js:42:5)
#   at Module._compile (loader.js:1063:30)
```

### Python

**Supported error types:**
-  Standard exceptions (ValueError, TypeError, AttributeError, etc.)
-  Syntax errors
-  Import/Module errors
-  Indentation errors
-  Full tracebacks

**Example:**
```bash
python script.py 2>&1 | err
# ModuleNotFoundError: No module named 'nonexistent_package'
#   File "main.py", line 3, in <module>
#   File "mymodule.py", line 10, in <module>
```

### Go

**Supported error types:**
-  Panics (runtime errors, nil pointer dereference, etc.)
-  Build errors (undefined references, type mismatches, etc.)
-  Test failures
-  Fatal errors (concurrent map writes, etc.)
-  Goroutine stack traces

**Example:**
```bash
go build 2>&1 | err
# build error: undefined: fmt.Printl
#   ./main.go:15:2
#   ./main.go:20:15
```

### Rust

**Supported error types:**
-  Compile errors (E0382, E0277, etc.)
-  Panics (index out of bounds, unwrap on None, etc.)
-  Assertion failures
-  Backtraces (with RUST_BACKTRACE=1)
-  Borrow checker errors

**Example:**
```bash
cargo build 2>&1 | err
# E0382: borrow of moved value: `s`
#   src/main.rs:5:20
```

### Java

**Supported error types:**
-  Exceptions (NullPointerException, IllegalArgumentException, etc.)
-  Stack traces
-  Caused by chains
-  Filtered framework internals

## What it does

Reads error messages and outputs clean, normalized versions by:

- Extracting error type and core message
- Removing timestamps, memory addresses, UUIDs, hex values
- Simplifying file paths to filenames
- Filtering relevant stack frames
- Deduplicating repeated frames
- Removing language-specific internals (std lib, framework code)

## Examples

### TypeScript Compile Error

Input:
```
src/index.ts:42:5 - error TS2322: Type 'string' is not assignable to type 'number'.

42     const count: number = "hello";
       ~~~~~
```

Output:
```
TS2322: Type 'string' is not assignable to type 'number'.
```

### Rust Compile Error

Input:
```
error[E0382]: borrow of moved value: `s`
  --> src/main.rs:5:20
   |
3  |     let s = String::from("hello");
   |         - move occurs because `s` has type `String`
5  |     println!("{}", s);
   |                    ^ value borrowed here after move
```

Output:
```
E0382: borrow of moved value: `s`
  src/main.rs:5:20
```

### Go Build Error

Input:
```
# github.com/user/myapp
./main.go:15:2: undefined: fmt.Printl
./main.go:20:15: cannot use "hello" (type untyped string) as type int in assignment
```

Output:
```
build error: undefined: fmt.Printl
  ./main.go:15:2
  ./main.go:20:15
```

### Python Syntax Error

Input:
```
  File "/home/user/projects/app/script.py", line 15
    if x == 5
            ^
SyntaxError: invalid syntax
```

Output:
```
SyntaxError: invalid syntax
  File "script.py", line 15
```

## Options

```
-format string
    Error format: auto, javascript, python, java, go, rust
    Default: auto (detect automatically)

-v  Verbose output with structured fields

-version
    Print version

-help
    Print help
```

## Building

```bash
go build
```

## Testing

```bash
go test -v
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

MIT
