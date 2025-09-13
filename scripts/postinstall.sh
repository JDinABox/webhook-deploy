#!/bin/sh
set -e

# Check if systemd is running
if [ -d /run/systemd/system ]; then
    # Reload the systemd daemon to recognize the new service
    systemctl daemon-reload
    # Enable the service to start on boot
    systemctl enable github-webhook-deploy.service
    # Start the service
    systemctl start github-webhook-deploy.service
elif [ -x /sbin/openrc-run ]; then
    # Add service to default runlevel
    rc-update add github-webhook-deploy default
    # Start the service
    rc-service github-webhook-deploy start
fi