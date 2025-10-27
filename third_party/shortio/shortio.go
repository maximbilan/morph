package shortio

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type ShortIO struct{}

type ShortenRequest struct {
	Domain      string `json:"domain"`
	OriginalURL string `json:"originalURL"`
}

type ShortenResponse struct {
	ShortURL string `json:"shortURL"`
}

type ErrorResponse struct {
	Message    string `json:"message"`
	Success    bool   `json:"success"`
	StatusCode int    `json:"statusCode"`
}

type ShortIOError struct {
	StatusCode int
	Message    string
}

func (e *ShortIOError) Error() string {
	return e.Message
}

func (s ShortIO) Shorten(URL string) (string, error) {
	// Try with primary domain first
	shortURL, err := s.shortenWithDomain(URL, "morph-service.short.gy")
	if err == nil {
		return shortURL, nil
	}

	// Check if it's a 402 error (domain limit exceeded)
	if s.is402Error(err) {
		// Retry with alternative domain
		shortURL, retryErr := s.shortenWithDomain(URL, "morph2")
		if retryErr == nil {
			return shortURL, nil
		}
		return "", fmt.Errorf("failed to shorten URL with both domains: primary error: %v, retry error: %v", err, retryErr)
	}

	return "", err
}

func (s ShortIO) shortenWithDomain(URL, domain string) (string, error) {
	apiURL := "https://api.short.io/links"
	requestBody := ShortenRequest{
		Domain:      domain,
		OriginalURL: URL,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %v", err)
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", os.Getenv("MORPH_REDIRECT_KEY"))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		// Create a custom error that includes the status code for better error detection
		return "", &ShortIOError{
			StatusCode: resp.StatusCode,
			Message:    fmt.Sprintf("failed to shorten URL with domain %s (status %d): %s", domain, resp.StatusCode, string(body)),
		}
	}

	var response ShortenResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode response: %v", err)
	}

	return response.ShortURL, nil
}

func (s ShortIO) is402Error(err error) bool {
	if err == nil {
		return false
	}

	// Check if it's our custom error with status code 402
	if shortIOErr, ok := err.(*ShortIOError); ok {
		return shortIOErr.StatusCode == 402
	}

	return false
}
