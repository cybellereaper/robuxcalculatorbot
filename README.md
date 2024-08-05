Robux Price Calculator Bot

This is a Discord bot designed to calculate the price of Robux in GBP and USD.

Features:

- Calculate Robux prices in GBP and USD with the /price command.
- Supports two pricing types: b/t and a/t.

Prerequisites:

- Go (installation guide: https://golang.org/doc/install)
- Git (installation guide: https://git-scm.com/book/en/v2/Getting-Started-Installing-Git)
- Discord account (for obtaining a bot token)

Setup Instructions:

1. Clone the repository:
   git clone https://github.com/your-username/your-repository.git
   cd your-repository

2. Install dependencies:
   go mod tidy

3. Create a .env file with the following content:
   DISCORD_TOKEN=your-discord-bot-token

4. Run the bot:
   go run main.go

Commands:

- /price

Options:

- type (required): Select either b/t or a/t.
- amount (required): Specify the amount of Robux.

Example Usage:

To calculate the price of 100 Robux with the b/t pricing type, use the command:
   /price type:b/t amount:100
