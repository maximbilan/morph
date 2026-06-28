package app

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/morph/internal/category"
	"github.com/morph/internal/taskservice"
)

// notificationRequest is the payload forwarded by the iOS Shortcut that parses a
// bank push notification into its source app name, title and body. Date is
// optional and, when present, sets the transaction date on the deep link.
type notificationRequest struct {
	App     string `json:"app"`
	Title   string `json:"title"`
	Message string `json:"message"`
	Date    string `json:"date"`
}

// parseNotificationDate interprets the optional date forwarded by the Shortcut.
// It accepts an absolute instant (RFC3339 timestamp carrying its own offset, or a
// Unix epoch in seconds or milliseconds) or a naive datetime copied from the
// notification text. Naive values have no timezone, so they are read in the same
// zone the MoneyWiz deep link renders in. It falls back to the current time when
// the value is absent or unrecognized.
func parseNotificationDate(raw string) time.Time {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Now()
	}

	if t, err := time.Parse(time.RFC3339, raw); err == nil {
		return t
	}

	if epoch, err := strconv.ParseInt(raw, 10, 64); err == nil {
		if epoch > 1e12 {
			return time.Unix(epoch/1000, 0)
		}
		return time.Unix(epoch, 0)
	}

	loc, err := time.LoadLocation("Europe/Kyiv")
	if err != nil {
		loc = time.FixedZone("EET", 2*60*60)
	}
	naiveLayouts := []string{
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"02-01-2006 15:04:05",
		"02-01-2006 15:04",
	}
	for _, layout := range naiveLayouts {
		if t, err := time.ParseInLocation(layout, raw, loc); err == nil {
			return t
		}
	}

	log.Printf("[Morph] Could not parse notification date %q, using current time", raw)
	return time.Now()
}

// appAccountMap maps a notification source app to the MoneyWiz account name used
// in deep links. The full app:account mapping will be provided later; until then
// unknown apps fall back to the app name itself.
var appAccountMap = map[string]string{}

// getAccountNameFromApp resolves the MoneyWiz account name for a notification app.
func getAccountNameFromApp(appName string) string {
	if accountName, ok := appAccountMap[appName]; ok {
		return accountName
	}

	log.Printf("[Morph] Unknown notification app: %q, using app name as account", appName)
	return appName
}

// NotificationHandler turns a parsed bank push notification (app, title, message)
// into a categorized MoneyWiz deep link and delivers it to Telegram.
func NotificationHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("[Morph] Started notification handling...")

	var notification notificationRequest
	if err := json.NewDecoder(r.Body).Decode(&notification); err != nil {
		log.Printf("[Morph] Could not parse notification %s", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Could not parse notification"))
		return
	}

	if notification.Title == "" && notification.Message == "" {
		log.Printf("[Morph] No content in notification")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
		return
	}

	// The notification is forwarded by a Shortcut and carries no chat ID, so
	// resolve the target chat from configuration.
	chatID, err := bot.GetChatID()
	if err != nil {
		log.Printf("[Morph] Error getting chat ID: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Could not get chat ID"))
		return
	}

	ctx := context.Background()
	taskService.Connect(&ctx)
	defer taskService.Close()

	categories := category.GetCategoriesInJSON()
	hints := category.GetHintsInJSON()

	systemPrompt := "You are a data analyst. Your task is to classify a bank push notification into a category, subcategory, and amount. You MUST ONLY use the categories and subcategories provided below—do not invent new ones. If the input does not match any, use 'Other' for category and an empty string for subcategory. Extract the transaction amount from the notification text as a number. Output a single-line JSON object with only these fields: category, subcategory, amount. Example of the output: {\"category\": \"Children\", \"subcategory\": \"Vocal\", \"amount\": 400.0}. Categories and subcategories: " + categories + " Hints: " + hints + " IMPORTANT: Do not add any explanation or extra text. Only output the JSON object."
	userPrompt := fmt.Sprintf("Classify this bank push notification.\nApp: %s\nTitle: %s\nMessage: %s", notification.App, notification.Title, notification.Message)

	response := aiService.Request("Morph", "Translates a bank push notification into: Category, Subcategory, Amount", systemPrompt, userPrompt, &ctx)
	if response == nil {
		log.Printf("[Morph] No response from AI")
		scheduledMessage := taskservice.ScheduledMessage{
			ChatID:           chatID,
			Text:             "No response from AI",
			ReplyToMessageID: nil,
		}
		taskService.ScheduleMessage(&ctx, scheduledMessage, time.Now())
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
		return
	}

	absoluteAmount := math.Abs(response.Amount)
	accountName := getAccountNameFromApp(notification.App)

	txTime := parseNotificationDate(notification.Date)

	log.Printf("[Morph] Response: %s %s %f (account: %s, date: %s)", response.Category, response.Subcategory, absoluteAmount, accountName, txTime)
	text := fmt.Sprintf("📲 %s\nCategory: %s\nSubcategory: %s\nAmount: %.2f", notification.App, response.Category, response.Subcategory, absoluteAmount)

	deepLink := deepLinkGenerator.Create(response.Category, response.Subcategory, accountName, absoluteAmount, txTime)

	url, err := shortURLService.Shorten(deepLink)
	if err != nil {
		log.Printf("[Morph] Error shortening URL: %v", err)
		text += "\nError shortening URL: " + err.Error()
	} else {
		log.Printf("[Morph] Shortened URL: %s", url)
		text += "\n" + url
	}

	log.Printf("[Morph] Sending message to chat %d", chatID)

	scheduledMessage := taskservice.ScheduledMessage{
		ChatID:           chatID,
		Text:             text,
		ReplyToMessageID: nil,
	}
	taskService.ScheduleMessage(&ctx, scheduledMessage, time.Now())

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
	log.Println("[Morph] Notification handler finished")
}
