package githubwebhookdeploy

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"os/exec"
	"time"

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

		var p Push
		json.Unmarshal(body, &p)

		deployment, ok := conf.Deployments[p.Repository.FullName]
		if !ok {
			w.WriteHeader(http.StatusNotImplemented)
			return
		}

		h := hmac.New(sha256.New, []byte(deployment.Secret))
		h.Write(body)

		webSig := r.Header.Get("X-Hub-Signature-256")

		if !hmac.Equal([]byte("sha256="+hex.EncodeToString(h.Sum(nil))), []byte(webSig)) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		log.Printf("[%s] Deploying...\n", p.Repository.FullName)
		start := time.Now()

		// Normal deployment logic
		for _, command := range deployment.Commands {
			cmd := exec.Command("sh", "-c", command)
			output, err := cmd.CombinedOutput()
			if err != nil {
				log.Printf("Error running deployment command: %s\n%s", err, output)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		log.Printf("[%s] Deployment complete [%.2fs]\n", p.Repository.FullName, time.Since(start).Seconds())

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
