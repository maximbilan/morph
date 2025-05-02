package app

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestMonoWebHook(t *testing.T) {
	// Set up test environment
	os.Setenv("MORPH_TELEGRAM_CHAT_ID", "123456789")

	tests := []struct {
		name           string
		requestBody    string
		expectedStatus int
		expectedMsg    string
	}{
		{
			name: "Valid statement item with string MCC",
			requestBody: `{
				"type": "StatementItem",
				"data": {
					"account": "a-dnHAO9ExLnboGJP_pdwA",
					"statementItem": {
						"id": "eNDOMTqO54IC8j_fEA",
						"time": 1746204057,
						"description": "Bolt",
						"mcc": "4121",
						"originalMcc": 4121,
						"amount": -8400,
						"operationAmount": -8400,
						"currencyCode": 980,
						"commissionRate": 0,
						"cashbackAmount": 0,
						"balance": 2956404,
						"hold": true,
						"receiptId": "TP4C-5M8K-2PT4-61H0"
					}
				}
			}`,
			expectedStatus: http.StatusOK,
			expectedMsg: "Transaction Details:\n" +
				"Description: Bolt\n" +
				"MCC: 4121\n" +
				"Category: \n" +
				"Amount: -8400\n" +
				"Balance: 2956404\n" +
				"Receipt: TP4C-5M8K-2PT4-61H0",
		},
		{
			name: "Valid statement item with numeric MCC",
			requestBody: `{
				"type": "StatementItem",
				"data": {
					"account": "a-dnHAO9ExLnboGJP_pdwA",
					"statementItem": {
						"id": "eNDOMTqO54IC8j_fEA",
						"time": 1746204057,
						"description": "Bolt",
						"mcc": 4121,
						"originalMcc": 4121,
						"amount": -8400,
						"operationAmount": -8400,
						"currencyCode": 980,
						"commissionRate": 0,
						"cashbackAmount": 0,
						"balance": 2956404,
						"hold": true,
						"receiptId": "TP4C-5M8K-2PT4-61H0"
					}
				}
			}`,
			expectedStatus: http.StatusOK,
		},
		{
			name: "Invalid JSON structure",
			requestBody: `{
				"type": "StatementItem",
				"data": {
					"account": "a-dnHAO9ExLnboGJP_pdwA"
				}
			}`,
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "Invalid JSON",
			requestBody:    `{invalid json}`,
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a request with the test body
			req, err := http.NewRequest("POST", "/mono-webhook", bytes.NewBufferString(tt.requestBody))
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			// Create a ResponseRecorder to record the response
			rr := httptest.NewRecorder()

			// Call the handler
			MonoWebHook(rr, req)

			// Check the status code
			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.expectedStatus)
			}

			// For successful requests, verify the response body
			if tt.expectedStatus == http.StatusOK {
				expected := "OK"
				if rr.Body.String() != expected {
					t.Errorf("handler returned unexpected body: got %v want %v",
						rr.Body.String(), expected)
				}
			}
		})
	}
}
