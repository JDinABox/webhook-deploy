# webhook-deploy

A webhook server to automatically deploy GitHub repositories

## Using

[Config](#config)

### Running

#### Docker Compose (Recommended)

Clone Repo

```sh
git clone https://github.com/JDinABox/webhook-deploy.git
cd webhook-deploy
```

Generate ssh key or add and existing one to the container

Edit `docker-compose.yaml`
Edit [Config](#config)

Deploy

```sh
docker-compose up --build -d
```

#### Manually

Edit [Config](#config)

```sh
CONFIG="path/to/config.yaml" webhook-deploy
```

### Endpoints

- `/webhook/push` - Trigger a deployment

- Web Interface [127.0.0.1:9080](127.0.0.1:9080)
  ![Web interface img](./docs/assets/web-interface.jpg "Web interface")

## Building

Go version: 1.25+

### With Docker (Recommended)

```sh
git clone https://github.com/JDinABox/webhook-deploy.git
cd webhook-deploy
docker build .
```

Packages and binaries will be in the `dist` directory.

### With Go

```sh
git clone https://github.com/JDinABox/webhook-deploy.git
cd webhook-deploy
GOEXPERIMENT=jsonv2 go build -o ./webhook-deploy ./cmd/webhook-deploy/
```

### Rebuilding web interface assets (Optional: Included in repo)

[TailwindCSS CLI](https://tailwindcss.com/docs/installation/tailwind-cli)

```sh
tailwindcss -m -o ./templates/assets/output.css
```

## Config

See [`config/config.yaml`](./config/config.yaml)

Installed at `/etc/webhook-deploy/config.yaml`

```yaml
listen: ":80" # Public port to listen on
ssh-known-hosts: "./path/to/known_hosts"
web-interface: # Web interface should not be exposed to the internet
  enabled: true
  listen: "127.0.0.1:9080"
  username: "admin"
  password: "password"
deployments:
  user/repo:
    remote:
        user: "user"
        server_ip: "server1"
        private_key: "./path/to/private/key" # SSH private key
    secret: "secret" # GitHub webhook secret
    commands:
      - /path/to/script.sh
      - /path/to/other/script.sh
   org/repo2:
      remote:
        user: "user"
        server_ip: "server2"
        private_key: "./path/to/private/key"
      secret: "secret2"
      commands:
      - cd /path/to/project; git pull; docker-compose up --build -d
```

### Github setup

On `https://github.com/user/repo/settings/hooks/new`:

- Payload URL - `http(s)://{DOMAIN/IP}:{PORT}/webhook/push`
- Content type - `application/json`
- Secret - `deployments` > `user/repo` > `secret`
- Which events would you like to trigger this webhook? - Just the `push` event
- Active - `Checked`

![Github webhook setup](./docs/assets/webhook-config.jpg "Github webhook setup")

## Developing

Tools:

- Go 1.25+
  ```sh
  GOEXPERIMENT=jsonv2 go run ./cmd/webhook-deploy/ ./config/config.yaml
  ```
- [TailwindCSS CLI](https://tailwindcss.com/docs/installation/tailwind-cli) - Web interface styles
  ```sh
  tailwindcss -w -m -o ./templates/assets/output.css
  ```
- [smee](https://github.com/probot/smee-client) (Optional but Recommended) - Webhook proxy
  Start `webhook-deploy`
  ```sh
  smee --url https://smee.io/{URL} --path /webhook/push --port 8080
  ```
