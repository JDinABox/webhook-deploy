#!/bin/sh

if [ ! -d /etc/github-webhook-deploy ]; then
    mkdir -p /etc/github-webhook-deploy
fi

chown -R root:root /etc/github-webhook-deploy
chmod -R 644 /etc/github-webhook-deploy
