package taskservice

type ScheduledMessage struct {
	ChatID int64  `json:"chatId"`
	Text   string `json:"text"`
}
