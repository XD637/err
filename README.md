# err

Clean and normalize error messages from Rust, Go, Python, and JavaScript/TypeScript.

Strip timestamps, paths, memory addresses, and noise. See what matters.

## Installation

### Homebrew

```bash
brew install XD637/err/err
```

### Go

```bash
go install github.com/XD637/err@latest
```

### From Source

```bash
git clone https://github.com/XD637/err.git
cd err
go build
```

## Usage

```bash
# Pipe from commands
npm test 2>&1 | err
python script.py 2>&1 | err
go build 2>&1 | err
cargo build 2>&1 | err

# From file
err error.log

# Verbose output
err -v error.log
```

## Examples

### TypeScript Error

Before:
```
src/index.ts:42:5 - error TS2322: Type 'string' is not assignable to type 'number'.

42     const count: number = "hello";
       ~~~~~

Found 1 error in src/index.ts:42
```

After:
```
TS2322: Type 'string' is not assignable to type 'number'.
```

### Rust Error

Before:
```
error[E0382]: borrow of moved value: `s`
  --> src/main.rs:5:20
   |
3  |     let s = String::from("hello");
   |         - move occurs because `s` has type `String`
4  |     let t = s;
   |             - value moved here
5  |     println!("{}", s);
   |                    ^ value borrowed here after move
```

After:
```
E0382: borrow of moved value: `s`
  src/main.rs:5:20
```

### Go Build Error

Before:
```
# github.com/user/myapp
./main.go:15:2: undefined: fmt.Printl
./main.go:20:15: cannot use "hello" (type untyped string) as type int in assignment
./utils.go:42:9: undefined: helper
```

After:
```
build error: undefined: fmt.Printl
  ./main.go:15:2
  ./main.go:20:15
  ./utils.go:42:9
```

### Python Error

Before:
```
Traceback (most recent call last):
  File "/home/user/projects/app/main.py", line 42, in <module>
    result = process_data(None)
  File "/home/user/projects/app/utils.py", line 15, in process_data
    return data.strip()
AttributeError: 'NoneType' object has no attribute 'strip'
```

After:
```
AttributeError: 'NoneType' object has no attribute 'strip'
  File "main.py", line 42, in <module>
  File "utils.py", line 15, in process_data
```

### npm Error

Before:
```
npm ERR! code ENOENT
npm ERR! syscall open
npm ERR! path /home/user/project/package.json
npm ERR! errno -2
npm ERR! enoent ENOENT: no such file or directory, open '/home/user/project/package.json'
npm ERR! enoent This is related to npm not being able to find a file.
```

After:
```
npm ENOENT: enoent ENOENT: no such file or directory, open 'package.json'
```

## Supported Languages

- **JavaScript/Node.js/TypeScript** - Standard errors, TypeScript compile errors, npm errors, unhandled promise rejections
- **Python** - Exceptions, tracebacks, syntax errors, import errors
- **Go** - Panics, build errors, test failures, fatal errors
- **Rust** - Compile errors, panics, backtraces (focuses on errors, ignores warnings)

## What It Does

- Extracts error type and message
- Removes timestamps, memory addresses, UUIDs, hex values
- Simplifies file paths to filenames
- Filters relevant stack frames
- Deduplicates repeated frames
- Removes language-specific internals

## Options

```
-format string
    Error format: auto, javascript, python, go, rust
    Default: auto

-v  Verbose output

-version
    Print version

-help
    Print help
```

## License

MIT
