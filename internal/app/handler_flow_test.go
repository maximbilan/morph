package app

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/morph/internal/aiservice"
	"github.com/morph/internal/botservice"
	"github.com/morph/internal/taskservice"
)

type fakeBot struct {
	message       *botservice.BotMessage
	chatID        int64
	chatIDErr     error
	parseCalls    int
	sentMessages  []taskservice.ScheduledMessage
	sendCallCount int
}

func (b *fakeBot) GetChatID() (int64, error) {
	return b.chatID, b.chatIDErr
}

func (b *fakeBot) Parse(body io.ReadCloser) *botservice.BotMessage {
	b.parseCalls++
	return b.message
}

func (b *fakeBot) SendMessage(chatID int64, text string, replyToMessageID *int64) {
	b.sendCallCount++
	b.sentMessages = append(b.sentMessages, taskservice.ScheduledMessage{
		ChatID:           chatID,
		Text:             text,
		ReplyToMessageID: replyToMessageID,
	})
}

type fakeAI struct {
	response   *aiservice.Response
	callCount  int
	userPrompt string
}

func (a *fakeAI) Request(name string, description string, systemPrompt string, userPrompt string, ctx *context.Context) *aiservice.Response {
	a.callCount++
	a.userPrompt = userPrompt
	return a.response
}

type deepLinkCall struct {
	category    string
	subcategory string
	account     string
	amount      float64
	date        time.Time
}

type fakeDeepLinkGenerator struct {
	link      string
	callCount int
	calls     []deepLinkCall
}

func (g *fakeDeepLinkGenerator) Create(category string, subcategory string, account string, amount float64, date time.Time) string {
	g.callCount++
	g.calls = append(g.calls, deepLinkCall{
		category:    category,
		subcategory: subcategory,
		account:     account,
		amount:      amount,
		date:        date,
	})
	if g.link != "" {
		return g.link
	}
	return "moneywiz://expense"
}

type fakeShortURL struct {
	url       string
	err       error
	callCount int
	inputs    []string
}

func (s *fakeShortURL) Shorten(URL string) (string, error) {
	s.callCount++
	s.inputs = append(s.inputs, URL)
	return s.url, s.err
}

type fakeTaskService struct {
	connectCount          int
	closeCount            int
	scheduledMessages     []taskservice.ScheduledMessage
	scheduledTransactions []taskservice.ScheduledTransaction
}

func (s *fakeTaskService) Connect(ctx *context.Context) {
	s.connectCount++
}

func (s *fakeTaskService) Close() {
	s.closeCount++
}

func (s *fakeTaskService) ScheduleMessage(ctx *context.Context, scheduledMessage taskservice.ScheduledMessage, timeOffset time.Time) {
	s.scheduledMessages = append(s.scheduledMessages, scheduledMessage)
}

func (s *fakeTaskService) ScheduleTransaction(ctx *context.Context, scheduledTransaction taskservice.ScheduledTransaction, timeOffset time.Time) {
	s.scheduledTransactions = append(s.scheduledTransactions, scheduledTransaction)
}

type appFakes struct {
	bot      *fakeBot
	ai       *fakeAI
	shortURL *fakeShortURL
	deepLink *fakeDeepLinkGenerator
	tasks    *fakeTaskService
}

func installAppFakes(t *testing.T) appFakes {
	t.Helper()

	oldBot := bot
	oldAIService := aiService
	oldShortURLService := shortURLService
	oldDeepLinkGenerator := deepLinkGenerator
	oldTaskService := taskService

	fakes := appFakes{
		bot:      &fakeBot{chatID: 12345},
		ai:       &fakeAI{},
		shortURL: &fakeShortURL{url: "https://short.example/link"},
		deepLink: &fakeDeepLinkGenerator{link: "moneywiz://expense"},
		tasks:    &fakeTaskService{},
	}

	bot = fakes.bot
	aiService = fakes.ai
	shortURLService = fakes.shortURL
	deepLinkGenerator = fakes.deepLink
	taskService = fakes.tasks

	t.Cleanup(func() {
		bot = oldBot
		aiService = oldAIService
		shortURLService = oldShortURLService
		deepLinkGenerator = oldDeepLinkGenerator
		taskService = oldTaskService
	})

	return fakes
}

func TestCashHandler_NoMessageReturnsOK(t *testing.T) {
	fakes := installAppFakes(t)

	req := httptest.NewRequest(http.MethodPost, "/cashHandler", strings.NewReader(`{}`))
	rr := httptest.NewRecorder()

	CashHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	if fakes.bot.parseCalls != 1 {
		t.Fatalf("parseCalls = %d, want 1", fakes.bot.parseCalls)
	}
	if fakes.tasks.connectCount != 0 {
		t.Fatalf("task service connected %d times, want 0", fakes.tasks.connectCount)
	}
}

