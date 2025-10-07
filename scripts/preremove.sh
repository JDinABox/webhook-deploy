#!/bin/sh
if [ -d /run/systemd/system ]; then
    systemctl stop webhook-deploy.service
    systemctl disable webhook-deploy.service
elif [ -x /sbin/openrc-run ]; then
    rc-service webhook-deploy stop
    rc-update del webhook-deploy default
fi