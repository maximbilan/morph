# Morph - Financial Transaction Management System

Morph is a serverless application built on Google Cloud Functions that helps manage and categorize financial transactions. The system integrates with Monobank for transaction monitoring and uses AI to automatically categorize expenses. The primary output of the system is deep links for the MoneyWiz app, allowing users to quickly add categorized expenses to their financial tracking.

## Features

- 🤖 **AI-Powered Categorization**: Automatically categorizes transactions using AI
- 🏦 **Monobank Integration**: Real-time transaction monitoring via webhooks
- 💰 **MoneyWiz Deep Links**: Generates deep links for quick expense entry
- 📱 **Telegram Integration**: Sends transaction notifications via Telegram bot
- ⚡ **Asynchronous Processing**: Uses Google Cloud Tasks for reliable message delivery
- 🔗 **URL Shortening**: Integrates with Short.io for compact deep links

## Prerequisites

- Go 1.25.0 or later
- Google Cloud Platform account with billing enabled
- Google Cloud SDK (`gcloud`) installed and configured
- Monobank API token
- Telegram Bot Token
- OpenAI API key (for AI categorization)
- Short.io API key (for URL shortening)

## Project Structure

```
morph/
├── cmd/                    # Main application entry point
├── internal/              # Internal application packages
│   ├── app/              # HTTP handlers and application logic
│   ├── aiservice/        # AI service integration
│   ├── botservice/       # Bot service logic
│   ├── category/         # Category management
│   ├── deeplinkgenerator/# MoneyWiz deep link generation
│   ├── shorturl/         # URL shortening service
│   └── taskservice/      # Google Cloud Tasks integration
├── third_party/          # Third-party service integrations
│   ├── googletasks/      # Google Cloud Tasks client
│   ├── moneywiz/         # MoneyWiz deep link generator
│   ├── mono/             # Monobank API client
│   ├── openai/           # OpenAI API client
│   ├── shortio/          # Short.io API client
│   └── telegram/         # Telegram Bot API client
├── scripts/              # Deployment and setup scripts
└── .github/workflows/    # CI/CD workflows
```

## Configuration

### Environment Variables

The following environment variables are required for the application to function:

#### Required Secrets (stored in Google Secret Manager)
- `MORPH_TELEGRAM_BOT_TOKEN`: Telegram bot token
- `MORPH_AI_KEY`: OpenAI API key
- `MORPH_REDIRECT_KEY`: Short.io API key
- `MORPH_TELEGRAM_CHAT_ID`: Telegram chat ID for notifications

#### Required Environment Variables
- `MORPH_PROJECT_ID`: Google Cloud Project ID
- `MORPH_SERVER_REGION`: Google Cloud region (e.g., `us-central1`)

#### Additional Setup Variables
- `MORPH_MONO_API_KEY`: Monobank API token (for webhook setup)
- `MORPH_MONO_WEBHOOK_URL`: Monobank webhook URL (for webhook setup)
- `MORPH_CLOUD_FUNCTION_URL`: Cloud Function URL (for Telegram webhook setup)

## Setup

### 1. Clone the Repository

```bash
git clone https://github.com/yourusername/morph.git
cd morph
```

### 2. Install Dependencies

```bash
go mod download
```

### 3. Set Up Google Cloud

1. Create a new Google Cloud project or use an existing one
2. Enable the following APIs:
   - Cloud Functions API
   - Cloud Tasks API
   - Secret Manager API
3. Set up authentication:
   ```bash
   ./scripts/setup_gcloud_access.sh
   ```

### 4. Configure Secrets

Store your secrets in Google Secret Manager:

```bash
# Store Telegram bot token
echo -n "YOUR_TELEGRAM_BOT_TOKEN" | gcloud secrets create telegram_bot_token --data-file=-

# Store OpenAI API key
echo -n "YOUR_OPENAI_API_KEY" | gcloud secrets create ai_key --data-file=-

# Store Short.io API key
echo -n "YOUR_SHORTIO_API_KEY" | gcloud secrets create redirect_key --data-file=-

# Store Telegram chat ID
echo -n "YOUR_CHAT_ID" | gcloud secrets create telegram_chat_id --data-file=-
```

### 5. Deploy Cloud Functions

```bash
export MORPH_PROJECT_ID="your-project-id"
export MORPH_SERVER_REGION="us-central1"
./scripts/deploy_functions.sh
```

### 6. Set Up Webhooks