func TestCashHandler_NoAIResponseSchedulesErrorMessage(t *testing.T) {
	fakes := installAppFakes(t)
	messageID := int64(77)
	fakes.bot.message = &botservice.BotMessage{
		MessageID: messageID,
		ChatID:    555,
		Text:      "coffee 12",
	}

	req := httptest.NewRequest(http.MethodPost, "/cashHandler", strings.NewReader(`{}`))
	rr := httptest.NewRecorder()

	CashHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	if fakes.tasks.connectCount != 1 || fakes.tasks.closeCount != 1 {
		t.Fatalf("task connect/close = %d/%d, want 1/1", fakes.tasks.connectCount, fakes.tasks.closeCount)
	}
	if len(fakes.tasks.scheduledMessages) != 1 {
		t.Fatalf("scheduled messages = %d, want 1", len(fakes.tasks.scheduledMessages))
	}
	got := fakes.tasks.scheduledMessages[0]
	if got.ChatID != 555 || got.Text != "No response from AI" {
		t.Fatalf("scheduled message = %+v, want chat 555 with AI error", got)
	}
	if got.ReplyToMessageID == nil || *got.ReplyToMessageID != messageID {
		t.Fatalf("reply message ID = %v, want %d", got.ReplyToMessageID, messageID)
	}
	if fakes.shortURL.callCount != 0 {
		t.Fatalf("short URL calls = %d, want 0", fakes.shortURL.callCount)
	}
}

func TestCashHandler_HappyPathUsesCashEURAndSchedulesShortLink(t *testing.T) {
	fakes := installAppFakes(t)
	messageID := int64(88)
	fakes.bot.message = &botservice.BotMessage{
		MessageID: messageID,
		ChatID:    777,
		Text:      "groceries 42.5",
	}
	fakes.ai.response = &aiservice.Response{
		Category:    "Food",
		Subcategory: "Shop",
		Amount:      -42.5,
	}

	req := httptest.NewRequest(http.MethodPost, "/cashHandler", strings.NewReader(`{}`))
	rr := httptest.NewRecorder()

	CashHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	if fakes.deepLink.callCount != 1 {
		t.Fatalf("deep link calls = %d, want 1", fakes.deepLink.callCount)
	}
	deepLink := fakes.deepLink.calls[0]
	if deepLink.account != cashAccountName {
		t.Fatalf("deep link account = %q, want %q", deepLink.account, cashAccountName)
	}
	if deepLink.amount != 42.5 {
		t.Fatalf("deep link amount = %.2f, want 42.50", deepLink.amount)
	}
	if len(fakes.shortURL.inputs) != 1 || fakes.shortURL.inputs[0] != "moneywiz://expense" {
		t.Fatalf("short URL inputs = %v, want moneywiz deep link", fakes.shortURL.inputs)
	}
	if len(fakes.tasks.scheduledMessages) != 1 {
		t.Fatalf("scheduled messages = %d, want 1", len(fakes.tasks.scheduledMessages))
	}
	got := fakes.tasks.scheduledMessages[0]
	wantText := "Category: Food\nSubcategory: Shop\nAmount: 42.50\nhttps://short.example/link"
	if got.Text != wantText {
		t.Fatalf("scheduled text = %q, want %q", got.Text, wantText)
	}
	if got.ReplyToMessageID == nil || *got.ReplyToMessageID != messageID {
		t.Fatalf("reply message ID = %v, want %d", got.ReplyToMessageID, messageID)
	}
}

func TestCashHandler_ShortURLErrorFallsBackToRawDeepLink(t *testing.T) {
	fakes := installAppFakes(t)
	fakes.bot.message = &botservice.BotMessage{
		MessageID: 1,
		ChatID:    2,
		Text:      "taxi 10",
	}
	fakes.ai.response = &aiservice.Response{
		Category:    "Transport",
		Subcategory: "Taxi",
		Amount:      10,
	}
	fakes.shortURL.err = errors.New("shortener down")

	req := httptest.NewRequest(http.MethodPost, "/cashHandler", strings.NewReader(`{}`))
	rr := httptest.NewRecorder()

	CashHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	if len(fakes.tasks.scheduledMessages) != 1 {
		t.Fatalf("scheduled messages = %d, want 1", len(fakes.tasks.scheduledMessages))
	}
	got := fakes.tasks.scheduledMessages[0].Text
	if !strings.Contains(got, "moneywiz://expense") {
		t.Fatalf("scheduled text = %q, want raw deep link fallback", got)
	}
	if strings.Contains(got, "shortener down") {
		t.Fatalf("scheduled text = %q, should not leak the shortener error", got)
	}
}

