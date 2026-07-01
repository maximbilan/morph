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
		Category:      "Bills",
		Subcategory:   "Utilities",
		Amount:        -79.81,
		IsTransaction: true,
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
	// BBVA always maps to a single MoneyWiz account, ignoring the account in the message.
	if deepLink.account != "BBVAEur" {
		t.Fatalf("deep link account = %q, want %q", deepLink.account, "BBVAEur")
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

func TestNotificationHandler_ShortURLErrorFallsBackToRawDeepLink(t *testing.T) {
	fakes := installAppFakes(t)
	fakes.ai.response = &aiservice.Response{
		Category:      "Multimedia",
		Subcategory:   "Applications",
		Amount:        149,
		IsTransaction: true,
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
	got := fakes.tasks.scheduledMessages[0].Text
	if !strings.Contains(got, "moneywiz://expense") {
		t.Fatalf("scheduled text = %q, want raw deep link fallback", got)
	}
	if strings.Contains(got, "shortener down") {
		t.Fatalf("scheduled text = %q, should not leak the shortener error", got)
	}
}

func TestNotificationHandler_NonTransactionIsIgnored(t *testing.T) {
	fakes := installAppFakes(t)
	fakes.ai.response = &aiservice.Response{
		Category:      "Other",
		Subcategory:   "",
		Amount:        0,
		IsTransaction: false,
	}

	req := httptest.NewRequest(http.MethodPost, "/notificationHandler", strings.NewReader(`{"app":"Privat24","title":"Privat24","message":"🎉 Отримайте 5% кешбек цими вихідними!"}`))
	rr := httptest.NewRecorder()

	NotificationHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	if fakes.tasks.connectCount != 1 || fakes.tasks.closeCount != 1 {
		t.Fatalf("task connect/close = %d/%d, want 1/1", fakes.tasks.connectCount, fakes.tasks.closeCount)
	}
	if fakes.deepLink.callCount != 0 {
		t.Fatalf("deep link calls = %d, want 0", fakes.deepLink.callCount)
	}
	if fakes.shortURL.callCount != 0 {
		t.Fatalf("short URL calls = %d, want 0", fakes.shortURL.callCount)
	}
	if len(fakes.tasks.scheduledMessages) != 0 {
		t.Fatalf("scheduled messages = %d, want 0", len(fakes.tasks.scheduledMessages))
	}
}

func TestResolveAccountName(t *testing.T) {
	tests := []struct {
		name    string
		app     string
		message string
		want    string
	}{
		{
			name:    "BBVA ignores account number in message",
			app:     "BBVA ES",
			message: "Se ha cargado en tu cuenta *3297 un adeudo de 79,81 EUR.",
			want:    "BBVAEur",
		},
		{
			name:    "PUMB UAH platinum from masked account",
			app:     "ПУМБ",
			message: "167.36UAH\n26-06-2026 13:13\nРахунок: *0451\nДоступно: 3155.99UAH",
			want:    "PumbUAHPlatinum",
		},
		{
			name:    "PUMB UAH platinum from secondary card token",
			app:     "ПУМБ",
			message: "25.60EUR / 1319.18UAH (курс 51.53)\nGASTROBAR B13 VALENCIA ES\n30-06-2026 12:46\nКартка: *5353...",
			want:    "PumbUAHPlatinum",
		},
		{
			name:    "PUMB USD from masked account",
			app:     "Pumb",
			message: "Рахунок: *5381\nДоступно: 100.00USD",
			want:    "PumbUSD",
		},
		{
			name:    "Privat24 online UAH from token",
			app:     "Privat24",
			message: "-149₴ Цифрові товари. YouTube Premium\n5*85 22:37\nБал. 429.4₴",
			want:    "PrivatOnlineUAH",
		},
		{
			name:    "Privat24 EUR from token",
			app:     "Privat24",
			message: "-10€ Some payment\n1*89 09:00\nБал. 100€",
			want:    "PrivatEUR",
		},
		{
			name:    "known bank but unrecognized account falls back to app name",
			app:     "Pumb",
			message: "Рахунок: *9999\nДоступно: 1.00UAH",
			want:    "Pumb",
		},
		{
			name:    "unknown app falls back to app name",
			app:     "Some Bank",
			message: "transaction *0451",
			want:    "Some Bank",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resolveAccountName(tt.app, tt.message); got != tt.want {
				t.Fatalf("resolveAccountName(%q, ...) = %q, want %q", tt.app, got, tt.want)
			}
		})
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
