package botservice

import "io"

type BotMessage struct {
	MessageID int64
	UserID    string
	ChatID    int64
	Text      string
}

type BotService interface {
	GetChatID() (int64, error)
	Parse(body io.ReadCloser) *BotMessage
	SendMessage(chatID int64, text string, replyToMessageID *int64)
}
