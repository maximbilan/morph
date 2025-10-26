package shorturl

import (
	"errors"
	"testing"
)

// Mock service for testing
type mockService struct {
	shouldFail bool
	errorMsg   string
	shortURL   string
}

func (m *mockService) Shorten(URL string) (string, error) {
	if m.shouldFail {
		return "", errors.New(m.errorMsg)
	}
	return m.shortURL, nil
}

func TestFallbackService_Shorten_Success(t *testing.T) {
	// Test successful shortening with first service
	service1 := &mockService{shouldFail: false, shortURL: "https://short.io/abc123"}
	service2 := &mockService{shouldFail: false, shortURL: "https://bit.ly/def456"}
	
	fallback := NewFallbackService(service1, service2)
	
	result, err := fallback.Shorten("https://example.com")
	
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	if result != "https://short.io/abc123" {
		t.Errorf("Expected https://short.io/abc123, got %s", result)
	}
}

func TestFallbackService_Shorten_Fallback(t *testing.T) {
	// Test fallback to second service when first fails with limit error
	service1 := &mockService{
		shouldFail: true, 
		errorMsg:   "failed to shorten URL: {\"message\":\"You are out of your account link or domain limit. Upgrade your account to add more links\",\"success\":false,\"statusCode\":402}",
	}
	service2 := &mockService{shouldFail: false, shortURL: "https://bit.ly/def456"}
	
	fallback := NewFallbackService(service1, service2)
	
	result, err := fallback.Shorten("https://example.com")
	
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	if result != "https://bit.ly/def456" {
		t.Errorf("Expected https://bit.ly/def456, got %s", result)
	}
}

func TestFallbackService_Shorten_AllFail(t *testing.T) {
	// Test when all services fail
	service1 := &mockService{shouldFail: true, errorMsg: "Service 1 failed"}
	service2 := &mockService{shouldFail: true, errorMsg: "Service 2 failed"}
	
	fallback := NewFallbackService(service1, service2)
	
	_, err := fallback.Shorten("https://example.com")
	
	if err == nil {
		t.Error("Expected error when all services fail")
	}
	
	if !contains(err.Error(), "all services failed") {
		t.Errorf("Expected 'all services failed' in error message, got %s", err.Error())
	}
}

func TestShouldFallback(t *testing.T) {
	fallback := &FallbackService{}
	
	// Test ShortIO limit error with full JSON
	err1 := errors.New("failed to shorten URL: {\"message\":\"You are out of your account link or domain limit. Upgrade your account to add more links\",\"success\":false,\"statusCode\":402}")
	if !fallback.shouldFallback(err1) {
		t.Error("Expected shouldFallback to return true for ShortIO limit error")
	}
	
	// Test HTTP 402 status with minimal JSON
	err2 := errors.New("failed to shorten URL: {\"statusCode\":402}")
	if !fallback.shouldFallback(err2) {
		t.Error("Expected shouldFallback to return true for HTTP 402 error")
	}
	
	// Test message-based fallback (without JSON)
	err3 := errors.New("failed to shorten URL: You are out of your account link or domain limit")
	if !fallback.shouldFallback(err3) {
		t.Error("Expected shouldFallback to return true for message-based limit error")
	}
	
	// Test other error
	err4 := errors.New("network error")
	if fallback.shouldFallback(err4) {
		t.Error("Expected shouldFallback to return false for network error")
	}
	
	// Test non-402 JSON error
	err5 := errors.New("failed to shorten URL: {\"message\":\"Invalid URL\",\"success\":false,\"statusCode\":400}")
	if fallback.shouldFallback(err5) {
		t.Error("Expected shouldFallback to return false for non-402 JSON error")
	}
}

func TestAddService(t *testing.T) {
	fallback := NewFallbackService()
	
	if fallback.GetServiceCount() != 0 {
		t.Errorf("Expected 0 services, got %d", fallback.GetServiceCount())
	}
	
	service1 := &mockService{shouldFail: false, shortURL: "https://test1.com"}
	service2 := &mockService{shouldFail: false, shortURL: "https://test2.com"}
	
	fallback.AddService(service1)
	fallback.AddService(service2)
	
	if fallback.GetServiceCount() != 2 {
		t.Errorf("Expected 2 services, got %d", fallback.GetServiceCount())
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr || 
		   len(s) > len(substr) && contains(s[1:], substr)
}