package app

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/morph/internal/aiservice"
)

func TestNotificationHandler_InvalidJSONReturnsBadRequest(t *testing.T) {
	fakes := installAppFakes(t)

	req := httptest.NewRequest(http.MethodPost, "/notificationHandler", strings.NewReader(`{invalid`))
	rr := httptest.NewRecorder()

	NotificationHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
	if rr.Body.String() != "Could not parse notification" {
		t.Fatalf("body = %q, want parse error", rr.Body.String())
	}
	if fakes.tasks.connectCount != 0 {
		t.Fatalf("task service connected %d times, want 0", fakes.tasks.connectCount)
	}
}

func TestNotificationHandler_EmptyContentReturnsOK(t *testing.T) {
	fakes := installAppFakes(t)

	req := httptest.NewRequest(http.MethodPost, "/notificationHandler", strings.NewReader(`{"app":"BBVA"}`))
	rr := httptest.NewRecorder()

	NotificationHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	if fakes.tasks.connectCount != 0 {
		t.Fatalf("task service connected %d times, want 0", fakes.tasks.connectCount)
	}
	if fakes.ai.callCount != 0 {
		t.Fatalf("AI called %d times, want 0", fakes.ai.callCount)
	}
}

func TestNotificationHandler_ChatIDFailureReturnsInternalServerError(t *testing.T) {
	fakes := installAppFakes(t)
	fakes.bot.chatIDErr = errors.New("missing chat id")

	req := httptest.NewRequest(http.MethodPost, "/notificationHandler", strings.NewReader(`{"app":"BBVA","title":"Recibo cargado","message":"79,81 EUR"}`))
	rr := httptest.NewRecorder()

	NotificationHandler(rr, req)

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

func TestNotificationHandler_NoAIResponseSchedulesErrorMessage(t *testing.T) {
	fakes := installAppFakes(t)
	fakes.bot.chatID = 909

	req := httptest.NewRequest(http.MethodPost, "/notificationHandler", strings.NewReader(`{"app":"ПУМБ","title":"Списання","message":"167.36UAH"}`))
	rr := httptest.NewRecorder()

	NotificationHandler(rr, req)

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
	if got.ChatID != 909 || got.Text != "No response from AI" || got.ReplyToMessageID != nil {
		t.Fatalf("scheduled message = %+v, want notification AI error", got)
	}
	if fakes.shortURL.callCount != 0 {
		t.Fatalf("short URL calls = %d, want 0", fakes.shortURL.callCount)
	}
}

func TestNotificationHandler_HappyPathSchedulesShortLink(t *testing.T) {
	fakes := installAppFakes(t)
	fakes.bot.chatID = 777
	fakes.ai.response = &aiservice.Response{
		Category:    "Bills",
		Subcategory: "Utilities",
		Amount:      -79.81,
	}

	req := httptest.NewRequest(http.MethodPost, "/notificationHandler", strings.NewReader(`{"app":"BBVA ES","title":"Recibo cargado","message":"Se ha cargado en tu cuenta *3297 un adeudo de AIGUES MUNICIPALS DE PATERNA, S.A. de 79,81 EUR.","date":"2026-06-26T13:13:00+03:00"}`))
	rr := httptest.NewRecorder()

	NotificationHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	if fakes.ai.callCount != 1 {
		t.Fatalf("AI calls = %d, want 1", fakes.ai.callCount)
	}
	if !strings.Contains(fakes.ai.userPrompt, "BBVA ES") || !strings.Contains(fakes.ai.userPrompt, "Recibo cargado") {
		t.Fatalf("user prompt = %q, want app and title", fakes.ai.userPrompt)
	}
	if len(fakes.deepLink.calls) != 1 {
		t.Fatalf("deep link calls = %d, want 1", len(fakes.deepLink.calls))
	}
	deepLink := fakes.deepLink.calls[0]
	// No app:account mapping configured yet, so the app name is used as the account.
	if deepLink.account != "BBVA ES" {
		t.Fatalf("deep link account = %q, want %q", deepLink.account, "BBVA ES")
	}
	if deepLink.amount != 79.81 {
		t.Fatalf("deep link amount = %.2f, want 79.81", deepLink.amount)
	}
	wantDate := time.Date(2026, time.June, 26, 13, 13, 0, 0, time.FixedZone("", 3*60*60))
	if !deepLink.date.Equal(wantDate) {
		t.Fatalf("deep link date = %s, want %s", deepLink.date, wantDate)
	}
	if len(fakes.shortURL.inputs) != 1 || fakes.shortURL.inputs[0] != "moneywiz://expense" {
		t.Fatalf("short URL inputs = %v, want moneywiz deep link", fakes.shortURL.inputs)
	}
	if len(fakes.tasks.scheduledMessages) != 1 {
		t.Fatalf("scheduled messages = %d, want 1", len(fakes.tasks.scheduledMessages))
	}
	got := fakes.tasks.scheduledMessages[0]
	wantText := "📲 BBVA ES\nCategory: Bills\nSubcategory: Utilities\nAmount: 79.81\nhttps://short.example/link"
	if got.Text != wantText {
		t.Fatalf("scheduled text = %q, want %q", got.Text, wantText)
	}
	if got.ChatID != 777 || got.ReplyToMessageID != nil {
		t.Fatalf("scheduled message = %+v, want chat 777 with no reply", got)
	}
}

func TestNotificationHandler_ShortURLErrorIsIncludedInScheduledMessage(t *testing.T) {
	fakes := installAppFakes(t)
	fakes.ai.response = &aiservice.Response{
		Category:    "Multimedia",
		Subcategory: "Applications",
		Amount:      149,
	}
	fakes.shortURL.err = errors.New("shortener down")

	req := httptest.NewRequest(http.MethodPost, "/notificationHandler", strings.NewReader(`{"app":"Privat24","title":"Privat24","message":"-149₴ Цифрові товари. YouTube Premium"}`))
	rr := httptest.NewRecorder()

	NotificationHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	if len(fakes.tasks.scheduledMessages) != 1 {
		t.Fatalf("scheduled messages = %d, want 1", len(fakes.tasks.scheduledMessages))
	}
	if !strings.Contains(fakes.tasks.scheduledMessages[0].Text, "Error shortening URL: shortener down") {
		t.Fatalf("scheduled text = %q, want shortening error", fakes.tasks.scheduledMessages[0].Text)
	}
}

func TestGetAccountNameFromApp(t *testing.T) {
	if got := getAccountNameFromApp("Unknown Bank"); got != "Unknown Bank" {
		t.Fatalf("unknown app account = %q, want app name fallback", got)
	}

	oldMap := appAccountMap
	appAccountMap = map[string]string{"BBVA ES": "BBVAEUR"}
	t.Cleanup(func() { appAccountMap = oldMap })

	if got := getAccountNameFromApp("BBVA ES"); got != "BBVAEUR" {
		t.Fatalf("mapped app account = %q, want BBVAEUR", got)
	}
	if got := getAccountNameFromApp("ПУМБ"); got != "ПУМБ" {
		t.Fatalf("unmapped app account = %q, want app name fallback", got)
	}
}

func TestParseNotificationDate(t *testing.T) {
	kyiv, err := time.LoadLocation("Europe/Kyiv")
	if err != nil {
		t.Fatalf("load Kyiv location: %v", err)
	}

	t.Run("RFC3339 absolute instant", func(t *testing.T) {
		want := time.Date(2026, time.June, 26, 13, 13, 0, 0, time.FixedZone("", 3*60*60))
		if got := parseNotificationDate("2026-06-26T13:13:00+03:00"); !got.Equal(want) {
			t.Fatalf("got %s, want %s", got, want)
		}
	})

	t.Run("unix epoch seconds", func(t *testing.T) {
		if got := parseNotificationDate("1746194127"); !got.Equal(time.Unix(1746194127, 0)) {
			t.Fatalf("got %s, want unix 1746194127", got)
		}
	})

	t.Run("unix epoch milliseconds", func(t *testing.T) {
		if got := parseNotificationDate("1746194127000"); !got.Equal(time.Unix(1746194127, 0)) {
			t.Fatalf("got %s, want unix 1746194127", got)
		}
	})

	t.Run("naive notification text in Kyiv time", func(t *testing.T) {
		want := time.Date(2026, time.June, 26, 13, 13, 0, 0, kyiv)
		if got := parseNotificationDate("26-06-2026 13:13"); !got.Equal(want) {
			t.Fatalf("got %s, want %s", got, want)
		}
	})

	t.Run("empty falls back to now", func(t *testing.T) {
		before := time.Now()
		got := parseNotificationDate("")
		if got.Before(before) || got.After(time.Now()) {
			t.Fatalf("got %s, want approximately now", got)
		}
	})

	t.Run("garbage falls back to now", func(t *testing.T) {
		before := time.Now()
		got := parseNotificationDate("not a date")
		if got.Before(before) || got.After(time.Now()) {
			t.Fatalf("got %s, want approximately now", got)
		}
	})
}
