package app

import (
	"context"
	"fmt"
	"log"
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
	categories := category.GetCategoriesInJSON()
	hints := category.GetHintsInJSON()

	systemPrompt := "You're a data analyst. You have to classify the input into categories and subcategories. The input is a free text. The output should be in JSON format with fields: category, subcategory, amount. The category and subcategory are strings. The amount is a float. The input usually is in Ukrainian language. If you can't find any proper categories, it should go to the Other category with no subcategory. For example, the input is: '400 Вокал'. The output should be like this: {\"category\": \"Children\", \"subcategory\": \"Vocal\", \"amount\": 400.0}. Here is the JSON of categories and subcategories: " + categories + "Also, here are some hints for categories: " + hints
	userPrompt := "The input is: " + message.Text

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

	log.Printf("[Morph] Response: %s %s %f", response.Category, response.Subcategory, response.Amount)
	text := "Category: " + response.Category + "\nSubcategory: " + response.Subcategory + "\nAmount: " + fmt.Sprintf("%.2f", response.Amount)
	deepLink := deepLinkGenerator.Create(response.Category, response.Subcategory, "Cash", response.Amount)

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
