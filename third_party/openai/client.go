package openai

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/morph/internal/aiservice"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

// defaultModel is overridable via MORPH_AI_MODEL. mini rather than nano: on 44
// replayed real transactions nano hits 66-77% of leaves, mini 91-95% with no
// "Other" fallbacks. See cmd/classifyeval.
const defaultModel = "gpt-5.4-mini"

type OpenAI struct{}

func createAI() *openai.Client {
	client := openai.NewClient(
		option.WithAPIKey(os.Getenv("MORPH_AI_KEY")),
	)
	return &client
}

func model() string {
	if name := os.Getenv("MORPH_AI_MODEL"); name != "" {
		return name
	}
	return defaultModel
}

// responseSchema is hand-built rather than reflected so categoryPath can carry
// the taxonomy as an enum, which makes an invalid category ungeneratable.
func responseSchema(allowedPaths []string) map[string]any {
	categoryPath := map[string]any{
		"type":        "string",
		"description": "The single best matching taxonomy path.",
	}
	if len(allowedPaths) > 0 {
		categoryPath["enum"] = allowedPaths
	}

	// Property order is generation order: merchant first, then the leaf.
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"merchant": map[string]any{
				"type":        "string",
				"description": "What the counterparty actually is, in 1-6 words, for example 'Spanish supermarket chain' or 'own account transfer'. Fill this in first.",
			},
			"isTransaction": map[string]any{
				"type":        "boolean",
				"description": "True only for a real debit or credit on an account. False for promotional, security, login or informational messages.",
			},
			"amount": map[string]any{
				"type":        "number",
				"description": "The transaction amount, or 0 when there is no transaction.",
			},
			"categoryPath": categoryPath,
		},
		"required":             []string{"merchant", "isTransaction", "amount", "categoryPath"},
		"additionalProperties": false,
	}
}

func (service OpenAI) Classify(req aiservice.Request, ctx *context.Context) *aiservice.Response {
	startTime := time.Now()
	defer func() {
		log.Printf("[OpenAI] Request took %v", time.Since(startTime))
	}()

	ai := createAI()
	name := model()

	params := openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(req.SystemPrompt),
			openai.UserMessage(req.UserPrompt),
		},
		ResponseFormat: openai.ChatCompletionNewParamsResponseFormatUnion{
			OfJSONSchema: &openai.ResponseFormatJSONSchemaParam{
				JSONSchema: openai.ResponseFormatJSONSchemaJSONSchemaParam{
					Name:        req.Name,
					Description: openai.String(req.Description),
					Schema:      responseSchema(req.AllowedPaths),
					Strict:      openai.Bool(true),
				},
			},
		},
		Model: name,
	}

	chat, err := ai.Chat.Completions.New(*ctx, params)
	if err != nil {
		log.Printf("[AI] Request to %s failed: %s", name, err.Error())
		return nil
	}
	if len(chat.Choices) == 0 {
		log.Printf("[AI] Model %s returned no choices", name)
		return nil
	}

	choice := chat.Choices[0]
	if choice.Message.Refusal != "" {
		log.Printf("[AI] Model %s refused: %s", name, choice.Message.Refusal)
		return nil
	}
	if choice.FinishReason != "stop" {
		// "length" means truncated JSON, which will not parse below.
		log.Printf("[AI] Model %s finished with reason %q", name, choice.FinishReason)
	}

	response := aiservice.Response{}
	if err := json.Unmarshal([]byte(choice.Message.Content), &response); err != nil {
		log.Printf("[AI] Could not parse analysis: %s", err.Error())
		return nil
	}

	log.Printf("[AI] %s: merchant=%q path=%q tokens=%d", name, response.Merchant, response.CategoryPath, chat.Usage.CompletionTokens)
	return &response
}
