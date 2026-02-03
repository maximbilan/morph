package app

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/morph/third_party/mono"
)

func TestMonoWebHook_AccountIDExtraction(t *testing.T) {
	tests := []struct {
		name      string
		payload   string
		wantAccountID string
	}{
		{
			name:      "Webhook with MonobankUAH account",
			payload:   `{"type":"StatementItem","data":{"account":"a-dnHAO9ExLnboGJP_pdwA","statementItem":{"id":"5ylwUXGpDyabl0HmHg","time":1746194127,"description":"Bolt","mcc":4121,"originalMcc":4121,"amount":-12000,"operationAmount":-12000,"currencyCode":980,"commissionRate":0,"cashbackAmount":0,"balance":2964804,"hold":true,"receiptId":"P5AE-PM51-X383-7M9A"}}}`,
			wantAccountID: "a-dnHAO9ExLnboGJP_pdwA",
		},
		{
			name:      "Webhook with MonobankEUR account",
			payload:   `{"type":"StatementItem","data":{"account":"WKl9I-LztrH1ZWeafLZEzQ","statementItem":{"id":"test123","time":1746194127,"description":"Purchase","mcc":5411,"originalMcc":5411,"amount":-5000,"operationAmount":-5000,"currencyCode":978,"commissionRate":0,"cashbackAmount":0,"balance":1000000,"hold":false}}}`,
			wantAccountID: "WKl9I-LztrH1ZWeafLZEzQ",
		},
		{
			name:      "Webhook with MonoFOPUAH account",
			payload:   `{"type":"StatementItem","data":{"account":"9mnHzIA1Fkjn7kmeKiAoGg","statementItem":{"id":"fop123","time":1746194127,"description":"FOP transaction","mcc":5812,"originalMcc":5812,"amount":-10000,"operationAmount":-10000,"currencyCode":980,"commissionRate":0,"cashbackAmount":0,"balance":500000,"hold":false}}}`,
			wantAccountID: "9mnHzIA1Fkjn7kmeKiAoGg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", "/monoWebHook", bytes.NewBufferString(tt.payload))
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			// Parse the webhook payload to verify account extraction
			payload, err := mono.ParseWebhookRequest(req)
			if err != nil {
				t.Fatalf("Failed to parse webhook request: %v", err)
			}

			if payload == nil {
				t.Fatal("Parsed payload is nil")
			}

			if payload.Data.Account != tt.wantAccountID {
				t.Errorf("Account ID = %q, want %q", payload.Data.Account, tt.wantAccountID)
			}
		})
	}
}
