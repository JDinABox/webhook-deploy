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
		if guid == "" || requestLogger.GUIDExists(guid) {
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
