#!/bin/bash

# Check if TELEGRAM_TOKEN and CLOUD_FUNCTION_URL are set
if [ -z "$MORPH_TELEGRAM_BOT_TOKEN" ]; then
  echo "Error: MORPH_TELEGRAM_BOT_TOKEN is not set."
  exit 1
fi

if [ -z "$MORPH_CLOUD_FUNCTION_URL" ]; then
  echo "Error: MORPH_CLOUD_FUNCTION_URL is not set."
  exit 1
fi

# Set the webhook
curl --data "url=$MORPH_CLOUD_FUNCTION_URL" https://api.telegram.org/bot$MORPH_TELEGRAM_BOT_TOKEN/setWebhook

# Check the response
if [ $? -eq 0 ]; then
  echo "Webhook has been set successfully."
else
  echo "Failed to set webhook."
fi