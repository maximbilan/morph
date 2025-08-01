package app

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"time"

	"github.com/morph/internal/category"
	"github.com/morph/internal/taskservice"
)

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

	categories := category.GetCategoriesInJSON()
	hints := category.GetHintsInJSON()

	systemPrompt := "You are a data analyst. Your task is to classify the input into a category, subcategory, and amount. You MUST ONLY use the categories and subcategories provided below—do not invent new ones. If the input does not match any, use 'Other' for category and an empty string for subcategory. Output a single-line JSON object with only these fields: category, subcategory, amount. Example of the output: {\"category\": \"Children\", \"subcategory\": \"Vocal\", \"amount\": 400.0}. Categories and subcategories: " + categories + " Hints: " + hints + " IMPORTANT: Do not add any explanation or extra text. Only output the JSON object."
	userPrompt := "Classify this input: " + message.Text

	response := aiService.Request("Morph", "Translares free input into: Category, Subcategory, Amount", systemPrompt, userPrompt, &ctx)
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

	log.Printf("[Morph] Response: %s %s %f", response.Category, response.Subcategory, absoluteAmount)
	text := "Category: " + response.Category + "\nSubcategory: " + response.Subcategory + "\nAmount: " + fmt.Sprintf("%.2f", absoluteAmount)
	deepLink := deepLinkGenerator.Create(response.Category, response.Subcategory, "Cash", absoluteAmount)

	url, err := shortURLService.Shorten(deepLink)
	if err != nil {
		log.Printf("[Morph] Error shortening URL: %v", err)
		text += "\nError shortening URL: " + err.Error()
	} else {
		log.Printf("[Morph] Shortened URL: %s", url)
		text += "\n" + url
	}

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

	transactionStr := fmt.Sprintf("{ mcc: %d, description: %s, category: %s, amount: %.2f }", transaction.MCC, transaction.Description, transaction.Category, transaction.Amount)
	categories := category.GetCategoriesInJSON()
	hints := category.GetHintsInJSON()

	systemPrompt := "You are a data analyst. Your task is to classify the bank transaction into a category, subcategory, and amount. You MUST ONLY use the categories and subcategories provided below—do not invent new ones. If the input does not match any, use 'Other' for category and an empty string for subcategory. Output a single-line JSON object with only these fields: category, subcategory, amount. Example of the output: {\"category\": \"Children\", \"subcategory\": \"Vocal\", \"amount\": 400.0}. Categories and subcategories: " + categories + " Hints: " + hints + " IMPORTANT: Do not add any explanation or extra text. Only output the JSON object."
	userPrompt := "Classify this bank transaction: " + transactionStr

	chatId := transaction.ChatID
	response := aiService.Request("Morph", "Translares Monobank transaction into: Category, Subcategory, Amount", systemPrompt, userPrompt, &ctx)
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

	absoluteAmount := math.Abs(response.Amount)

	log.Printf("[Morph] Response: %s %s %f", response.Category, response.Subcategory, absoluteAmount)
	linkMsg := fmt.Sprintf("Category: %s\nSubcategory: %s\nAmount: %.2f\n", response.Category, response.Subcategory, absoluteAmount)
	deepLink := deepLinkGenerator.Create(response.Category, response.Subcategory, "MonobankUAH", absoluteAmount)

	url, err := shortURLService.Shorten(deepLink)
	if err != nil {
		log.Printf("[Morph] Error shortening URL: %v", err)
		linkMsg += "\nError shortening URL: " + err.Error()
	} else {
		log.Printf("[Morph] Shortened URL: %s", url)
		linkMsg += url
	}

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
