package app

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/morph/internal/category"
	"github.com/morph/internal/taskservice"
)

// notificationRequest is the bank push notification forwarded by the iOS Shortcut.
type notificationRequest struct {
	App     string `json:"app"`
	Title   string `json:"title"`
	Message string `json:"message"`
	Date    string `json:"date"`
}

// parseNotificationDate parses an RFC3339 instant, a Unix epoch (seconds or
// milliseconds), or a naive datetime (read as Kyiv time). Falls back to now.
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

// bbvaAccount is the single MoneyWiz account for every BBVA notification.
const bbvaAccount = "BBVAEur"

// pumbAccounts maps the masked account token in a PUMB notification (e.g. "*0451").
var pumbAccounts = map[string]string{
	"*0451": "PumbUAHPlatinum",
	"*5353": "PumbUAHPlatinum",
	"*2164": "PumbUAHVirtual",
	"*8043": "PumbUAHVirtual",
	"*5381": "PumbUSD",
	"*2985": "PumbUSD",
	"*0404": "PumbEUR",
	"*4249": "PumbEUR",
}

// privatAccounts maps the masked account token in a Privat24 notification (e.g. "5*85").
var privatAccounts = map[string]string{
	"5*85": "PrivatOnlineUAH",
	"5*87": "PrivatOnlineUSD",
	"4*62": "PrivatOnlineEUR",
	"5*30": "StartupPrivatUAH",
	"5*88": "StartupPrivatUSD",
	"5*64": "PrivatEntrepreneurUAH",
	"5*31": "PrivatPaymentsUAH",
	"4*99": "PrivatUniversalUAH",
	"5*07": "PrivatPaymentsUSD",
	"5*89": "PrivatEUR",
}

// resolveAccountName picks the MoneyWiz account from the app and the masked
// account in the message, falling back to the app name when unrecognized.
func resolveAccountName(app string, message string) string {
	switch detectBank(app) {
	case "bbva":
		return bbvaAccount
	case "pumb":
		if account := matchAccountToken(message, pumbAccounts); account != "" {
			return account
		}
	case "privat":
		if account := matchAccountToken(message, privatAccounts); account != "" {
			return account
		}
	}

	log.Printf("[Morph] Could not resolve account for app %q, using app name", app)
	return app
}

// detectBank maps an app name to a known bank identifier.
func detectBank(app string) string {
	normalized := strings.ToLower(strings.TrimSpace(app))
	switch {
	case strings.Contains(normalized, "bbva"):
		return "bbva"
	case strings.Contains(normalized, "pumb") || strings.Contains(normalized, "пумб"):
		return "pumb"
	case strings.Contains(normalized, "privat"):
		return "privat"
	default:
		return ""
	}
}

// matchAccountToken returns the account whose masked token appears in the message.
// The "*" in tokens avoids collisions with amounts, balances or dates.
func matchAccountToken(message string, accounts map[string]string) string {
	for token, account := range accounts {
		if strings.Contains(message, token) {
			return account
		}
	}
	return ""
}

// NotificationHandler turns a bank push notification into a MoneyWiz deep link
// and delivers it to Telegram.
func NotificationHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("[Morph] Started notification handling...")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("[Morph] Could not read notification body %s", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Could not read notification"))
		return
	}

	// Log the raw payload so we can see exactly what the iOS Shortcut sends.
	log.Printf("[Morph] Raw notification body (%d bytes): %s", len(body), string(body))

	var notification notificationRequest
	if err := json.Unmarshal(body, &notification); err != nil {
		log.Printf("[Morph] Could not parse notification %s", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Could not parse notification"))
		return
	}

	log.Printf("[Morph] Parsed notification: app=%q title=%q message=%q date=%q", notification.App, notification.Title, notification.Message, notification.Date)

	if notification.Title == "" && notification.Message == "" {
		log.Printf("[Morph] No content in notification")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
		return
	}

	// The notification carries no chat ID, so resolve it from configuration.
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

	systemPrompt := "You are a data analyst. Your task is to analyze a bank push notification and classify it into a category, subcategory, and amount. First decide whether the notification represents an actual financial transaction (a debit or credit on an account): set isTransaction to false for anything that is not a transaction, such as promotional or marketing messages, security or login alerts, or general informational messages, and set it to true only for real transactions. You MUST ONLY use the categories and subcategories provided below—do not invent new ones. If the input does not match any, use 'Other' for category and an empty string for subcategory. Extract the transaction amount from the notification text as a number (use 0 when there is no transaction). Output a single-line JSON object with only these fields: category, subcategory, amount, isTransaction. Example of the output: {\"category\": \"Children\", \"subcategory\": \"Vocal\", \"amount\": 400.0, \"isTransaction\": true}. Categories and subcategories: " + categories + " Hints: " + hints + " IMPORTANT: Do not add any explanation or extra text. Only output the JSON object."
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

	// Promotional, informational and other non-transaction pushes are dropped.
	if !response.IsTransaction {
		log.Printf("[Morph] Notification is not a transaction, ignoring")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
		return
	}

	absoluteAmount := math.Abs(response.Amount)
	accountName := resolveAccountName(notification.App, notification.Message)

	txTime := parseNotificationDate(notification.Date)

	log.Printf("[Morph] Response: %s %s %f (account: %s, date: %s)", response.Category, response.Subcategory, absoluteAmount, accountName, txTime)
	text := fmt.Sprintf("📲 %s\nCategory: %s\nSubcategory: %s\nAmount: %.2f", notification.App, response.Category, response.Subcategory, absoluteAmount)

	deepLink := deepLinkGenerator.Create(response.Category, response.Subcategory, accountName, absoluteAmount, txTime)

	text = appendShortLink(text, deepLink)

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
