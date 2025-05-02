package telegram

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/morph/internal/botservice"
)

type Telegram struct{}

var baseURL string

func init() {
	baseURL = "https://api.telegram.org/bot" + os.Getenv("MORPH_TELEGRAM_BOT_TOKEN")
}

func (t Telegram) GetChatID() (int64, error) {
	chatIDStr := os.Getenv("MORPH_TELEGRAM_CHAT_ID")
	chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil {
		return 0, err
	}
	return chatID, nil
}

// Parse an incoming update
func (t Telegram) Parse(body io.ReadCloser) *botservice.BotMessage {
	var update Update
	if err := json.NewDecoder(body).Decode(&update); err != nil {
		log.Printf("[Parse] Could not decode incoming update %s", err.Error())
		return nil
	}

	// Check if the message is valid
	if update.Message == nil {
		log.Printf("[User] Invalid update: %d", update.ID)
		return nil
	}

	var telegramUser = update.Message.From

	// Check if the user is valid
	if telegramUser == nil || telegramUser.ID == 0 {
		return nil
	}

	var input = update.Message.Text

	// Check if the input is valid
	if input == "" {
		return nil
	}

	// Create a user from the telegram user
	message := botservice.BotMessage{
		MessageID: update.Message.ID,
		UserID:    telegramUser.StringID(),
		ChatID:    update.Message.Chat.ID,
		Text:      input,
	}

	return &message
}

func (t Telegram) SendMessage(chatID int64, text string, replyToMessageID *int64) {
	var url string = baseURL + "/sendMessage"

	message := SendMessageRequest{
		ChatID:           chatID,
		Text:             text,
		ReplyToMessageID: replyToMessageID,
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		log.Printf("[SendMessage] JSON parsing error: %s", err)
		return
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("[SendMessage] POST error: %s", err)
		return
	}
	defer resp.Body.Close()
}
