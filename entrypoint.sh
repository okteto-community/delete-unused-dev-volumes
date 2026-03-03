#!/bin/bash
set -e

echo "Downloading latest Okteto CLI..."

# Download and install the latest Okteto CLI
curl -fsSL https://get.okteto.com -o /tmp/install-okteto.sh
chmod +x /tmp/install-okteto.sh
/tmp/install-okteto.sh

echo "Okteto CLI installed successfully"
okteto version

# Run the main application
echo "Starting application..."
exec /usr/local/bin/app
