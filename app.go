package webhookdeploy

import (
	"encoding/json/v2"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// newApp init fiber app
func newApp(conf *Config) (*chi.Mux, *chi.Mux) {
	r := chi.NewRouter()
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.CleanPath)

	r.Post("/webhook/push", webhookHandler(*conf))

	var webInterface *chi.Mux
	if conf.WebInterface.Enabled {
		webInterface = chi.NewRouter()
		webInterface.Use(middleware.RealIP)
		webInterface.Use(middleware.Logger)
		webInterface.Use(middleware.Recoverer)
		webInterface.Use(middleware.CleanPath)

		webInterface.Group(func(r chi.Router) {
			r.Use(func(next http.Handler) http.Handler {
				return basicAuth(next, conf.WebInterface.Username, conf.WebInterface.Password)
			})
			r.Get("/", webInterfaceHandler(conf))
			r.Post("/deploy", deployHandler(conf))
			r.Post("/logout", logoutHandler())
			r.Handle("/assets/*", http.StripPrefix("/assets", http.FileServer(http.FS(AssetsFs))))

		})
	}

	// 404 error
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.MarshalWrite(w, map[string]any{
			"error":     "404",
			"not-found": r.URL.String(),
		})
	})

	return r, webInterface
}
