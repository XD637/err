# Contributing

## Development

Clone and build:

```bash
git clone https://github.com/XD637/err.git
cd err
go build
```

Run tests:

```bash
go test -v
go test -race
go test -cover
```

## Code style

- Follow standard Go conventions
- Run `go fmt` before committing
- Keep functions focused and testable
- Add tests for new features
- Update README with examples

## Adding language support

To add support for a new error format:

1. Add detection logic in `detectFormat()` in `cleaner.go`
2. Create parser function in `parsers.go` (e.g., `cleanRuby()`)
3. Add test cases in `cleaner_test.go`
4. Update README with examples

Example parser structure:

```go
func cleanNewLang(text string) *CleanedError {
    lines := strings.Split(text, "\n")
    result := &CleanedError{}
    
    // Parse error type and message
    // Extract stack frames
    // Apply stripNoise()
    // Deduplicate frames
    
    return result
}
```

## Pull requests

- One feature per PR
- Include tests
- Update documentation
- Keep commits focused

## Issues

Report bugs with:

- Input that caused the issue
- Expected output
- Actual output
- Version (`err -version`)

## Release process

Maintainers only:

1. Update version in `main.go`
2. Update `CHANGELOG.md`
3. Tag release: `git tag v0.x.0`
4. Push: `git push origin v0.x.0`
5. CI builds and publishes binaries
