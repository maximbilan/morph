package taskservice

type ScheduledMessage struct {
	ChatID           int64  `json:"chatId"`
	Text             string `json:"text"`
	ReplyToMessageID *int64 `json:"reply_to_message_id,omitempty"`
}
