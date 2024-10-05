package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"

	"github.com/bwmarrin/discordgo"
)

var (
	Token = "YOUR_TOKEN" // Replace with your actual bot token
)

// Command definition
var commands = []*discordgo.ApplicationCommand{
	{
		Name:        "echo",
		Description: "Say something through the bot",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "message",
				Description: "Contents of the message",
				Type:        discordgo.ApplicationCommandOptionString,
				Required:    true,
			},
			{
				Name:        "author",
				Description: "Whether to prepend the author's name",
				Type:        discordgo.ApplicationCommandOptionBoolean,
			},
		},
	},
}

func main() {
	// Create a new Discord session
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		log.Fatalf("error creating Discord session: %v", err)
		return
	}

	// Register the command handler
	dg.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		// Handle slash commands
		if i.Type == discordgo.InteractionApplicationCommand {
			data := i.ApplicationCommandData()
			if data.Name == "echo" {
				handleEcho(s, i, parseOptions(data.Options))
			}
		}
	})

	// Log bot ready event
	dg.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as %s", r.User.String())

		// Register commands once the bot is ready
		_, err := s.ApplicationCommandBulkOverwrite(s.State.User.ID, "", commands)
		if err != nil {
			log.Fatalf("could not register commands: %v", err)
		}
	})

	// Open a connection to Discord
	err = dg.Open()
	if err != nil {
		log.Fatalf("error opening connection: %v", err)
		return
	}
	fmt.Println("Bot is now running. Press CTRL-C to exit.")

	// Wait for an interrupt signal to gracefully shutdown the bot
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, os.Interrupt, os.Kill)
	<-sc

	// Clean up and close the session
	err = dg.Close()
	if err != nil {
		log.Printf("error closing session: %v", err)
	}
}

// Helper function to parse options from the command
func parseOptions(options []*discordgo.ApplicationCommandInteractionDataOption) map[string]*discordgo.ApplicationCommandInteractionDataOption {
	om := make(map[string]*discordgo.ApplicationCommandInteractionDataOption)
	for _, opt := range options {
		om[opt.Name] = opt
	}
	return om
}

// Handle the echo command
func handleEcho(s *discordgo.Session, i *discordgo.InteractionCreate, opts map[string]*discordgo.ApplicationCommandInteractionDataOption) {
	builder := new(strings.Builder)
	if v, ok := opts["author"]; ok && v.BoolValue() {
		author := i.Member.User
		builder.WriteString("**" + author.String() + "** says: ")
	}
	builder.WriteString(opts["message"].StringValue())

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: builder.String(),
		},
	})
	if err != nil {
		log.Printf("could not respond to interaction: %v", err)
	}
}
