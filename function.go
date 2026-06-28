package morph

import (
	"net/http"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/morph/internal/app"
)

func init() {
	functions.HTTP("cashHandler", cashHandler)
	functions.HTTP("monoHandler", monoHandler)
	functions.HTTP("monoWebHook", monoWebHook)
	functions.HTTP("sendMessage", sendMessage)
	functions.HTTP("notificationHandler", notificationHandler)
}

func cashHandler(w http.ResponseWriter, r *http.Request) {
	app.CashHandler(w, r)
}

func monoHandler(w http.ResponseWriter, r *http.Request) {
	app.MonoHandler(w, r)
}

func monoWebHook(w http.ResponseWriter, r *http.Request) {
	app.MonoWebHook(w, r)
}

func sendMessage(w http.ResponseWriter, r *http.Request) {
	app.SendMessage(w, r)
}

func notificationHandler(w http.ResponseWriter, r *http.Request) {
	app.NotificationHandler(w, r)
}
