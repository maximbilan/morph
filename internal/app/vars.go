package app

import (
	"github.com/morph/internal/shorturl"
	"github.com/morph/third_party/googletasks"
	"github.com/morph/third_party/moneywiz"
	"github.com/morph/third_party/openai"
	"github.com/morph/third_party/telegram"
)

var bot telegram.Telegram
var aiService openai.OpenAI
var shortURLService shorturl.ShortURL
var deepLinkGenerator moneywiz.DeepLinkGenerator
var taskService googletasks.GoogleTasks

func init() {
	// Initialize the fallback URL shortening service
	shortURLService = shorturl.CreateDefaultFallbackService()
}
