#!/bin/sh
if [ -d /run/systemd/system ]; then
    systemctl stop github-webhook-deploy.service
    systemctl disable github-webhook-deploy.service
elif [ -x /sbin/openrc-run ]; then
    rc-service github-webhook-deploy stop
    rc-update del github-webhook-deploy default
fi