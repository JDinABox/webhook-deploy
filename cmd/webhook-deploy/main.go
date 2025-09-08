// / 2>/dev/null ; gorun "$0" "$@" ; exit $?

package main

import (
	"log"
	"os"
	"strings"

	webhookdeploy "github.com/JDinABox/github-webhook-deploy"
)

func main() {
	configs := []webhookdeploy.Option{}

	if listenAddr := strings.TrimSpace(os.Getenv("LISTEN")); listenAddr != "" {
		configs = append(configs, webhookdeploy.WithListenAddr(listenAddr))
	}
	log.Fatal(webhookdeploy.Start(configs...))
}
