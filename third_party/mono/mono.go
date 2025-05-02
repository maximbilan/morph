package mono

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
)

// StatementItem represents a single transaction in Mono
type StatementItem struct {
	ID              string  `json:"id"`
	Time            int64   `json:"time"`
	Description     string  `json:"description"`
	MCC             string  `json:"mcc"`
	OriginalMCC     int     `json:"originalMcc"`
	Amount          int     `json:"amount"`
	OperationAmount int     `json:"operationAmount"`
	CurrencyCode    int     `json:"currencyCode"`
	CommissionRate  float64 `json:"commissionRate"`
	CashbackAmount  int     `json:"cashbackAmount"`
	Balance         int     `json:"balance"`
	Hold            bool    `json:"hold"`
	ReceiptID       string  `json:"receiptId"`
}

// StatementData represents the data field in the webhook payload
type StatementData struct {
	Account       string        `json:"account"`
	StatementItem StatementItem `json:"statementItem"`
}

// WebhookPayload represents the complete Mono webhook payload
type WebhookPayload struct {
	Type string        `json:"type"`
	Data StatementData `json:"data"`
}

// ParseWebhookRequest parses the webhook request and returns the payload
func ParseWebhookRequest(r *http.Request) (*WebhookPayload, error) {
	log.Printf("[Mono] Received request: %s", r.URL.Path)
	log.Printf("[Mono] Headers: %v", r.Header)
	log.Printf("[Mono] Method: %s", r.Method)
	log.Printf("[Mono] RemoteAddr: %s", r.RemoteAddr)
	log.Printf("[Mono] Content-Type: %s", r.Header.Get("Content-Type"))
	log.Printf("[Mono] User-Agent: %s", r.Header.Get("User-Agent"))

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("[Mono] Error reading body: %s", err.Error())
		return nil, err
	}

	log.Printf("[Mono] Received message: %s", string(body))

	var payload WebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		log.Printf("[Mono] Error parsing webhook: %s", err.Error())
		return nil, err
	}

	return &payload, nil
}