#### Monobank Webhook
```bash
export MORPH_MONO_API_KEY="your-monobank-api-key"
export MORPH_MONO_WEBHOOK_URL="https://YOUR_REGION-YOUR_PROJECT.cloudfunctions.net/monoWebHook"
./scripts/setup_mono_web_hook.sh
```

#### Telegram Bot Webhook
```bash
export MORPH_TELEGRAM_BOT_TOKEN="your-telegram-bot-token"
export MORPH_CLOUD_FUNCTION_URL="https://YOUR_REGION-YOUR_PROJECT.cloudfunctions.net/YOUR_FUNCTION"
./scripts/setup_telegram_bot.sh
```

## Architecture Overview

The application consists of five main Google Cloud Functions that work together to process and manage financial transactions:

### 1. `cashHandler`
- **Purpose**: Processes manual cash transactions entered by users via Telegram
- **Flow**:
  1. Receives transaction details via HTTP request
  2. Uses AI to categorize the transaction
  3. Generates a MoneyWiz deep link for expense tracking
  4. Schedules a message to be sent to the user with transaction details

### 2. `monoHandler`
- **Purpose**: Processes Monobank transactions
- **Flow**:
  1. Receives transaction data from Monobank webhook
  2. Uses AI to categorize the transaction
  3. Generates a MoneyWiz deep link
  4. Schedules a message to be sent to the user with transaction details

### 3. `monoWebHook`
- **Purpose**: Receives webhook notifications from Monobank
- **Flow**:
  1. Validates incoming webhook requests
  2. Extracts transaction details
  3. Schedules the transaction for processing via Google Tasks

### 4. `sendMessage`
- **Purpose**: Sends messages to users via Telegram
- **Flow**:
  1. Receives message details via HTTP request
  2. Sends the message to the specified chat ID
  3. Supports message replies through `ReplyToMessageID`

### 5. `notificationHandler`
- **Purpose**: Processes bank push notifications forwarded by an iOS Shortcut (iOS 27+ can parse incoming push notifications and call a service)
- **Request body** (`application/json`):
  ```json
  { "app": "BBVA ES", "title": "Recibo cargado", "message": "Se ha cargado en tu cuenta *3297 un adeudo de ... de 79,81 EUR.", "date": "2026-06-26T13:13:00+03:00" }
  ```
  `date` is optional. It accepts an absolute instant (RFC3339/ISO 8601 with timezone, or a Unix epoch in seconds/milliseconds) or a naive datetime copied from the notification text (interpreted as Europe/Kyiv). When omitted or unrecognized, the current server time is used.
- **Flow**:
  1. Receives the source app name, notification title, message and optional date
  2. Uses AI to classify the notification into a category, subcategory, and amount
  3. Resolves the MoneyWiz account name from the source app and the masked account number in the message — BBVA maps to a single account, while PUMB (`Рахунок: *0451`) and Privat24 (`5*85`) resolve per card (see `resolveAccountName` in `internal/app/notifications.go`). Unrecognized apps/accounts fall back to the app name.
  4. Generates a MoneyWiz deep link (with the provided date) and shortens it
  5. Schedules a Telegram message with the categorized transaction and deep link

## Task Processing

The application uses Google Cloud Tasks for asynchronous processing:

- **Task Types**:
  - `ScheduledMessage`: Contains chat ID, message text, and optional reply message ID
  - `ScheduledTransaction`: Contains transaction details (MCC, category, description, amount)

- **Task Scheduling**:
  - Messages and transactions are scheduled for immediate processing
  - Uses Google Cloud Tasks API for reliable delivery
  - Tasks are processed in the order they are received

## Development

### Running Locally

For local development, you can run the main server:

```bash
go run cmd/main.go
```

The server will start on port 8080.

### Running Tests

```bash
go test ./...
```

### Building

```bash
go build ./...
```

### Updating Dependencies

```bash
./scripts/update_deps.sh
```

## CI/CD

The project includes GitHub Actions workflows for:

- **Build and Test**: Automatically builds and tests the project on pull requests
- **Deploy**: Deploys Cloud Functions to production (manual trigger)
- **Setup Scripts**: Workflows for setting up webhooks and bot configurations

## Scripts

- `deploy_functions.sh`: Deploys all Cloud Functions to Google Cloud
- `setup_gcloud_access.sh`: Sets up Google Cloud authentication
- `setup_mono_web_hook.sh`: Configures Monobank webhook
- `setup_telegram_bot.sh`: Configures Telegram bot webhook
- `cleanup_shortio_links.sh`: Cleans up expired Short.io links
- `update_deps.sh`: Updates Go dependencies
- `create_credentials_json.sh`: Creates credentials JSON for local development
