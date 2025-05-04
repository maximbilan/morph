# Morph - Financial Transaction Management System

Morph is a serverless application built on Google Cloud Functions that helps manage and categorize financial transactions. The system integrates with Monobank for transaction monitoring and uses AI to automatically categorize expenses. The primary output of the system is deep links for the MoneyWiz app, allowing users to quickly add categorized expenses to their financial tracking.

## Architecture Overview

The application consists of four main Google Cloud Functions that work together to process and manage financial transactions:

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

## Task Processing

The application uses Google Cloud Tasks for asynchronous processing:

- **Task Types**:
  - `ScheduledMessage`: Contains chat ID, message text, and optional reply message ID
  - `ScheduledTransaction`: Contains transaction details (MCC, category, description, amount)

- **Task Scheduling**:
  - Messages and transactions are scheduled for immediate processing
  - Uses Google Cloud Tasks API for reliable delivery
  - Tasks are processed in the order they are received
