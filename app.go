package githubwebhookdeploy

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// newApp init fiber app
func newApp() *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.CleanPath)

	r.Post("/webhook/push", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%v\n", r.Header)
		var j map[string]any
		json.NewDecoder(r.Body).Decode(&j)
		defer r.Body.Close()
		js, _ := json.Marshal(j)
		log.Printf("%v\n", string(js))
		w.WriteHeader(http.StatusOK)
	})

	// 404 error
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(
			map[string]any{
				"error":     "404",
				"not-found": r.URL.String(),
			})
	})

	return r
}
