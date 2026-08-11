package taskservice

type ScheduledTransaction struct {
	ChatID int64 `json:"chatId"`
	MCC    int32 `json:"mcc"`
	// What the MCC stands for, e.g. "Automated fuel dispensers". The wire name
	// stays "category" so already-queued tasks still decode.
	MCCDescription string  `json:"category"`
	Description    string  `json:"description"`
	Amount         float64 `json:"amount"`
	Time           int64   `json:"time"`
	IsRefund       bool    `json:"isRefund"`
	AccountID      string  `json:"accountId"`
}
