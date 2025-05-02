package mono

import (
	"encoding/json"
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

// ParseWebhook parses the webhook payload from the request body
func ParseWebhook(body []byte) (*WebhookPayload, error) {
	var payload WebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}
	return &payload, nil
}
