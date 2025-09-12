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

	configPath := "/etc/github-webhook-deploy/config.yaml"

	if len(os.Args) < 2 {
		log.Println("Warning: Missing configuration file input")
	} else {
		configPath = strings.TrimSpace(os.Args[1])
	}

	configs = append(configs, webhookdeploy.WithConfigFile(configPath))

	log.Fatal(webhookdeploy.Start(configs...))
}
