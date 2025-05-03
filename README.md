# Morph

A Go-based service that transforms financial transactions into MoneyWiz deep links using AI analysis.

## Features

### Telegram Bot Integration
- Monitors a Telegram chat for expense messages
- Parses free-format expense descriptions
- Uses AI to categorize expenses
- Generates MoneyWiz deep links for easy expense tracking
- Shortens URLs for better sharing

### Monobank Integration
- Provides webhook endpoint for Monobank transactions
- Automatically processes incoming bank transactions
- Analyzes transaction data using AI categorization
- Generates MoneyWiz deep links for automatic expense tracking
- Sends results to configured Telegram chat