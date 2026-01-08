package whisper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

const apiURL = "https://api.openai.com/v1/audio/transcriptions"

const DefaultPrompt = `Transcribe only the actual sung or spoken lyrics. Do not add metadata such as composer, lyricist, arranger, artist names, song titles, or credits. If there is silence or instrumental sections, output nothing for those parts.`

// Segment represents a transcribed segment with timing
type Segment struct {
	Start float64 `json:"start"`
	End   float64 `json:"end"`
	Text  string  `json:"text"`
}

// TranscriptionResult holds the complete transcription
type TranscriptionResult struct {
	Text     string    `json:"text"`
	Language string    `json:"language"`
	Segments []Segment `json:"segments"`
}

// Client handles OpenAI Whisper API communication
type Client struct {
	apiKey     string
	httpClient *http.Client
}

// NewClient creates a new Whisper API client
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey:     apiKey,
		httpClient: &http.Client{},
	}
}

// Transcribe sends an audio file to Whisper API and returns the result
func (c *Client) Transcribe(audioPath string, language string, prompt string) (*TranscriptionResult, error) {
	file, err := os.Open(audioPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open audio file: %w", err)
	}
	defer file.Close()

	// Create multipart form
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Add file
	part, err := writer.CreateFormFile("file", filepath.Base(audioPath))
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}
	if _, err := io.Copy(part, file); err != nil {
		return nil, fmt.Errorf("failed to copy file: %w", err)
	}

	// Add model
	if err := writer.WriteField("model", "whisper-1"); err != nil {
		return nil, fmt.Errorf("failed to write model field: %w", err)
	}

	// Add response format for timestamps
	if err := writer.WriteField("response_format", "verbose_json"); err != nil {
		return nil, fmt.Errorf("failed to write response_format field: %w", err)
	}

	// Add timestamp granularities
	if err := writer.WriteField("timestamp_granularities[]", "segment"); err != nil {
		return nil, fmt.Errorf("failed to write timestamp_granularities field: %w", err)
	}

	if language != "" {
		if err := writer.WriteField("language", language); err != nil {
			return nil, fmt.Errorf("failed to write language field: %w", err)
		}
	}

	if prompt != "" {
		if err := writer.WriteField("prompt", prompt); err != nil {
			return nil, fmt.Errorf("failed to write prompt field: %w", err)
		}
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close writer: %w", err)
	}

	// Create request
	req, err := http.NewRequest("POST", apiURL, &buf)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	var result TranscriptionResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}