func TestMonoHandler_InvalidJSONReturnsBadRequest(t *testing.T) {
	installAppFakes(t)

	req := httptest.NewRequest(http.MethodPost, "/monoHandler", strings.NewReader(`{invalid`))
	rr := httptest.NewRecorder()

	MonoHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
	if rr.Body.String() != "Could not parse transaction" {
		t.Fatalf("body = %q, want parse error", rr.Body.String())
	}
}

func TestMonoHandler_NoAIResponseSchedulesErrorMessage(t *testing.T) {
	fakes := installAppFakes(t)

	req := httptest.NewRequest(http.MethodPost, "/monoHandler", strings.NewReader(`{"chatId":321,"mcc":4121,"category":"Transport","description":"Bolt","amount":120,"time":1746194127}`))
	rr := httptest.NewRecorder()

	MonoHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	if len(fakes.tasks.scheduledMessages) != 1 {
		t.Fatalf("scheduled messages = %d, want 1", len(fakes.tasks.scheduledMessages))
	}
	got := fakes.tasks.scheduledMessages[0]
	if got.ChatID != 321 || got.Text != "No response from AI" || got.ReplyToMessageID != nil {
		t.Fatalf("scheduled message = %+v, want Mono AI error", got)
	}
}

func TestMonoHandler_HappyPathUsesAccountMappingRefundAndMillisecondTime(t *testing.T) {
	fakes := installAppFakes(t)
	fakes.ai.response = &aiservice.Response{
		Category:    "Transport",
		Subcategory: "Taxi",
		Amount:      -176,
	}

	req := httptest.NewRequest(http.MethodPost, "/monoHandler", strings.NewReader(`{"chatId":654,"mcc":4121,"category":"Transport","description":"Скасування. Bolt","amount":176,"time":1746194127000,"isRefund":true,"accountId":"WKl9I-LztrH1ZWeafLZEzQ"}`))
	rr := httptest.NewRecorder()

	MonoHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	if len(fakes.deepLink.calls) != 1 {
		t.Fatalf("deep link calls = %d, want 1", len(fakes.deepLink.calls))
	}
	deepLink := fakes.deepLink.calls[0]
	if deepLink.account != "MonobankEUR" {
		t.Fatalf("deep link account = %q, want MonobankEUR", deepLink.account)
	}
	if deepLink.amount != 176 {
		t.Fatalf("deep link amount = %.2f, want 176.00", deepLink.amount)
	}
	if !deepLink.date.Equal(time.Unix(1746194127, 0)) {
		t.Fatalf("deep link date = %s, want unix 1746194127", deepLink.date)
	}
	if len(fakes.tasks.scheduledMessages) != 1 {
		t.Fatalf("scheduled messages = %d, want 1", len(fakes.tasks.scheduledMessages))
	}
	got := fakes.tasks.scheduledMessages[0]
	if !strings.Contains(got.Text, "🔄 Refund") {
		t.Fatalf("scheduled text = %q, want refund marker", got.Text)
	}
	if !strings.Contains(got.Text, "https://short.example/link") {
		t.Fatalf("scheduled text = %q, want short link", got.Text)
	}
}

