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

func (s ShortIO) Shorten(URL string) (string, error) {
	apiURL := "https://api.short.io/links"
	requestBody := ShortenRequest{
		Domain:      "morph-service.short.gy",
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
		return "", fmt.Errorf("failed to shorten URL: %s", string(body))
	}

	var response ShortenResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode response: %v", err)
	}

	return response.ShortURL, nil
}
