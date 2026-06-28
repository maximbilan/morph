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

// bbvaAccount is the single MoneyWiz account used for every BBVA notification,
// regardless of the account number shown in the message.
const bbvaAccount = "BBVAEur"

// pumbAccounts maps the masked account number as it appears in a PUMB notification
// (e.g. "Рахунок: *0451") to the MoneyWiz account name.
var pumbAccounts = map[string]string{
	"*0451": "PumbUAHPlatinum",
	"*2164": "PumbUAHVirtual",
	"*5381": "PumbUSD",
	"*0404": "PumbEUR",
}

// privatAccounts maps the masked account token as it appears in a Privat24
// notification (e.g. "5*85") to the MoneyWiz account name.
var privatAccounts = map[string]string{
	"5*85": "PrivatOnlineUAH",
	"3*87": "PrivatOnlineUSD",
	"2*62": "PrivatOnlineEUR",
	"5*30": "StartupPrivatUAH",
	"9*88": "StartupPrivatUSD",
	"6*64": "PrivatEntrepreneurUAH",
	"6*31": "PrivatPaymentsUAH",
	"6*99": "PrivatUniversalUAH",
	"0*07": "PrivatPaymentsUSD",
	"1*89": "PrivatEUR",
}

// resolveAccountName determines the MoneyWiz account for a notification. The bank
// is identified from the app name; PUMB and Privat24 then resolve the specific
// account from the masked account number embedded in the message, while BBVA
// always maps to a single account. Unrecognized apps/accounts fall back to the
// app name so the transaction is still delivered (for manual account selection).
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

// detectBank maps a notification source app name to a known bank identifier.
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

// matchAccountToken returns the account name whose masked token appears in the
// message. Tokens carry a "*" (e.g. "*0451", "5*85"), which keeps the match from
// colliding with amounts, balances or dates in the notification text.
func matchAccountToken(message string, accounts map[string]string) string {
	for token, account := range accounts {
		if strings.Contains(message, token) {
			return account
		}
	}
	return ""
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
	accountName := resolveAccountName(notification.App, notification.Message)

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
