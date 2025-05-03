package aiservice

import "context"

type Response struct {
	Category    string  `json:"text"`
	Subcategory string  `json:"subcategory"`
	Amount      float64 `json:"amount"`
}

type AIService interface {
	Request(name string, description string, systemPrompt string, userPrompt string, ctx *context.Context) *Response
}
