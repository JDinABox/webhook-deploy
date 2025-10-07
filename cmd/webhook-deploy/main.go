// / 2>/dev/null ; gorun "$0" "$@" ; exit $?

package main

import (
	"log"
	"os"

	webhookdeploy "github.com/JDinABox/webhook-deploy"
)

func main() {
	configs := []webhookdeploy.Option{}

	configPath := "/etc/webhook-deploy/config.yaml"

	envConfigPath := os.Getenv("CONFIG")
	if envConfigPath != "" {
		configPath = envConfigPath
	} else {
		log.Println("Warn: Missing configuration file input. Using default /etc/webhook-deploy/config.yaml")
	}

	configs = append(configs, webhookdeploy.WithConfigFile(configPath))

	log.Fatal(webhookdeploy.Start(configs...))
}
