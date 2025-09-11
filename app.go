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
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Push struct {
	Repository struct {
		FullName string `json:"full_name"`
	} `json:"repository"`
}

type guidChecker struct {
	mux   sync.Mutex
	guids map[string]bool
}

func (g *guidChecker) Add(guid string) {
	g.mux.Lock()
	g.guids[guid] = true
	g.mux.Unlock()
}
func (g *guidChecker) Exists(guid string) bool {
	g.mux.Lock()
	defer g.mux.Unlock()
	return g.guids[guid]
}

var guidCheck guidChecker

func init() {
	guidCheck = guidChecker{guids: make(map[string]bool)}
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

		guid := r.Header.Get("X-GitHub-Delivery")
		if guid == "" || guidCheck.Exists(guid) {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		guidCheck.Add(guid)

		log.Printf("[%s] Received valid webhook\n", p.Repository.FullName)

		go func(name, guid string, commands []string) {
			log.Printf("[%s] Deploying...\n", name)
			start := time.Now()

			for _, command := range commands {
				cmd := exec.Command("sh", "-c", command)
				output, err := cmd.CombinedOutput()
				if err != nil {
					log.Printf("Error running deployment command: %s\n%s", err, output)
					return
				}
			}
			log.Printf("[%s] Deployment complete [%.2fs]\n", name, time.Since(start).Seconds())
		}(p.Repository.FullName, guid, deployment.Commands)

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
