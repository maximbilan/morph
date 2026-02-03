package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/morph/internal/category"
	"github.com/morph/internal/taskservice"
	"github.com/morph/third_party/mono"
)

// responseWriter wraps http.ResponseWriter to intercept status codes
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	chatID     int64
	notified   bool
	ctx        *context.Context
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	if code == http.StatusInternalServerError && !rw.notified && rw.chatID != 0 && rw.ctx != nil {
		// Schedule general 500 error notification
		errorMessage := "❌ [Mono] POST 500 error: Internal server error occurred"
		scheduledMessage := taskservice.ScheduledMessage{
			ChatID:           rw.chatID,
			Text:             errorMessage,
			ReplyToMessageID: nil,
		}
		taskService.ScheduleMessage(rw.ctx, scheduledMessage, time.Now())
		rw.notified = true
		log.Printf("[Mono] Scheduled Telegram notification for 500 error")
	}
	rw.ResponseWriter.WriteHeader(code)
}

func MonoWebHook(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling Mono WebHook...")

	if r.Method == http.MethodGet {
		log.Println("Mono WebHook is working")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
		return
	}

	// Get chat ID early so we can send error notifications
	chatID, err := bot.GetChatID()
	if err != nil {
		log.Printf("[Mono] Error getting chat ID: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Could not get chat ID"))
		return
	}

	// Initialize context and task service early for error notifications
	ctx := context.Background()
	taskService.Connect(&ctx)
	defer taskService.Close()

	// Wrap response writer to catch 500 errors
	rw := &responseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
		chatID:         chatID,
		notified:       false,
		ctx:            &ctx,
	}

	payload, err := mono.ParseWebhookRequest(r)
	if err != nil {
		log.Printf("[Mono] Error parsing webhook: %s", err.Error())
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte("Could not parse data"))
		return
	}

	if payload == nil {
		log.Printf("[Mono] No payload to process")
		rw.WriteHeader(http.StatusOK)
		rw.Write([]byte("OK"))
		return
	}

	mmcCategory, err := category.GetCategoryFromMCC(payload.Data.StatementItem.MCC)
	if err != nil {
		log.Printf("[Mono] Error getting category: %v", err)
		
		// Check if it's an MCC code not found error
		errMsg := err.Error()
		if strings.Contains(errMsg, "MCC code not found") || strings.Contains(errMsg, "mcc code not found") {
			mccCode := payload.Data.StatementItem.MCC
			errorMessage := fmt.Sprintf("⚠️ MCC code not found: %d", mccCode)
			scheduledMessage := taskservice.ScheduledMessage{
				ChatID:           chatID,
				Text:             errorMessage,
				ReplyToMessageID: nil,
			}
			taskService.ScheduleMessage(&ctx, scheduledMessage, time.Now())
			rw.notified = true // Mark as notified to avoid duplicate notification
			log.Printf("[Mono] Scheduled Telegram notification for missing MCC code: %d", mccCode)
		} else {
			// For other category errors, schedule a general error notification
			errorMessage := fmt.Sprintf("❌ [Mono] Error getting category: %v", err)
			scheduledMessage := taskservice.ScheduledMessage{
				ChatID:           chatID,
				Text:             errorMessage,
				ReplyToMessageID: nil,
			}
			taskService.ScheduleMessage(&ctx, scheduledMessage, time.Now())
			rw.notified = true // Mark as notified to avoid duplicate notification
			log.Printf("[Mono] Scheduled Telegram notification for category error")
		}
		
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte("Could not get category"))
		return
	}

	scheduledTransaction := taskservice.ScheduledTransaction{
		ChatID:      chatID,
		MCC:         payload.Data.StatementItem.MCC,
		Category:    mmcCategory,
		Description: payload.Data.StatementItem.Description,
		Amount:      payload.Data.StatementItem.AmountFloat(),
		Time:        payload.Data.StatementItem.Time,
		IsRefund:    payload.Data.StatementItem.IsRefund(),
		AccountID:   payload.Data.Account,
	}

	taskService.ScheduleTransaction(&ctx, scheduledTransaction, time.Now())

	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte("OK"))

	log.Printf("[Mono] Scheduled transaction successfully")
}
