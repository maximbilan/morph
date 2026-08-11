package app

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"time"

	"github.com/morph/internal/aiservice"
	"github.com/morph/internal/category"
	"github.com/morph/internal/taskservice"
)

const cashAccountName = "CashEUR"

// appendShortLink shortens deepLink and appends the result to text on a new
// line. If shortening fails it logs the full error and falls back to the raw
// deep link, so the user still receives a usable link instead of the
// shortener's (potentially huge) error page.
func appendShortLink(text, deepLink string) string {
	url, err := shortURLService.Shorten(deepLink)
	if err != nil {
		log.Printf("[Morph] Error shortening URL: %v", err)
		return text + "\n⚠️ Link not shortened:\n" + deepLink
	}
	log.Printf("[Morph] Shortened URL: %s", url)
	return text + "\n" + url
}

func CashHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("[Morph] Started cash handling...")

	message := bot.Parse(r.Body)
	if message == nil {
		log.Printf("[Morph] No message to process")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
		return
	}
	log.Printf("[Morph] Update: %s", message.Text)

	if message.Text == "" {
		log.Printf("[Morph] No text in message")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
		return
	}

	ctx := context.Background()
	taskService.Connect(&ctx)
	defer taskService.Close()

	response := aiService.Classify(aiservice.Request{
		Name:         "Morph",
		Description:  "Translates free input into: Category, Subcategory, Amount",
		SystemPrompt: category.ClassificationPrompt(),
		UserPrompt:   "Classify this free-text cash expense.\nText: " + message.Text,
		AllowedPaths: category.Paths(),
	}, &ctx)
	if response == nil {
		log.Printf("[Morph] No response from AI")

		scheduledMessage := taskservice.ScheduledMessage{
			ChatID:           message.ChatID,
			Text:             "No response from AI",
			ReplyToMessageID: &message.MessageID,
		}

		taskService.ScheduleMessage(&ctx, scheduledMessage, time.Now())
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
		return
	}

	absoluteAmount := math.Abs(response.Amount)
	cat, subcategory := category.SplitPath(response.CategoryPath)

	log.Printf("[Morph] Response: %s %s %f", cat, subcategory, absoluteAmount)
	text := "Category: " + cat + "\nSubcategory: " + subcategory + "\nAmount: " + fmt.Sprintf("%.2f", absoluteAmount)
	deepLink := deepLinkGenerator.Create(cat, subcategory, cashAccountName, absoluteAmount, time.Now())

	text = appendShortLink(text, deepLink)

	log.Printf("[Morph] Sending message to chat %d", message.ChatID)

	scheduledMessage := taskservice.ScheduledMessage{
		ChatID:           message.ChatID,
		Text:             text,
		ReplyToMessageID: &message.MessageID,
	}

	taskService.ScheduleMessage(&ctx, scheduledMessage, time.Now())

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
	log.Println("Cash handler finished")
}

// getAccountNameFromID maps Monobank account IDs to account names for deep links
func getAccountNameFromID(accountID string) string {
	accountMap := map[string]string{
		"a-dnHAO9ExLnboGJP_pdwA": "MonobankUAH",
		"Llx31dyYA8dahhShny5Vvw": "MonobankUAHWhite",
		"WKl9I-LztrH1ZWeafLZEzQ": "MonobankEUR",
		"uHsC3WXdFl0H5CucFXfTHg": "MonobankUSD",
		"NnyWiNGakLsDRXkTe-EQ9A": "MonoeAid",
		"9mnHzIA1Fkjn7kmeKiAoGg": "MonoFOPUAH",
		"uUms_k2kDlN6Uyofrs72gw": "MonoFOPUSD",
	}

	if accountName, ok := accountMap[accountID]; ok {
		return accountName
	}

	// Default fallback if account ID is not found
	log.Printf("[Morph] Unknown account ID: %s, using MonobankUAH as default", accountID)
	return "MonobankUAH"
}

func MonoHandler(w http.ResponseWriter, r *http.Request) {
	var transaction taskservice.ScheduledTransaction
	if err := json.NewDecoder(r.Body).Decode(&transaction); err != nil {
		log.Printf("[Morph] Could not parse transaction %s", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Could not parse transaction"))
		return
	}

	ctx := context.Background()
	taskService.Connect(&ctx)
	defer taskService.Close()

	// The MCC description used to be labelled "category:", colliding with the
	// category the model had to produce.
	userPrompt := fmt.Sprintf("Classify this bank transaction.\nMerchant: %s\nMCC: %d (%s)\nAmount: %.2f",
		transaction.Description, transaction.MCC, transaction.MCCDescription, transaction.Amount)

	chatId := transaction.ChatID
	response := aiService.Classify(aiservice.Request{
		Name:         "Morph",
		Description:  "Translates a Monobank transaction into: Category, Subcategory, Amount",
		SystemPrompt: category.ClassificationPrompt(),
		UserPrompt:   userPrompt,
		AllowedPaths: category.Paths(),
	}, &ctx)
	if response == nil {
		log.Printf("[Morph] No response from AI")
		scheduledMessage := taskservice.ScheduledMessage{
			ChatID:           chatId,
			Text:             "No response from AI",
			ReplyToMessageID: nil,
		}
		taskService.ScheduleMessage(&ctx, scheduledMessage, time.Now())
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
		return
	}

	// The webhook already carries the exact amount; don't trust the model's echo.
	absoluteAmount := math.Abs(transaction.Amount)
	cat, subcategory := category.SplitPath(response.CategoryPath)

	log.Printf("[Morph] Response: %s %s %f", cat, subcategory, absoluteAmount)
	linkMsg := fmt.Sprintf("Category: %s\nSubcategory: %s\nAmount: %.2f", cat, subcategory, absoluteAmount)
	if transaction.IsRefund {
		linkMsg += "\n🔄 Refund"
	}

	// Determine the transaction time. Mono API may provide time in seconds or milliseconds since epoch.
	// Use a heuristic: treat large values as milliseconds.
	var txTime time.Time
	if transaction.Time > 1e12 {
		txTime = time.Unix(transaction.Time/1000, 0)
	} else {
		txTime = time.Unix(transaction.Time, 0)
	}

	// Get account name from account ID
	accountName := getAccountNameFromID(transaction.AccountID)
	log.Printf("[Morph] Account ID: %s, Account Name: %s", transaction.AccountID, accountName)

	deepLink := deepLinkGenerator.Create(cat, subcategory, accountName, absoluteAmount, txTime)

	linkMsg = appendShortLink(linkMsg, deepLink)

	log.Printf("[Morph] Sending message to chat %d", chatId)

	scheduledMessage := taskservice.ScheduledMessage{
		ChatID:           chatId,
		Text:             linkMsg,
		ReplyToMessageID: nil,
	}
	taskService.ScheduleMessage(&ctx, scheduledMessage, time.Now())

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
	log.Println("[Morph] Mono handler finished")
}
