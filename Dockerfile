
FROM golang:alpine AS builder-go
WORKDIR /go/src/github.com/JDinABox/webhook-deploy
COPY go.* ./
RUN go mod download
COPY ./cmd ./cmd
COPY ./templates ./templates
COPY *.go ./
ENV GOEXPERIMENT=jsonv2
RUN --mount=type=cache,target=/root/.cache/go-build go build -ldflags="-s -w" -o ./cmd/webhook-deploy/webhook-deploy.so ./cmd/webhook-deploy


FROM alpine:latest

RUN apk --no-cache -U upgrade \
    && apk --no-cache add --upgrade ca-certificates openssh-client \
    && ARCH=$(uname -m) && wget -O /bin/dumb-init https://github.com/Yelp/dumb-init/releases/download/v1.2.5/dumb-init_1.2.5_${ARCH} \
    && chmod +x /bin/dumb-init

COPY --from=builder-go /go/src/github.com/JDinABox/webhook-deploy/cmd/webhook-deploy/webhook-deploy.so /bin/webhook-deploy.so

ENTRYPOINT ["/bin/dumb-init", "--"]
CMD ["/bin/webhook-deploy.so"]