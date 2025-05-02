package botservice

import "net/http"

type BotMessage struct {
	MessageID int64
	UserID    string
	ChatID    int64
	Text      string
}

type BotService interface {
	GetChatID() (int64, error)
	Parse(r *http.Request) *BotMessage
	SendMessage(chatID int64, text string, replyToMessageID *int64)
}
