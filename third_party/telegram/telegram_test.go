package telegram

import (
	"bytes"
	"io"
	"testing"

	"github.com/morph/internal/botservice"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *botservice.BotMessage
	}{
		{
			name: "valid message",
			input: `{
				"update_id": 123,
				"message": {
					"message_id": 456,
					"from": {
						"id": 789,
						"username": "testuser"
					},
					"chat": {
						"id": 101112
					},
					"text": "hello world"
				}
			}`,
			expected: &botservice.BotMessage{
				MessageID: 456,
				UserID:    "789",
				ChatID:    101112,
				Text:      "hello world",
			},
		},
		{
			name:     "invalid json",
			input:    `{invalid json}`,
			expected: nil,
		},
		{
			name:     "empty message",
			input:    `{"update_id": 123}`,
			expected: nil,
		},
		{
			name: "no user",
			input: `{
				"update_id": 123,
				"message": {
					"message_id": 456,
					"chat": {"id": 101112},
					"text": "hello"
				}
			}`,
			expected: nil,
		},
		{
			name: "empty text",
			input: `{
				"update_id": 123,
				"message": {
					"message_id": 456,
					"from": {"id": 789},
					"chat": {"id": 101112}
				}
			}`,
			expected: nil,
		},
	}

	telegram := Telegram{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := io.NopCloser(bytes.NewBufferString(tt.input))
			result := telegram.Parse(reader)

			if tt.expected == nil && result != nil {
				t.Errorf("expected nil, got %v", result)
				return
			}
			if tt.expected != nil && result == nil {
				t.Errorf("expected %v, got nil", tt.expected)
				return
			}
			if tt.expected != nil {
				if tt.expected.MessageID != result.MessageID ||
					tt.expected.UserID != result.UserID ||
					tt.expected.ChatID != result.ChatID ||
					tt.expected.Text != result.Text {
					t.Errorf("expected %v, got %v", tt.expected, result)
				}
			}
		})
	}
}
