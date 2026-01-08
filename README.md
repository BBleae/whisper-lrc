# whisper-lrc

[![CI](https://github.com/BBleae/whisper-lrc/actions/workflows/ci.yml/badge.svg)](https://github.com/BBleae/whisper-lrc/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/BBleae/whisper-lrc)](https://github.com/BBleae/whisper-lrc/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Extract synchronized lyrics from audio files using OpenAI's Whisper API.

> **Note**: This is my first pure Vibe Coding project, created to try out [OpenCode](https://github.com/opencode-ai/opencode) and [oh-my-opencode](https://github.com/code-yeongyu/oh-my-opencode).

## Features

- **Multiple input sources**: Local files, direct URLs, YouTube (via yt-dlp)
- **Output formats**: LRC (lyrics) and SRT (subtitles)
- **Batch processing**: Process multiple files at once
- **Language support**: Auto-detection or manual specification
- **Progress display**: Real-time processing status

## Installation

### From Release (Recommended)

Download the latest binary from [Releases](https://github.com/BBleae/whisper-lrc/releases).

### Using Go

```bash
go install github.com/BBleae/whisper-lrc@latest
```

### Build from Source

```bash
git clone https://github.com/BBleae/whisper-lrc.git
cd whisper-lrc
go build -o whisper-lrc .
```

## Prerequisites

- OpenAI API key with access to the Whisper API
- (Optional) [yt-dlp](https://github.com/yt-dlp/yt-dlp) for YouTube support

## Usage

### Basic Usage

```bash
# Set your API key
export OPENAI_API_KEY="sk-..."

# Extract lyrics from a local file
whisper-lrc song.mp3

# Output will be saved as song.lrc in the same directory
```

### Output Formats

```bash
# LRC format (default)
whisper-lrc song.mp3 -f lrc

# SRT format
whisper-lrc song.mp3 -f srt
```

### Batch Processing

```bash
# Process multiple files
whisper-lrc song1.mp3 song2.mp3 song3.mp3

# Process all MP3 files in current directory
whisper-lrc *.mp3

# Save all outputs to a specific directory
whisper-lrc *.mp3 -o ./lyrics
```

### URL Support

```bash
# Direct URL to audio file
whisper-lrc https://example.com/song.mp3

# YouTube (requires yt-dlp)
whisper-lrc --yt-dlp "https://www.youtube.com/watch?v=VIDEO_ID"
```

### Language Options

```bash
# Auto-detect language (default)
whisper-lrc song.mp3

# Specify language
whisper-lrc song.mp3 -l ja    # Japanese
whisper-lrc song.mp3 -l zh    # Chinese
whisper-lrc song.mp3 -l en    # English
```

### All Options

```
Flags:
      --api-key string    OpenAI API key (or set OPENAI_API_KEY env)
  -f, --format string     Output format: lrc or srt (default "lrc")
  -h, --help              help for whisper-lrc
  -l, --language string   Language code (e.g., en, zh, ja). Auto-detect if not specified
  -o, --output string     Output directory (default: same as input)
  -v, --verbose           Verbose output
      --yt-dlp            Use yt-dlp for YouTube/video URLs
```

## Supported Audio Formats

- MP3
- WAV
- M4A
- FLAC
- OGG
- WebM
- MP4

## Output Examples

### LRC Format

```
[re:whisper-lrc]
[la:en]

[00:00.50]Never gonna give you up
[00:03.20]Never gonna let you down
[00:05.80]Never gonna run around and desert you
```

### SRT Format

```
1
00:00:00,500 --> 00:00:03,200
Never gonna give you up

2
00:00:03,200 --> 00:00:05,800
Never gonna let you down

3
00:00:05,800 --> 00:00:09,100
Never gonna run around and desert you
```

## License

[MIT](LICENSE)
