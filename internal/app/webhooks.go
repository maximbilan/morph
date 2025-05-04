package app

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/morph/internal/category"
	"github.com/morph/internal/taskservice"
	"github.com/morph/third_party/mono"
)

func MonoWebHook(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling Mono WebHook...")

	if r.Method == http.MethodGet {
		log.Println("Mono WebHook is working")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
		return
	}

	payload, err := mono.ParseWebhookRequest(r)
	if err != nil {
		log.Printf("[Mono] Error parsing webhook: %s", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Could not parse data"))
		return
	}

	if payload == nil {
		log.Printf("[Mono] No payload to process")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
		return
	}

	mmcCategory, err := category.GetCategoryFromMCC(payload.Data.StatementItem.MCC)
	if err != nil {
		log.Printf("[Mono] Error getting category: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Could not get category"))
		return
	}

	chatID, err := bot.GetChatID()
	if err != nil {
		log.Printf("[Mono] Error getting chat ID: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Could not get chat ID"))
		return
	}

	scheduledTransaction := taskservice.ScheduledTransaction{
		ChatID:      chatID,
		MCC:         payload.Data.StatementItem.MCC,
		Category:    mmcCategory,
		Description: payload.Data.StatementItem.Description,
		Amount:      payload.Data.StatementItem.AmountFloat(),
	}

	ctx := context.Background()
	taskService.ScheduleTransaction(&ctx, scheduledTransaction, time.Now())

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))

	log.Printf("[Mono] Scheduled transaction successfully")
}
