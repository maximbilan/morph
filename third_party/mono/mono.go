package mono

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
)

// StatementItem represents a single transaction in Mono
type StatementItem struct {
	ID              string `json:"id"`
	Time            int64  `json:"time"`
	Description     string `json:"description"`
	MCC             int32  `json:"mcc"`
	OriginalMCC     int32  `json:"originalMcc"`
	Hold            bool   `json:"hold"`
	Amount          int64  `json:"amount"`
	OperationAmount int64  `json:"operationAmount"`
	CurrencyCode    int32  `json:"currencyCode"`
	CommissionRate  int64  `json:"commissionRate"`
	CashbackAmount  int64  `json:"cashbackAmount"`
	Balance         int64  `json:"balance"`
	Comment         string `json:"comment"`
	ReceiptID       string `json:"receiptId"`
	InvoiceID       string `json:"invoiceId"`
	CounterEdrpou   string `json:"counterEdrpou"`
	CounterIban     string `json:"counterIban"`
	CounterName     string `json:"counterName"`
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
	body, err := io.ReadAll(r.Body)
	log.Printf("[Mono] Received message: %s", string(body))

	if err != nil {
		return nil, err
	}

	var payload WebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}

	return &payload, nil
}
