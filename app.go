package githubwebhookdeploy

import (
	"encoding/json/v2"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Push struct {
	Repository struct {
		FullName string `json:"full_name"`
	} `json:"repository"`
}

// newApp init fiber app
func newApp(conf *Config) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.CleanPath)

	r.Post("/webhook/push", webhookHandler(*conf))

	// 404 error
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.MarshalWrite(w, map[string]any{
			"error":     "404",
			"not-found": r.URL.String(),
		})
	})

	return r
}
