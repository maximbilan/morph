package morph

import (
	"net/http"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/morph/internal/app"
)

func init() {
	functions.HTTP("handler", handler)
}

func handler(w http.ResponseWriter, r *http.Request) {
	app.Handle(w, r)
}
