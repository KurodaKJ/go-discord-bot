<h1 align="center">Gemini Discord Bot</h1>

<p align="center">
  Gemini Discord Bot is a powerful Discord bot that utilizes generative AI to provide interactive and engaging chat experiences. Built with Go and integrated with the Gemini AI API, this bot can handle various AI queries and respond intelligently to users on Discord.
</p>

## Getting Started

These instructions will get you a copy of the project up and running on your local machine for development and testing purposes.

### Prerequisites

- Golang
- A Discord bot token
- A Gemini API key

1. **Clone the repository**:

      ```sh
      git clone https://github.com/KurodaKJ/go-discord-bot.git
      cd go-discord-bot
      ```

2. **Install required dependencies**:

      ```sh
      go get ./...
      ```

3. **Run the main server**:

      ```sh
      go run server.go -token "YOUR_DISCORD_BOT_TOKEN" -apikey "YOUR_GEMINI_API_KEY" -system-instruction "YOUR_CUSTOM_INSTRUCTION"
      ```

## Docker Deployment

If your goal is just to deploy your Discord bot, then it's highly recommended to deploy with Docker.

### Steps

1. **Build the Docker image**:

      ```sh
      docker build -t gemini-discord-bot .
      ```

2. **Run the container with the following command**:

      ```sh
      docker run -d --name gemini-discord-bot -e TOKEN_VARIABLE="YOUR_DISCORD_TOKEN" -e API_KEY_VARIABLE="YOUR_GEMINI_API_KEY" -e SYSTEM_INSTRUCTION_VARIABLE="YOUR_CUSTOM_INSTRUCTION" gemini-discord-bot
      ```

## Contributing

Contributions are welcome! Please open an issue or submit a pull request with your changes.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [discordgo](https://github.com/bwmarrin/discordgo) - Discord API library for Go
- [generative-ai-go](https://github.com/google/generative-ai-go) - Gemini AI client for Go

## Contact

For questions or feedback, feel free to open an issue or reach out to the repository maintainer.
