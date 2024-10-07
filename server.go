package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal" // Import strings package
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

var client *genai.Client
var err error

// Chat history stored per user ID
var userSessions = make(map[string]*genai.ChatSession)
var mu sync.Mutex // to safely handle concurrent access to userSessions

func init() {
	// Initialize the genai client with the API key
	ctx := context.Background()
	client, err = genai.NewClient(ctx, option.WithAPIKey("YOUR_KEY"))
	if err != nil {
		log.Fatal(err)
	}
}

// Get or create a chat session for a user
func getUserChatSession(userID string) *genai.ChatSession {
	mu.Lock()
	defer mu.Unlock()

	// Check if the user already has a session
	if session, exists := userSessions[userID]; exists {
		return session
	}

	// Create a new session for the user if one doesn't exist
	model := client.GenerativeModel("gemini-1.5-flash")
	model.GenerationConfig.Temperature = 0.7
	model.TopP = 0.95
	model.TopK = 40
	model.GenerationConfig.MaxOutputTokens = 8192

	newSession := model.StartChat()

	system_instruction := "Doesn't work yet"

	// Optionally initialize the session with some predefined history
	newSession.History = []*genai.Content{
		{
			Parts: []genai.Part{
				genai.Text(system_instruction),
			},
			Role: "user",
		},
	}

	// Store the new session
	userSessions[userID] = newSession
	return newSession
}

func handleAskAI(s *discordgo.Session, i *discordgo.InteractionCreate, query string) {
	ctx := context.Background()

	// Get the user's chat session
	userID := i.Member.User.ID
	session := getUserChatSession(userID)

	// Using the existing genai model
	model := client.GenerativeModel("gemini-1.5-flash")

	// Generate content based on the query using the user's session
	resp, err := model.GenerateContent(ctx, genai.Text(query))
	if err != nil {
		log.Printf("Error generating content: %s", err)
		return
	}

	// Log the response for debugging
	log.Printf("Response: %+v\n", resp)

	// Variable to hold the response text
	contentText := ""

	// Check if there are candidates and parts
	if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
		part := resp.Candidates[0].Content.Parts[0]

		// Type assertion to extract the string content
		if textPart, ok := part.(genai.Text); ok {
			contentText = string(textPart) // Convert the Text type back to string
		} else {
			log.Println("The part is not of type Text.")
		}
	} else {
		contentText = "No response available."
	}

	// Update session history
	session.History = append(session.History, &genai.Content{
		Parts: []genai.Part{
			genai.Text(query),
		},
		Role: "user",
	})
	session.History = append(session.History, &genai.Content{
		Parts: []genai.Part{
			genai.Text(contentText),
		},
		Role: "model",
	})

	// Respond to the Discord interaction
	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: contentText,
		},
	})
	if err != nil {
		log.Printf("Error responding to interaction: %s", err)
	}
}

func main() {
	Token := flag.String("token", "", "Bot authentication token")
	App := flag.String("app", "", "Application ID")
	Guild := flag.String("guild", "", "Guild ID")
	flag.Parse()

	if *App == "" {
		log.Fatal("application id is not set")
	}

	session, err := discordgo.New("Bot " + *Token)
	if err != nil {
		log.Fatalf("error creating Discord session: %s", err)
	}

	// Add command handler
	session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Type != discordgo.InteractionApplicationCommand {
			return
		}

		data := i.ApplicationCommandData()
		if data.Name == "askkaty" { // Change here to 'askkaty'
			query := data.Options[0].StringValue() // assuming your command has one string option
			handleAskAI(s, i, query)
		}
	})

	// Log in notification
	session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as %s", r.User.String())
	})

	// Register commands with Discord
	_, err = session.ApplicationCommandBulkOverwrite(*App, *Guild, []*discordgo.ApplicationCommand{
		{
			Name:        "askkaty", // Change here to 'askkaty'
			Description: "Ask Katy a question",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "question",
					Description: "Your question to the AI",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    true,
				},
			},
		},
	})
	if err != nil {
		log.Fatalf("could not register commands: %s", err)
	}

	// Open Discord session
	err = session.Open()
	if err != nil {
		log.Fatalf("could not open session: %s", err)
	}

	// Wait for a termination signal
	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, os.Interrupt)
	<-sigch

	err = session.Close()
	if err != nil {
		log.Printf("could not close session gracefully: %s", err)
	}
}
