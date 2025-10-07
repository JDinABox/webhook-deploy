#!/bin/sh

if [ ! -d /etc/webhook-deploy ]; then
    mkdir -p /etc/webhook-deploy
fi

chown -R root:root /etc/webhook-deploy
chmod -R 644 /etc/webhook-deploy
