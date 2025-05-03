package app

import (
	"context"
	"fmt"
	"log"
	"net/http"

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

func CashHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Started handling...")

	message := bot.Parse(r.Body)
	if message == nil {
		log.Printf("[Bot] No message to process")
		return
	}
	log.Printf("[Bot] Update: %s", message.Text)

	if message.Text != "" {
		ctx := context.Background()
		categories := category.GetCategoriesInJSON()
		hints := category.GetHintsInJSON()

		systemPrompt := "You're a data analyst. You have to classify the input into categories and subcategories. The input is a free text. The output should be in JSON format with fields: category, subcategory, amount. The category and subcategory are strings. The amount is a float. The input usually is in Ukrainian language. If you can't find any proper categories, it should go to the Other category with no subcategory. For example, the input is: '400 Вокал'. The output should be like this: {\"category\": \"Children\", \"subcategory\": \"Vocal\", \"amount\": 400.0}. Here is the JSON of categories and subcategories: " + categories + "Also, here are some hints for categories: " + hints
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
			msg += "\nError shortening URL: " + err.Error()
			return
		}
		log.Printf("[Bot] Shortened URL: %s", url)
		log.Printf("[Bot] Sending message to chat %d", message.ChatID)
		bot.SendMessage(message.ChatID, msg+"\n"+url, &message.MessageID)
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
	log.Println("Cash message handled")
}

func MonoWebHook(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling Mono WebHook...")

	payload, err := mono.ParseWebhookRequest(r)
	if err != nil {
		log.Printf("[Mono] Error parsing webhook: %s", err.Error())
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

	msg := transaction + "\n\n"

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
		deepLink := deepLinkGenerator.Create(response.Category, response.Subcategory, "MonobankUAH", response.Amount)

		url, err := shortURLService.Shorten(deepLink)
		if err != nil {
			log.Printf("[Bot] Error shortening URL: %v", err)
		} else {
			log.Printf("[Bot] Shortened URL: %s", url)
		}

		if deepLink != "" {
			msg += deepLink
		}
	}

	chatID, err := bot.GetChatID()
	if err != nil {
		log.Printf("[Mono] Error getting chat ID: %v", err)
	}

	bot.SendMessage(chatID, msg, nil)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
