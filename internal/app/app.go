package app

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/morph/internal/category"
	"github.com/morph/third_party/mono"
)

func MonoHandler(w http.ResponseWriter, r *http.Request) {
}

func MonoWebHook(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling Mono WebHook...")

	payload, err := mono.ParseWebhookRequest(r)
	if err != nil {
		log.Printf("[Mono] Error parsing webhook: %s", err.Error())
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
	}

	transaction := fmt.Sprintf("Transaction Details:\n"+
		"Description: %s\n"+
		"MCC: %d\n"+
		"Category: %s\n"+
		"Amount: %.2f\n",
		payload.Data.StatementItem.Description,
		payload.Data.StatementItem.MCC,
		mmcCategory,
		payload.Data.StatementItem.AmountFloat())

	chatID, err := bot.GetChatID()
	if err != nil {
		log.Printf("[Mono] Error getting chat ID: %v", err)
	}

	bot.SendMessage(chatID, transaction, nil)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))

	go func() {
		categories := category.GetCategoriesInJSON()
		hints := category.GetHintsInJSON()

		ctx := context.Background()
		systemPrompt := "You're a data analyst. You have to classify the input into categories and subcategories. The input is a transaction from Bank. The output should be in JSON format with fields: category, subcategory, amount. The category and subcategory are strings. The amount is a float. If you can't find any proper categories, it should go to the Other category with no subcategory. The output should be like this: {\"category\": \"Children\", \"subcategory\": \"Vocal\", \"amount\": 400.0}. Here is the JSON of categories and subcategories: " + categories + "Also, here are some hints for categories: " + hints
		userPrompt := "The transaction from Bank is: " + transaction

		response := aiService.Request("Morph", "Translares Monobank transaction into: Category, Subcategory, Amount", systemPrompt, userPrompt, &ctx)
		if response == nil {
			log.Printf("[Bot] No response from AI")
		} else {
			log.Printf("[Bot] Response: %s %s %f", response.Category, response.Subcategory, response.Amount)
			linkMsg := fmt.Sprintf("Category: %s\nSubcategory: %s\nAmount: %.2f\n", response.Category, response.Subcategory, response.Amount)

			deepLink := deepLinkGenerator.Create(response.Category, response.Subcategory, "MonobankUAH", response.Amount)

			url, err := shortURLService.Shorten(deepLink)
			if err != nil {
				log.Printf("[Bot] Error shortening URL: %v", err)
			} else {
				log.Printf("[Bot] Shortened URL: %s", url)
				linkMsg += url
			}

			log.Printf("[Bot] Sending message to chat %d", chatID)
			bot.SendMessage(chatID, linkMsg, nil)
			log.Printf("[Bot] Sent message to chat %d", chatID)
		}
	}()
}
