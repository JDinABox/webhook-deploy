package githubwebhookdeploy

import (
	"context"
	"crypto/subtle"
	"embed"
	_ "embed"
	"encoding/json/v2"
	"html/template"
	"io/fs"
	"log"
	"net/http"
)

//go:embed templates/index.html
var templateRaw string

//go:embed templates/assets/**
var assetsDir embed.FS

var AssetsFs fs.FS

func init() {
	var err error
	if AssetsFs, err = fs.Sub(assetsDir, "templates/assets"); err != nil {
		log.Fatal(err)
	}
}

var (
	homePageTemplate = template.Must(template.New("home").Parse(templateRaw))
)

const contextKeyAuthUser = "authUser"

func webInterfaceHandler(conf *Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := struct {
			Username    string
			Logs        []string
			Deployments map[string]Deployments
		}{
			Username:    "",
			Logs:        requestLogger.GetLogs(),
			Deployments: conf.Deployments,
		}

		if user, ok := r.Context().Value(contextKeyAuthUser).(string); ok && user != "" {
			data.Username = user
		} else {
			log.Fatalf("No auth user in request context in WebInterfaceHandler")
			return
		}

		if err := homePageTemplate.Execute(w, data); err != nil {
			log.Printf("Error executing template: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}

func deployHandler(conf *Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if user, ok := r.Context().Value(contextKeyAuthUser).(string); !ok && user == "" {
			log.Fatalf("No auth user in request context in deployHandler")
			return
		}
		var payload struct {
			Repository string `json:"repository"`
		}

		if err := json.UnmarshalRead(r.Body, &payload); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		deployment, ok := conf.Deployments[payload.Repository]
		if !ok {
			http.Error(w, "Deployment not found for repository", http.StatusNotFound)
			return
		}

		requestLogger.Logf("[%s] Manual deployment initiated by web interface.", payload.Repository)
		go DeploymentRunner(payload.Repository, deployment.Commands)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Deployment initiated successfully."))
	}
}

func logoutHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
		http.Error(w, "Logged out", http.StatusUnauthorized)
	}
}

func basicAuth(next http.Handler, username, password string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || subtle.ConstantTimeCompare([]byte(user), []byte(username)) != 1 || subtle.ConstantTimeCompare([]byte(pass), []byte(password)) != 1 {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		r = r.WithContext(context.WithValue(r.Context(), contextKeyAuthUser, user))
		next.ServeHTTP(w, r)
	})
}
