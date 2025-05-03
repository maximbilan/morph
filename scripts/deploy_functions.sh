#!/bin/bash

HANDLER_FUNC_NAME="cashHandler"
MONOWEBHOOK_FUNC_NAME="monoWebHook"

# Set the runtime
RUNTIME="go122"
# Set the project ID
PROJECT_ID=$MORPH_PROJECT_ID

# Set environment variables
ENV_PARAMS=("MORPH_PROJECT_ID=$MORPH_PROJECT_ID" "MORPH_SERVER_REGION=$MORPH_SERVER_REGION" "CLOUD=true")
ENV_VARS=""
for PARAM in "${ENV_PARAMS[@]}"; do
  ENV_VARS+="$PARAM,"
done
ENV_VARS=${ENV_VARS%,}

# Set the secret environment variables
SECRET_PARAMS=("MORPH_TELEGRAM_BOT_TOKEN=telegram_bot_token" "MORPH_AI_KEY=ai_key" "MORPH_REDIRECT_KEY=redirect_key" "MORPH_TELEGRAM_CHAT_ID=telegram_chat_id")
SECRETS=""
for PARAM in "${SECRET_PARAMS[@]}"; do
  SECRETS+="$PARAM:latest,"
done
SECRETS=${SECRETS%,}

# Set memory parameter
MEMORY="256MB"

# Timeout
WEB_HOOK_TIMEOUT=300

# Deploy the cash handler function
gcloud functions deploy $HANDLER_FUNC_NAME \
    --runtime $RUNTIME \
    --trigger-http \
    --allow-unauthenticated \
    --entry-point $HANDLER_FUNC_NAME \
    --project $PROJECT_ID \
    --gen2 \
    --region $MORPH_SERVER_REGION \
    --set-env-vars $ENV_VARS \
    --set-secrets $SECRETS \
    --memory $MEMORY

# Print the deployment status
if [ $? -eq 0 ]; then
    echo "Function $HANDLER_FUNC_NAME deployed successfully."
else
    echo "Failed to deploy function $HANDLER_FUNC_NAME."
fi

# Deploy Mono Web Hook function
gcloud functions deploy $MONOWEBHOOK_FUNC_NAME \
    --runtime $RUNTIME \
    --trigger-http \
    --allow-unauthenticated \
    --entry-point $MONOWEBHOOK_FUNC_NAME \
    --project $PROJECT_ID \
    --gen2 \
    --region $MORPH_SERVER_REGION \
    --set-env-vars $ENV_VARS \
    --set-secrets $SECRETS \
    --memory $MEMORY \
    --timeout $WEB_HOOK_TIMEOUT

# Print the deployment status
if [ $? -eq 0 ]; then
    echo "Function $MONOWEBHOOK_FUNC_NAME deployed successfully."
else
    echo "Failed to deploy function $MONOWEBHOOK_FUNC_NAME."
fi
