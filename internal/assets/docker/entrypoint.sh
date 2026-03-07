#!/bin/bash
set -e

# Ensure the node user's home directories exist and are writable.
# The .openclaw directory is typically a bind-mount from the host.
mkdir -p /home/node/.vnc /home/node/.openclaw
chown -R node:node /home/node 2>/dev/null || true

# Optional VNC password via env var (defaults to no auth / SecurityTypes None)
if [ -n "$VNC_PASSWORD" ]; then
    echo "$VNC_PASSWORD" | vncpasswd -f > /home/node/.vnc/passwd
    chmod 600 /home/node/.vnc/passwd
    chown node:node /home/node/.vnc/passwd
fi

exec /usr/bin/supervisord -n -c /etc/supervisor/supervisord.conf
