package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BBleae/whisper-lrc/internal/input"
	"github.com/BBleae/whisper-lrc/internal/output"
	"github.com/BBleae/whisper-lrc/internal/progress"
	"github.com/BBleae/whisper-lrc/internal/whisper"
	"github.com/spf13/cobra"
)

var (
	outputFormat string
	outputDir    string
	language     string
	apiKey       string
	prompt       string
	useYtDlp     bool
	verbose      bool
)

var rootCmd = &cobra.Command{
	Use:   "whisper-lrc [files or URLs...]",
	Short: "Extract lyrics from audio files using OpenAI Whisper",
	Long: `whisper-lrc is a CLI tool that extracts lyrics from songs using OpenAI's Whisper API.

Supported inputs:
  - Local audio files (mp3, wav, m4a, flac, ogg, webm)
  - Direct URLs to audio files
  - YouTube URLs (requires yt-dlp)

Supported output formats:
  - LRC (synchronized lyrics format)
  - SRT (subtitle format)

Examples:
  whisper-lrc song.mp3
  whisper-lrc song1.mp3 song2.mp3 -f srt
  whisper-lrc https://example.com/song.mp3
  whisper-lrc --yt-dlp "https://youtube.com/watch?v=..."
  whisper-lrc *.mp3 -o ./lyrics -f lrc`,
	Args: cobra.MinimumNArgs(1),
	RunE: runExtract,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVarP(&outputFormat, "format", "f", "lrc", "Output format: lrc or srt")
	rootCmd.Flags().StringVarP(&outputDir, "output", "o", "", "Output directory (default: same as input)")
	rootCmd.Flags().StringVarP(&language, "language", "l", "", "Language code (e.g., en, zh, ja). Auto-detect if not specified")
	rootCmd.Flags().StringVar(&apiKey, "api-key", "", "OpenAI API key (or set OPENAI_API_KEY env)")
	rootCmd.Flags().StringVarP(&prompt, "prompt", "p", "", "Custom prompt for Whisper (overrides default anti-hallucination prompt)")
	rootCmd.Flags().BoolVar(&useYtDlp, "yt-dlp", false, "Use yt-dlp for YouTube/video URLs")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
}

func runExtract(cmd *cobra.Command, args []string) error {
	// Get API key
	key := apiKey
	if key == "" {
		key = os.Getenv("OPENAI_API_KEY")
	}
	if key == "" {
		return fmt.Errorf("OpenAI API key required. Set --api-key or OPENAI_API_KEY environment variable")
	}

	// Validate output format
	outputFormat = strings.ToLower(outputFormat)
	if outputFormat != "lrc" && outputFormat != "srt" {
		return fmt.Errorf("invalid output format: %s. Use 'lrc' or 'srt'", outputFormat)
	}

	// Initialize components
	client := whisper.NewClient(key)
	inputHandler := input.NewHandler(useYtDlp)
	var formatter output.Formatter
	if outputFormat == "lrc" {
		formatter = output.NewLRCFormatter()
	} else {
		formatter = output.NewSRTFormatter()
	}

	// Create progress tracker
	tracker := progress.NewTracker(len(args))
	tracker.Start()
	defer tracker.Stop()

	// Process each input
	var errors []string
	for i, arg := range args {
		tracker.SetCurrent(i+1, filepath.Base(arg))

		// Resolve input to local file
		audioPath, cleanup, err := inputHandler.Resolve(arg)
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", arg, err))
			tracker.Error(arg, err)
			continue
		}

		tracker.SetStatus("Transcribing...")
		effectivePrompt := prompt
		if effectivePrompt == "" {
			effectivePrompt = whisper.DefaultPrompt
		}
		result, err := client.Transcribe(audioPath, language, effectivePrompt)
		if err != nil {
			if cleanup != nil {
				cleanup()
			}
			errors = append(errors, fmt.Sprintf("%s: %v", arg, err))
			tracker.Error(arg, err)
			continue
		}

		// Format output
		content := formatter.Format(result)

		// Determine output path
		outPath := getOutputPath(arg, outputDir, outputFormat)

		// Write output file
		if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
			if cleanup != nil {
				cleanup()
			}
			errors = append(errors, fmt.Sprintf("%s: %v", arg, err))
			tracker.Error(arg, err)
			continue
		}

		if err := os.WriteFile(outPath, []byte(content), 0644); err != nil {
			if cleanup != nil {
				cleanup()
			}
			errors = append(errors, fmt.Sprintf("%s: %v", arg, err))
			tracker.Error(arg, err)
			continue
		}

		// Cleanup temp files
		if cleanup != nil {
			cleanup()
		}

		tracker.Complete(arg, outPath)
	}

	tracker.Stop()

	// Print summary
	fmt.Println()
	if len(errors) > 0 {
		fmt.Printf("Completed with %d error(s):\n", len(errors))
		for _, e := range errors {
			fmt.Printf("  - %s\n", e)
		}
		return fmt.Errorf("some files failed to process")
	}

	fmt.Printf("Successfully processed %d file(s)\n", len(args))
	return nil
}

func getOutputPath(input, outputDir, format string) string {
	// Get base name without extension
	base := filepath.Base(input)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)

	// Handle URLs
	if strings.HasPrefix(input, "http://") || strings.HasPrefix(input, "https://") {
		// Use a sanitized version of URL as filename
		name = sanitizeFilename(name)
		if name == "" {
			name = "output"
		}
	}

	// Determine output directory
	dir := outputDir
	if dir == "" {
		if strings.HasPrefix(input, "http://") || strings.HasPrefix(input, "https://") {
			dir = "."
		} else {
			dir = filepath.Dir(input)
		}
	}

	return filepath.Join(dir, name+"."+format)
}

func sanitizeFilename(name string) string {
	// Remove invalid characters for filenames
	invalid := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	result := name
	for _, char := range invalid {
		result = strings.ReplaceAll(result, char, "_")
	}
	return result
}
