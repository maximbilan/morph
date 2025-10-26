#!/bin/bash

FUNCTIONS=("cashHandler" "monoHandler" "monoWebHook" "sendMessage")
RUNTIME="go125"
PROJECT_ID=$MORPH_PROJECT_ID
MEMORY="256MB"
TIMEOUT=180 # 3 minutes

ENV_PARAMS=("MORPH_PROJECT_ID=$MORPH_PROJECT_ID" "MORPH_SERVER_REGION=$MORPH_SERVER_REGION" "CLOUD=true")
ENV_VARS=""
for PARAM in "${ENV_PARAMS[@]}"; do
  ENV_VARS+="$PARAM,"
done
ENV_VARS=${ENV_VARS%,}

SECRET_PARAMS=("MORPH_TELEGRAM_BOT_TOKEN=telegram_bot_token" "MORPH_AI_KEY=ai_key" "MORPH_REDIRECT_KEY=redirect_key" "MORPH_REDIRECT_KEY_2=redirect_key_2" "MORPH_TELEGRAM_CHAT_ID=telegram_chat_id")
SECRETS=""
for PARAM in "${SECRET_PARAMS[@]}"; do
  SECRETS+="$PARAM:latest,"
done
SECRETS=${SECRETS%,}

for FUNC_NAME in "${FUNCTIONS[@]}"; do
    echo "Deploying function: $FUNC_NAME"

    gcloud functions deploy $FUNC_NAME \
        --runtime $RUNTIME \
        --trigger-http \
        --allow-unauthenticated \
        --entry-point $FUNC_NAME \
        --project $PROJECT_ID \
        --gen2 \
        --region $MORPH_SERVER_REGION \
        --set-env-vars $ENV_VARS \
        --set-secrets $SECRETS \
        --memory $MEMORY \
        --timeout $TIMEOUT

    # Print the deployment status
    if [ $? -eq 0 ]; then
        echo "Function $FUNC_NAME deployed successfully."
    else
        echo "Failed to deploy function $FUNC_NAME."
        exit 1
    fi
done