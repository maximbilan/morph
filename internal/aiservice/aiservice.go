package aiservice

import "context"

// Request is one classification call. The provider turns AllowedPaths into a
// JSON schema enum, so an out-of-taxonomy answer is impossible rather than just
// discouraged.
type Request struct {
	Name         string
	Description  string
	SystemPrompt string
	UserPrompt   string
	AllowedPaths []string
}

// Response is the model's verdict. Merchant comes before CategoryPath so the
// model names the counterparty before picking a leaf.
type Response struct {
	Merchant      string  `json:"merchant"`
	IsTransaction bool    `json:"isTransaction"`
	Amount        float64 `json:"amount"`
	CategoryPath  string  `json:"categoryPath"`
}

type AIService interface {
	Classify(req Request, ctx *context.Context) *Response
}
