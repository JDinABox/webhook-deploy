#!/bin/sh
set -e

# Check if systemd is running
if [ -d /run/systemd/system ]; then
    # Reload the systemd daemon to recognize the new service
    systemctl daemon-reload
    # Enable the service to start on boot
    systemctl enable webhook-deploy.service
    # Start the service
    systemctl start webhook-deploy.service
elif [ -x /sbin/openrc-run ]; then
    # Add service to default runlevel
    rc-update add webhook-deploy default
    # Start the service
    rc-service webhook-deploy start
fi