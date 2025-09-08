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

	var configPath string

	if len(os.Args) < 2 {
		log.Fatal("Missing configuration file input")
	}
	configPath = strings.TrimSpace(os.Args[1])

	configs = append(configs, webhookdeploy.WithConfigFile(configPath))

	log.Fatal(webhookdeploy.Start(configs...))
}
