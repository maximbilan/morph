package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/maximbilan/mcc"
	"github.com/morph/internal/category"
	"github.com/morph/third_party/moneywiz"
	"github.com/morph/third_party/mono"
	"github.com/morph/third_party/openai"
	"github.com/morph/third_party/shortio"
	"github.com/morph/third_party/telegram"
)

var bot telegram.Telegram
var aiService openai.OpenAI
var shortURLService shortio.ShortIO
var deepLinkGenerator moneywiz.DeepLinkGenerator

func Handle(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling...")

	message := bot.Parse(r.Body)
	if message == nil {
		log.Printf("[Bot] No message to process")
		return
	}
	log.Printf("[Bot] Update: %s", message.Text)

	if message.Text != "" {
		ctx := context.Background()

		categories := category.GetCategoriesInJSON()
		systemPrompt := "You're a data analyst. You have to classify the input into categories and subcategories. The input is a free text. The output should be in JSON format with fields: category, subcategory, amount. The category and subcategory are strings. The amount is a float. The input usually is in Ukrainian language. For example, the input is: '400 Вокал'. The output should be like this: {\"category\": \"Children\", \"subcategory\": \"Vocal\", \"amount\": 400.0}. Here is the JSON of categories and subcategories: " + categories
		userPrompt := "The input is: " + message.Text

		response := aiService.Request("Morph", "Translares free input into: Category, Subcategory, Amount", systemPrompt, userPrompt, &ctx)
		if response == nil {
			log.Printf("[Bot] No response from AI")
			return
		}
		log.Printf("[Bot] Response: %s %s %f", response.Category, response.Subcategory, response.Amount)

		msg := "Category: " + response.Category + "\nSubcategory: " + response.Subcategory + "\nAmount: " + fmt.Sprintf("%.2f", response.Amount)

		deepLink := deepLinkGenerator.Create(response.Category, response.Subcategory, "Cash", response.Amount)

		url, err := shortURLService.Shorten(deepLink)
		if err != nil {
			log.Printf("[Bot] Error shortening URL: %v", err)
			return
		}
		log.Printf("[Bot] Shortened URL: %s", url)
		log.Printf("[Bot] Sending message to chat %d", message.ChatID)

		bot.SendMessage(message.ChatID, msg+"\n"+url, &message.MessageID)
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
	log.Println("Handled")
}

func MonoWebHook(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling Mono WebHook...")

	payload, err := mono.ParseWebhookRequest(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	chatIDStr := os.Getenv("MORPH_TELEGRAM_CHAT_ID")
	chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil {
		log.Printf("[Mono] Error converting chatID to int64: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Printf("[Mono] Sending message to chat %d", chatID)

	category, err := mcc.GetCategory(payload.Data.StatementItem.MCC)
	if err != nil {
		log.Printf("[Mono] Error getting category: %v", err)
	}

	msg := fmt.Sprintf("Transaction Details:\n"+
		"Description: %s\n"+
		"MCC: %s\n"+
		"Category: %s\n"+
		"Amount: %d\n"+
		"Balance: %d\n"+
		"Receipt: %s",
		payload.Data.StatementItem.Description,
		payload.Data.StatementItem.MCC,
		category,
		payload.Data.StatementItem.Amount,
		payload.Data.StatementItem.Balance,
		payload.Data.StatementItem.ReceiptID)

	bot.SendMessage(chatID, msg, nil)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
	log.Println("Handled Mono WebHook")
}
