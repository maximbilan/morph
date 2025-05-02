package category

import (
	"testing"
)

func TestGetCodeAsString(t *testing.T) {
	code := int32(123456)
	result := getCodeAsString(code)
	if result != "123456" {
		t.Errorf("Expected 123456, got %s", result)
	}
}
