package githubwebhookdeploy

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json/v2"
	"net/http"
	"os/exec"
	"time"
)

type Push struct {
	Repository struct {
		FullName string `json:"full_name"`
	} `json:"repository"`
}

func webhookHandler(conf Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		requestLogger.Logf("[%s] Received webhook request\n", p.Repository.FullName)

		deployment, ok := conf.Deployments[p.Repository.FullName]
		if !ok {
			requestLogger.Logf("[%s] Request not implemented\n", p.Repository.FullName)
			w.WriteHeader(http.StatusNotImplemented)
			return
		}

		h := hmac.New(sha256.New, []byte(deployment.Secret))
		h.Write(body)

		webSig := r.Header.Get("X-Hub-Signature-256")

		if !hmac.Equal([]byte("sha256="+hex.EncodeToString(h.Sum(nil))), []byte(webSig)) {
			requestLogger.Logf("[%s] Invalid signature\n", p.Repository.FullName)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		guid := r.Header.Get("X-GitHub-Delivery")
		if guid == "" || requestLogger.GUIDExists(guid) {
			requestLogger.Logf("[%s] GUID already used\n", p.Repository.FullName)
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		requestLogger.AddGUID(guid)

		requestLogger.Logf("[%s] Received valid webhook\n", p.Repository.FullName)

		go DeploymentRunner(p.Repository.FullName, deployment.Commands)

		w.WriteHeader(http.StatusNoContent)
	}
}

func DeploymentRunner(name string, commands []string) {
	requestLogger.Logf("[%s] Deploying...\n", name)
	start := time.Now()

	for _, command := range commands {
		cmd := exec.Command("sh", "-c", command)
		output, err := cmd.CombinedOutput()
		if err != nil {
			requestLogger.Logf("Error running deployment command: %s\n%s", err, output)
			requestLogger.Logf("[%s] Deployment Failed [%.2fs]\n", name, time.Since(start).Seconds())
			return
		}
	}
	requestLogger.Logf("[%s] Deployment complete [%.2fs]\n", name, time.Since(start).Seconds())
}
