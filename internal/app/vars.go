package app

import (
	"github.com/morph/internal/aiservice"
	"github.com/morph/internal/botservice"
	"github.com/morph/internal/deeplinkgenerator"
	"github.com/morph/internal/shorturl"
	"github.com/morph/internal/taskservice"
	"github.com/morph/third_party/googletasks"
	"github.com/morph/third_party/moneywiz"
	"github.com/morph/third_party/openai"
	"github.com/morph/third_party/shortio"
	"github.com/morph/third_party/telegram"
)

var bot botservice.BotService = telegram.Telegram{}
var aiService aiservice.AIService = openai.OpenAI{}
var shortURLService shorturl.ShortURL = shortio.ShortIO{}
var deepLinkGenerator deeplinkgenerator.DeepLinkGenerator = moneywiz.DeepLinkGenerator{}
var taskService taskservice.TaskService = googletasks.GoogleTasks{}
