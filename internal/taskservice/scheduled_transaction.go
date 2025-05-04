package taskservice

type ScheduledTransaction struct {
	ChatID      int64   `json:"chatId"`
	MCC         int32   `json:"mcc"`
	Category    string  `json:"category"`
	Description string  `json:"description"`
	Amount      float64 `json:"amount"`
}
