package main

import (
	"flag"
	"fmt"
	"io"
	"os"
)

const version = "0.1.0"

var (
	flagVersion = flag.Bool("version", false, "print version")
	flagHelp    = flag.Bool("help", false, "print help")
	flagVerbose = flag.Bool("v", false, "verbose output")
	flagFormat  = flag.String("format", "auto", "error format (auto|javascript|python|go|rust)")
)

func main() {
	flag.Parse()

	if *flagVersion {
		fmt.Println(version)
		os.Exit(0)
	}

	if *flagHelp {
		printHelp()
		os.Exit(0)
	}

	args := flag.Args()
	var data string

	// Read from file if provided
	if len(args) > 0 {
		f, err := os.Open(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()

		bytes, err := io.ReadAll(f)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error reading file: %v\n", err)
			os.Exit(1)
		}
		data = string(bytes)
	} else {
		// Read from stdin (piped input)
		bytes, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error reading input: %v\n", err)
			os.Exit(1)
		}
		data = string(bytes)
	}

	// Process the error
	cleaner := NewCleaner(*flagFormat)
	result := cleaner.Clean(data)

	// Add separator in interactive mode
	if len(args) == 0 {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) != 0 {
			fmt.Fprintln(os.Stderr, "\n---")
		}
	}

	// Output
	if *flagVerbose {
		fmt.Printf("Type: %s\n", result.Type)
		fmt.Printf("Message: %s\n", result.Message)
		if len(result.Stack) > 0 {
			fmt.Println("\nStack:")
			for _, frame := range result.Stack {
				fmt.Printf("  %s\n", frame)
			}
		}
	} else {
		fmt.Print(result.Format())
	}
}

func printHelp() {
	fmt.Println(`err - clean and normalize error messages

USAGE
    err [OPTIONS] [FILE]

    Read error messages from FILE or stdin, strip noise, and output
    clean, normalized error information.

OPTIONS
    -format string
        Error format: auto, javascript, python, java, go, rust
        Default: auto (detect automatically)
    
    -v  Verbose output with structured fields
    
    -version
        Print version
    
    -help
        Print this help

EXAMPLES
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

OUTPUT
    Cleaned error with:
    - Type and message extracted
    - Noise removed (timestamps, addresses, paths, UUIDs)
    - Relevant stack frames
    - Duplicates removed

DOCUMENTATION
    https://github.com/XD637/err`)
}
