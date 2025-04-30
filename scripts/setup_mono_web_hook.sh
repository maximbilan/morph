#!/bin/bash

if [ -z "$MORPH_MONO_API_KEY" ]; then
  echo "Error: MORPH_MONO_API_KEY is not set."
  exit 1
fi

if [ -z "$MORPH_MONO_WEBHOOK_URL" ]; then
  echo "Error: MORPH_MONO_WEBHOOK_URL is not set."
  exit 1
fi

#
# Set the Monobank webhook
curl -X POST https://api.monobank.ua/personal/webhook \
     -H "X-Token: $MORPH_MONO_API_KEY" \
     -H "Content-Type: application/json" \
     -d "{\"webHookUrl\": \"$MORPH_MONO_WEBHOOK_URL\"}"

# Check the response
if [ $? -eq 0 ]; then
  echo "Webhook has been set successfully."
else
  echo "Failed to set webhook."
fi