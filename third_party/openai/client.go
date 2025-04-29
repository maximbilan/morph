package openai

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/invopop/jsonschema"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

type Response struct {
	Category    string  `json:"text"`
	Subcategory string  `json:"subcategory"`
	Amount      float64 `json:"amount"`
}

type OpenAI struct{}

func createAI() *openai.Client {
	client := openai.NewClient(
		option.WithAPIKey(os.Getenv("MORPH_AI_KEY")),
	)
	return &client
}

func generateSchema[T any]() interface{} {
	reflector := jsonschema.Reflector{
		AllowAdditionalProperties: false,
		DoNotReference:            true,
	}
	var v T
	schema := reflector.Reflect(v)
	return schema
}

func (service OpenAI) Request(name string, description string, systemPrompt string, userPrompt string, ctx *context.Context) *Response {
	ai := createAI()

	var responseSchema = generateSchema[Response]()

	schemaParam := openai.ResponseFormatJSONSchemaJSONSchemaParam{
		Name:        name,
		Description: openai.String(description),
		Schema:      responseSchema,
		Strict:      openai.Bool(true),
	}

	chat, err := ai.Chat.Completions.New(*ctx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(systemPrompt),
			openai.UserMessage(userPrompt),
		},
		ResponseFormat: openai.ChatCompletionNewParamsResponseFormatUnion{
			OfJSONSchema: &openai.ResponseFormatJSONSchemaParam{
				JSONSchema: schemaParam,
			},
		},
		Model: openai.ChatModelO3Mini,
	})

	response := Response{}
	if err != nil {
		log.Printf("[AI] Error parsing analysis: %s", err.Error())
		return nil
	}

	err = json.Unmarshal([]byte(chat.Choices[0].Message.Content), &response)
	if err != nil {
		log.Printf("[AI] Error parsing analysis: %s", err.Error())
		return nil
	}

	return &response
}
