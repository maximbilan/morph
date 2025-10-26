package shorturl

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/morph/third_party/bitly"
	"github.com/morph/third_party/shortio"
)

// FallbackService implements ShortURL interface and tries multiple services in order
type FallbackService struct {
	services []ShortURL
}

// NewFallbackService creates a new fallback service with the given services
func NewFallbackService(services ...ShortURL) *FallbackService {
	return &FallbackService{
		services: services,
	}
}

// Shorten tries each service in order until one succeeds
func (f *FallbackService) Shorten(URL string) (string, error) {
	var lastErr error
	
	for i, service := range f.services {
		log.Printf("[FallbackService] Trying service %d", i+1)
		
		shortURL, err := service.Shorten(URL)
		if err != nil {
			log.Printf("[FallbackService] Service %d failed: %v", i+1, err)
			lastErr = err
			
			// Check if this is a ShortIO limit error that should trigger fallback
			if f.shouldFallback(err) {
				log.Printf("[FallbackService] Service %d hit limit, trying next service", i+1)
				continue
			}
			
			// For other errors, we might want to continue or stop based on the error type
			// For now, we'll continue to the next service for any error
			continue
		}
		
		log.Printf("[FallbackService] Service %d succeeded: %s", i+1, shortURL)
		return shortURL, nil
	}
	
	return "", fmt.Errorf("all services failed, last error: %v", lastErr)
}

// shouldFallback determines if an error should trigger fallback to the next service
func (f *FallbackService) shouldFallback(err error) bool {
	if err == nil {
		return false
	}
	
	errorMsg := err.Error()
	
	// Check for ShortIO limit error message
	if strings.Contains(errorMsg, "You are out of your account link or domain limit") {
		return true
	}
	
	// Try to parse the error message as JSON to check for statusCode 402
	if f.isStatusCode402(errorMsg) {
		return true
	}
	
	// Add more conditions as needed for other services
	return false
}

// isStatusCode402 checks if the error message contains a JSON response with statusCode 402
func (f *FallbackService) isStatusCode402(errorMsg string) bool {
	// Look for JSON structure in the error message
	// The error format is: "failed to shorten URL: {"message":"...","success":false,"statusCode":402}"
	
	// Find the JSON part after "failed to shorten URL: "
	jsonStart := strings.Index(errorMsg, "failed to shorten URL: ")
	if jsonStart == -1 {
		return false
	}
	
	jsonPart := errorMsg[jsonStart+len("failed to shorten URL: "):]
	
	// Try to parse the JSON
	var errorResponse struct {
		Message    string `json:"message"`
		Success    bool   `json:"success"`
		StatusCode int    `json:"statusCode"`
	}
	
	if err := json.Unmarshal([]byte(jsonPart), &errorResponse); err != nil {
		// If JSON parsing fails, fall back to string matching
		return strings.Contains(errorMsg, `"statusCode":402`) || 
			   strings.Contains(errorMsg, `\"statusCode\":402`)
	}
	
	// Check if statusCode is 402
	return errorResponse.StatusCode == 402
}

// AddService adds a new service to the fallback chain
func (f *FallbackService) AddService(service ShortURL) {
	f.services = append(f.services, service)
}

// GetServiceCount returns the number of services in the fallback chain
func (f *FallbackService) GetServiceCount() int {
	return len(f.services)
}

// CreateDefaultFallbackService creates a fallback service with ShortIO and Bitly
func CreateDefaultFallbackService() *FallbackService {
	shortIOService := shortio.ShortIO{}
	bitlyService := bitly.Bitly{}
	
	return NewFallbackService(&shortIOService, &bitlyService)
}