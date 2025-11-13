# SmartExpenseAI

SmartExpenseAI is an AI-powered Telegram bot that helps users track expenses using natural language processing. Users can send messages like "bought coffee 20k" and the bot will automatically extract the expense information and store it in a PostgreSQL database.

## Features
- Natural language expense tracking
- AI-powered text processing to extract expense details
- Telegram bot integration
- Expense categorization
- Expense summaries and recaps
- PostgreSQL database storage
- View last 10 expenses with IDs
- Monthly expense recap (last 30 days, sorted by month)
- Delete specific expenses by ID
- Update existing expenses
- Command-based interface (/lihat, /bulan, /hapus, /update, /bantuan)

## Architecture
- **Backend**: Go with Fiber framework
- **Database**: PostgreSQL with GORM
- **AI**: OpenRouter API for natural language processing
- **Telegram**: Telegram Bot API for messaging

## Project Structure
```
SmartExpenseAI/
├── .env
├── go.mod
├── cmd/
│   └── main.go
├── internal/
│   ├── routes/
│   │   └── telegram.go
│   ├── models/
│   │   └── expense.go
│   ├── database/
│   │   └── connection.go
│   └── services/
│       ├── ai.go
│       └── recap.go
```

## Environment Variables
- `TELEGRAM_BOT_TOKEN`: Telegram bot token
- `DATABASE_URL`: PostgreSQL database connection string
- `OPENROUTER_API_KEY`: API key for OpenRouter AI service
- `TELEGRAM_USER_ID`: Authorized user ID (optional, for single-user mode)

## Setup
1. Create a Telegram bot via BotFather and get the token
2. Set up a PostgreSQL database
3. Get an API key from OpenRouter
4. Fill in the .env file with required values
5. Run `go mod tidy` to install dependencies
6. Run `go run cmd/main.go` to start the application
7. Set up the webhook via `/setup-webhook` endpoint

## Usage
1. Send the `/start` or `/bantuan` command to see available commands
2. Send natural language expense messages (AI will extract expense details):
   - "makan nasi padang 25000"
   - "beli buku 50k"
   - "bensin 75.000"
3. Use command-based features:
   - `/lihat` - View last 10 expenses (with IDs for deletion)
   - `/bulan` - View monthly recap of last 30 days sorted by month
   - `/hapus ID` - Delete expense by ID (example: /hapus 5)
   - `/update ID description amount category` - Update expense (example: /update 5 beli buku 50000 Pendidikan)
   - `/recap` - View weekly expense summary