package app

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/morph/internal/category"
	"github.com/morph/third_party/openai"
	"github.com/morph/third_party/telegram"
)

var bot telegram.Telegram
var aiService openai.OpenAI

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

		systemPrompt := "You're a data analyst. You have to classify the input into categories and subcategories. The input is a free text. The output should be in JSON format with fields: category, subcategory, amount. The category and subcategory are strings. The amount is a float. The input usually is in Ukrainian language. For example, the input is: '400 Вокал'. The output should be like this: {\"category\": \"Children\", \"subcategory\": \"Vocal\", \"amount\": 400.0}. Here is the JSON of categories and subcategories: " + category.GetCategoriesInJSON()
		userPrompt := "The input is: " + message.Text

		response := aiService.Request("Morph", "Translares free input into: Category, Subcategory, Amount", systemPrompt, userPrompt, &ctx)
		if response == nil {
			log.Printf("[Bot] No response from AI")
			return
		}
		log.Printf("[Bot] Response: %s %s %f", response.Category, response.Subcategory, response.Amount)

		bot.SendMessage(message.ChatID, "Category: "+response.Category+" Subcategory: "+response.Subcategory+" Amount: "+fmt.Sprintf("%f", response.Amount))
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
	log.Println("Handled")
}
