package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

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

		categories := category.GetCategoriesInJSON()
		systemPrompt := "You're a data analyst. You have to classify the input into categories and subcategories. The input is a free text. The output should be in JSON format with fields: category, subcategory, amount. The category and subcategory are strings. The amount is a float. The input usually is in Ukrainian language. For example, the input is: '400 Вокал'. The output should be like this: {\"category\": \"Children\", \"subcategory\": \"Vocal\", \"amount\": 400.0}. Here is the JSON of categories and subcategories: " + categories
		userPrompt := "The input is: " + message.Text

		response := aiService.Request("Morph", "Translares free input into: Category, Subcategory, Amount", systemPrompt, userPrompt, &ctx)
		if response == nil {
			log.Printf("[Bot] No response from AI")
			return
		}
		log.Printf("[Bot] Response: %s %s %f", response.Category, response.Subcategory, response.Amount)

		bot.SendMessage(message.ChatID, "Category: "+response.Category+"\nSubcategory: "+response.Subcategory+"\nAmount: "+fmt.Sprintf("%f", response.Amount))

		deepLink := fmt.Sprintf("moneywiz://expense?amount=%.2f&account=Cash&category=%s/%s&save=true", response.Amount, response.Category, response.Subcategory)

		url, err := shortenURL(deepLink)
		if err != nil {
			log.Printf("[Bot] Error shortening URL: %v", err)
			return
		}
		log.Printf("[Bot] Shortened URL: %s", url)

		bot.SendMessage(message.ChatID, url)
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
	log.Println("Handled")
}

type ShortenRequest struct {
	Domain      string `json:"domain"`
	OriginalURL string `json:"originalURL"`
}

type ShortenResponse struct {
	ShortURL string `json:"shortURL"`
}

func shortenURL(originalURL string) (string, error) {
	apiURL := "https://api.short.io/links"
	requestBody := ShortenRequest{
		Domain:      "morph-service.short.gy",
		OriginalURL: originalURL,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %v", err)
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", os.Getenv("MORPH_REDIRECT_KEY"))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to shorten URL: %s", string(body))
	}

	var response ShortenResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode response: %v", err)
	}

	return response.ShortURL, nil
}
