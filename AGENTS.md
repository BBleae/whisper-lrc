# AGENTS.md

Guidelines for AI coding agents working on this project.

## Project Overview

whisper-lrc is a CLI tool that extracts synchronized lyrics from audio files using OpenAI's Whisper API.

**Tech Stack**: Go 1.22+, Cobra (CLI framework)

## Project Structure

```
whisper-lrc/
├── main.go                      # Entry point
├── cmd/
│   └── root.go                  # CLI commands and flags
└── internal/
    ├── whisper/
    │   └── client.go            # OpenAI Whisper API client
    ├── input/
    │   └── handler.go           # Input handling (local/URL/yt-dlp)
    ├── output/
    │   └── formatter.go         # LRC/SRT formatters
    └── progress/
        └── tracker.go           # Progress display
```

## Development Guidelines

### Go Version

**IMPORTANT**: Use Go 1.22 in `go.mod`. Do NOT use newer versions (e.g., 1.25.x) as they are incompatible with golangci-lint in CI.

```go
// go.mod
go 1.22  // Correct
go 1.25  // WRONG - breaks CI
```

### Building

```bash
go build -o whisper-lrc .
```

### Testing

```bash
go test ./...
```

### Code Style

- Follow standard Go conventions
- Run `go fmt` before committing
- Ensure `golangci-lint` passes (CI will check this)

### Adding New Features

1. Input sources → `internal/input/handler.go`
2. Output formats → `internal/output/formatter.go`
3. CLI flags → `cmd/root.go`
4. API changes → `internal/whisper/client.go`

## CI/CD

- **CI**: Runs on every push/PR to `main` (build, test, lint)
- **Release**: Triggered by pushing a version tag (e.g., `v0.1.0`)

### Creating a Release

```bash
git tag v0.x.x
git push origin v0.x.x
```

GoReleaser will automatically build binaries for:
- Linux (amd64, arm64)
- macOS (amd64, arm64)
- Windows (amd64, arm64)

## Common Pitfalls

| Issue | Solution |
|-------|----------|
| CI lint fails with Go version error | Ensure `go.mod` uses `go 1.22`, not newer |
| Import path errors | Use `github.com/BBleae/whisper-lrc/...` |
| yt-dlp not working | Ensure `--yt-dlp` flag is passed for YouTube URLs |
