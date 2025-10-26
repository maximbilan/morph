package bitly

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type Bitly struct{}

type BitlyRequest struct {
	LongURL string `json:"long_url"`
}

type BitlyResponse struct {
	CreatedAt string `json:"created_at"`
	ID        string `json:"id"`
	Link      string `json:"link"`
	LongURL   string `json:"long_url"`
}

type BitlyErrorResponse struct {
	Message    string `json:"message"`
	Success    bool   `json:"success"`
	StatusCode int    `json:"statusCode"`
}

func (b Bitly) Shorten(URL string) (string, error) {
	apiURL := "https://api-ssl.bitly.com/v4/shorten"
	requestBody := BitlyRequest{
		LongURL: URL,
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
	req.Header.Set("Authorization", "Bearer "+os.Getenv("MORPH_REDIRECT_KEY_2"))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		// Try to parse error response
		var errorResp BitlyErrorResponse
		if json.Unmarshal(body, &errorResp) == nil {
			return "", fmt.Errorf("failed to shorten URL: %s", errorResp.Message)
		}
		return "", fmt.Errorf("failed to shorten URL: %s", string(body))
	}

	var response BitlyResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to decode response: %v", err)
	}

	return response.Link, nil
}