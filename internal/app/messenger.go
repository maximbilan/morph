package app

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/morph/internal/taskservice"
)

func SendMessage(w http.ResponseWriter, r *http.Request) {
	var msg taskservice.ScheduledMessage
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		log.Printf("[Morph] Could not parse message %s", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Could not parse message"))
		return
	}

	bot.SendMessage(msg.ChatID, msg.Text, msg.ReplyToMessageID)
	log.Printf("[Scheduler] Message sent to user: %d", msg.ChatID)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
