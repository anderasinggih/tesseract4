#!/bin/bash
set -e

APP_DIR="/www/wwwroot/tesseract4"
BINARY_NAME="tesseract"
SERVICE_NAME="tesseract"

echo "🚀 [Deploy] Navigating to $APP_DIR..."
cd $APP_DIR

echo "📥 [Deploy] Pulling latest code from GitHub..."
git pull origin main

echo "📦 [Deploy] Downloading Go dependencies..."
/usr/local/go/bin/go mod tidy

echo "🔨 [Deploy] Building Go binary..."
/usr/local/go/bin/go build -o $BINARY_NAME .

echo "🔄 [Deploy] Restarting systemd service..."
sudo systemctl restart $SERVICE_NAME

echo "✅ [Deploy] Deployment successful! Service status:"
sudo systemctl status $SERVICE_NAME --no-pager
