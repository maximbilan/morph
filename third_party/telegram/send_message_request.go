package telegram

type SendMessageRequest struct {
	ChatID           int64  `json:"chat_id"`
	Text             string `json:"text"`
	ReplyToMessageID *int64 `json:"reply_to_message_id,omitempty"`
}
