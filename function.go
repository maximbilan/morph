package morph

import (
	"net/http"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/morph/internal/app"
)

func init() {
	functions.HTTP("handler", handler)
	functions.HTTP("monoWebHook", monoWebHook)
}

func handler(w http.ResponseWriter, r *http.Request) {
	app.Handle(w, r)
}

func monoWebHook(w http.ResponseWriter, r *http.Request) {
	app.MonoWebHook(w, r)
}