func TestSendMessage_InvalidJSONReturnsBadRequest(t *testing.T) {
	installAppFakes(t)

	req := httptest.NewRequest(http.MethodPost, "/sendMessage", strings.NewReader(`{invalid`))
	rr := httptest.NewRecorder()

	SendMessage(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestSendMessage_HappyPathSendsViaBot(t *testing.T) {
	fakes := installAppFakes(t)

	req := httptest.NewRequest(http.MethodPost, "/sendMessage", strings.NewReader(`{"chatId":123,"text":"hello","reply_to_message_id":456}`))
	rr := httptest.NewRecorder()

	SendMessage(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	if fakes.bot.sendCallCount != 1 {
		t.Fatalf("send calls = %d, want 1", fakes.bot.sendCallCount)
	}
	got := fakes.bot.sentMessages[0]
	if got.ChatID != 123 || got.Text != "hello" || got.ReplyToMessageID == nil || *got.ReplyToMessageID != 456 {
		t.Fatalf("sent message = %+v, want chat 123 text hello reply 456", got)
	}
}

func TestMonoWebHook_GETReturnsOK(t *testing.T) {
	fakes := installAppFakes(t)

	req := httptest.NewRequest(http.MethodGet, "/monoWebHook", nil)
	rr := httptest.NewRecorder()

	MonoWebHook(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	if rr.Body.String() != "OK" {
		t.Fatalf("body = %q, want OK", rr.Body.String())
	}
	if fakes.bot.chatIDErr != nil || fakes.tasks.connectCount != 0 {
		t.Fatalf("GET should not initialize dependencies")
	}
}

func TestMonoWebHook_ChatIDFailureReturnsInternalServerError(t *testing.T) {
	fakes := installAppFakes(t)
	fakes.bot.chatIDErr = errors.New("missing chat id")

	req := httptest.NewRequest(http.MethodPost, "/monoWebHook", strings.NewReader(`{}`))
	rr := httptest.NewRecorder()

	MonoWebHook(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusInternalServerError)
	}
	if rr.Body.String() != "Could not get chat ID" {
		t.Fatalf("body = %q, want chat ID error", rr.Body.String())
	}
	if fakes.tasks.connectCount != 0 {
		t.Fatalf("task service connected %d times, want 0", fakes.tasks.connectCount)
	}
}

func TestMonoWebHook_BadJSONReturnsBadRequest(t *testing.T) {
	fakes := installAppFakes(t)

	req := httptest.NewRequest(http.MethodPost, "/monoWebHook", strings.NewReader(`{invalid`))
	rr := httptest.NewRecorder()

	MonoWebHook(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
	if rr.Body.String() != "Could not parse data" {
		t.Fatalf("body = %q, want parse error", rr.Body.String())
	}
	if fakes.tasks.connectCount != 1 || fakes.tasks.closeCount != 1 {
		t.Fatalf("task connect/close = %d/%d, want 1/1", fakes.tasks.connectCount, fakes.tasks.closeCount)
	}
	if len(fakes.tasks.scheduledMessages) != 0 || len(fakes.tasks.scheduledTransactions) != 0 {
		t.Fatalf("scheduled messages/transactions = %d/%d, want 0/0", len(fakes.tasks.scheduledMessages), len(fakes.tasks.scheduledTransactions))
	}
}

func TestMonoWebHook_HappyPathSchedulesTransaction(t *testing.T) {
	fakes := installAppFakes(t)
	fakes.bot.chatID = 987

	req := httptest.NewRequest(http.MethodPost, "/monoWebHook", strings.NewReader(`{"type":"StatementItem","data":{"account":"WKl9I-LztrH1ZWeafLZEzQ","statementItem":{"id":"test123","time":1746194127,"description":"Bolt","mcc":4121,"originalMcc":4121,"amount":-12000,"operationAmount":-12000,"currencyCode":980,"commissionRate":0,"cashbackAmount":0,"balance":1000000,"hold":false}}}`))
	rr := httptest.NewRecorder()

	MonoWebHook(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	if len(fakes.tasks.scheduledTransactions) != 1 {
		t.Fatalf("scheduled transactions = %d, want 1", len(fakes.tasks.scheduledTransactions))
	}
	got := fakes.tasks.scheduledTransactions[0]
	if got.ChatID != 987 || got.MCC != 4121 || got.Description != "Bolt" || got.Amount != 120 || got.Time != 1746194127 || got.AccountID != "WKl9I-LztrH1ZWeafLZEzQ" {
		t.Fatalf("scheduled transaction = %+v, want parsed Mono transaction", got)
	}
	if got.Category == "" {
		t.Fatalf("scheduled transaction category is empty")
	}
	if got.IsRefund {
		t.Fatalf("scheduled transaction refund = true, want false")
	}
}

func TestMonoWebHook_UnknownMCCSchedulesErrorNotification(t *testing.T) {
	fakes := installAppFakes(t)
	fakes.bot.chatID = 432

	req := httptest.NewRequest(http.MethodPost, "/monoWebHook", strings.NewReader(`{"type":"StatementItem","data":{"account":"a-dnHAO9ExLnboGJP_pdwA","statementItem":{"id":"test123","time":1746194127,"description":"Unknown shop","mcc":999999,"amount":-5000,"balance":1000000,"hold":false}}}`))
	rr := httptest.NewRecorder()

	MonoWebHook(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusInternalServerError)
	}
	if rr.Body.String() != "Could not get category" {
		t.Fatalf("body = %q, want category error", rr.Body.String())
	}
	if len(fakes.tasks.scheduledMessages) != 1 {
		t.Fatalf("scheduled messages = %d, want 1", len(fakes.tasks.scheduledMessages))
	}
	got := fakes.tasks.scheduledMessages[0]
	if got.ChatID != 432 {
		t.Fatalf("scheduled message chat = %d, want 432", got.ChatID)
	}
	if !strings.Contains(got.Text, "999999") {
		t.Fatalf("scheduled message text = %q, want MCC code", got.Text)
	}
	if len(fakes.tasks.scheduledTransactions) != 0 {
		t.Fatalf("scheduled transactions = %d, want 0", len(fakes.tasks.scheduledTransactions))
	}
}
