package githubwebhookdeploy

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log"
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

	r.Post("/webhook/push", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-GitHub-Event") != "push" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		var buf bytes.Buffer
		buf.ReadFrom(r.Body)
		defer r.Body.Close()

		body := buf.Bytes()

		h := hmac.New(sha256.New, []byte(conf.Secret))
		h.Write(body)

		webSig := r.Header.Get("X-Hub-Signature-256")

		if !hmac.Equal([]byte("sha256="+hex.EncodeToString(h.Sum(nil))), []byte(webSig)) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		var p Push
		json.Unmarshal(body, &p)

		deployments, ok := conf.Deployments[p.Repository.FullName]
		if !ok {
			w.WriteHeader(http.StatusNotImplemented)
			return
		}

		log.Printf("%+v\n", deployments)
		w.WriteHeader(http.StatusNoContent)
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
