# err

Clean and normalize error messages.

**Copy any error. Pipe to `err`. Get clean output.**

Strips timestamps, paths, memory addresses, and noise. Shows you what matters.

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

# From file
err error.log

# Save and process
npm test 2>&1 | tee error.log | err

# Verbose output
err -v error.log

# Specific format
err -format python < traceback.txt
```

## What it does

Reads error messages and outputs clean, normalized versions by:

- Extracting error type and core message
- Removing timestamps, memory addresses, UUIDs, hex values
- Simplifying file paths to filenames
- Filtering relevant stack frames
- Deduplicating repeated frames

## Supported formats

- JavaScript/Node.js
- Python
- Java
- Go
- Rust

Auto-detects format by default.

## Examples

### JavaScript

Input:
```
2024-01-28T14:10:36.123Z TypeError: Cannot read property 'foo' of undefined
    at Object.<anonymous> (/Users/dev/projects/myapp/src/index.js:42:5)
    at Module._compile (internal/modules/cjs/loader.js:1063:30)
    at Object.Module._extensions..js (internal/modules/cjs/loader.js:1092:10)
```

Output:
```
TypeError: Cannot read property 'foo' of undefined
  at Object.<anonymous> (index.js:42:5)
  at Module._compile (loader.js:1063:30)
  at Object.Module._extensions..js (loader.js:1092:10)
```

### Python

Input:
```
Traceback (most recent call last):
  File "/home/user/projects/app/main.py", line 42, in <module>
    result = process_data(None)
  File "/home/user/projects/app/utils.py", line 15, in process_data
    return data.strip()
AttributeError: 'NoneType' object has no attribute 'strip'
```

Output:
```
AttributeError: 'NoneType' object has no attribute 'strip'
  File "main.py", line 42, in <module>
  File "utils.py", line 15, in process_data
```

### Go

Input:
```
panic: runtime error: invalid memory address or nil pointer dereference
[signal SIGSEGV: segmentation violation code=0x1 addr=0x0 pc=0x4a5f23]

goroutine 1 [running]:
main.processData(0x0, 0x0, 0x0)
	/Users/dev/go/src/app/main.go:42 +0x123
main.main()
	/Users/dev/go/src/app/main.go:10 +0x45
```

Output:
```
panic: runtime error: invalid memory address or nil pointer dereference
  main.processData main.go:42 +[HEX]
  main.main main.go:10 +[HEX]
```

## Options

```
-format string
    Error format: auto, javascript, python, java, go, rust
    Default: auto

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
