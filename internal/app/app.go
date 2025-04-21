package app

import (
	"log"
	"net/http"

	"github.com/morph/third_party/telegram"
)

var bot telegram.Telegram

func Handle(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling...")

	message := bot.Parse(r.Body)
	if message == nil {
		log.Printf("[Bot] No message to process")
		return
	}
	log.Printf("[Bot] Update: %s", message.Text)

	bot.SendMessage(message.ChatID, "Hello, "+message.UserID+"! You said: "+message.Text)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
	log.Println("Handled")
}
