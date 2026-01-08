package input

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Supported audio extensions
var supportedExtensions = map[string]bool{
	".mp3":  true,
	".wav":  true,
	".m4a":  true,
	".flac": true,
	".ogg":  true,
	".webm": true,
	".mp4":  true,
}

// Handler resolves various input sources to local audio files
type Handler struct {
	useYtDlp bool
}

// NewHandler creates a new input handler
func NewHandler(useYtDlp bool) *Handler {
	return &Handler{
		useYtDlp: useYtDlp,
	}
}

// Resolve converts an input (file path, URL, etc.) to a local file path
// Returns the path and a cleanup function (nil if no cleanup needed)
func (h *Handler) Resolve(input string) (string, func(), error) {
	// Check if it's a URL
	if strings.HasPrefix(input, "http://") || strings.HasPrefix(input, "https://") {
		return h.resolveURL(input)
	}

	// Local file
	return h.resolveLocalFile(input)
}

func (h *Handler) resolveLocalFile(path string) (string, func(), error) {
	// Check if file exists
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil, fmt.Errorf("file not found: %s", path)
		}
		return "", nil, fmt.Errorf("failed to access file: %w", err)
	}

	if info.IsDir() {
		return "", nil, fmt.Errorf("path is a directory: %s", path)
	}

	// Check extension
	ext := strings.ToLower(filepath.Ext(path))
	if !supportedExtensions[ext] {
		return "", nil, fmt.Errorf("unsupported audio format: %s", ext)
	}

	return path, nil, nil
}

func (h *Handler) resolveURL(url string) (string, func(), error) {
	// Check if it's a YouTube URL and yt-dlp is enabled
	if h.useYtDlp && isYouTubeURL(url) {
		return h.downloadWithYtDlp(url)
	}

	// Direct download for regular URLs
	return h.downloadDirect(url)
}

func (h *Handler) downloadDirect(inputURL string) (string, func(), error) {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return fmt.Errorf("too many redirects")
			}
			return nil
		},
	}

	resp, err := client.Get(inputURL)
	if err != nil {
		return "", nil, fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", nil, fmt.Errorf("download failed with status: %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "text/html") {
		return "", nil, fmt.Errorf("URL returned HTML instead of audio (possibly a redirect to error page)")
	}

	ext := getAudioExtension(resp)

	tmpFile, err := os.CreateTemp("", "whisper-lrc-*"+ext)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()

	cleanup := func() {
		os.Remove(tmpPath)
	}

	_, err = io.Copy(tmpFile, resp.Body)
	tmpFile.Close()
	if err != nil {
		cleanup()
		return "", nil, fmt.Errorf("failed to save download: %w", err)
	}

	return tmpPath, cleanup, nil
}

// getAudioExtension determines the audio file extension from HTTP response
func getAudioExtension(resp *http.Response) string {
	// Map Content-Type to extension
	contentTypeMap := map[string]string{
		"audio/mpeg":               ".mp3",
		"audio/mp3":                ".mp3",
		"audio/wav":                ".wav",
		"audio/x-wav":              ".wav",
		"audio/wave":               ".wav",
		"audio/mp4":                ".m4a",
		"audio/x-m4a":              ".m4a",
		"audio/m4a":                ".m4a",
		"audio/aac":                ".m4a",
		"audio/flac":               ".flac",
		"audio/x-flac":             ".flac",
		"audio/ogg":                ".ogg",
		"audio/webm":               ".webm",
		"video/mp4":                ".mp4",
		"video/webm":               ".webm",
		"application/ogg":          ".ogg",
		"application/octet-stream": "", // Will fallback to URL check
	}

	// Try Content-Type first
	contentType := resp.Header.Get("Content-Type")
	// Strip charset or other parameters
	if idx := strings.Index(contentType, ";"); idx != -1 {
		contentType = strings.TrimSpace(contentType[:idx])
	}

	if ext, ok := contentTypeMap[contentType]; ok && ext != "" {
		return ext
	}

	// Fallback: check final URL after redirects
	finalURL := resp.Request.URL.String()
	urlPath := resp.Request.URL.Path

	// Check for extension in URL path
	for ext := range supportedExtensions {
		if strings.HasSuffix(strings.ToLower(urlPath), ext) {
			return ext
		}
		// Also check if extension appears before query params
		if strings.Contains(strings.ToLower(finalURL), ext+"?") ||
			strings.Contains(strings.ToLower(finalURL), ext+"&") {
			return ext
		}
	}

	// Default to mp3
	return ".mp3"
}

func (h *Handler) downloadWithYtDlp(url string) (string, func(), error) {
	// Check if yt-dlp is available
	if _, err := exec.LookPath("yt-dlp"); err != nil {
		return "", nil, fmt.Errorf("yt-dlp not found. Please install it: https://github.com/yt-dlp/yt-dlp")
	}

	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "whisper-lrc-ytdlp-")
	if err != nil {
		return "", nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	// Output template
	outputTemplate := filepath.Join(tmpDir, "audio.%(ext)s")

	// Run yt-dlp
	cmd := exec.Command("yt-dlp",
		"-x",                    // Extract audio
		"--audio-format", "mp3", // Convert to mp3
		"--audio-quality", "0", // Best quality
		"-o", outputTemplate, // Output path
		"--no-playlist", // Single video only
		url,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		cleanup()
		return "", nil, fmt.Errorf("yt-dlp failed: %w\nOutput: %s", err, string(output))
	}

	// Find the downloaded file
	files, err := filepath.Glob(filepath.Join(tmpDir, "audio.*"))
	if err != nil || len(files) == 0 {
		cleanup()
		return "", nil, fmt.Errorf("yt-dlp download completed but no audio file found")
	}

	return files[0], cleanup, nil
}

func isYouTubeURL(url string) bool {
	ytPatterns := []string{
		"youtube.com/watch",
		"youtu.be/",
		"youtube.com/shorts/",
		"youtube.com/live/",
		"music.youtube.com/watch",
	}

	for _, pattern := range ytPatterns {
		if strings.Contains(url, pattern) {
			return true
		}
	}
	return false
}
