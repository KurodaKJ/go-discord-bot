package main

import (
	"context"
	"flag"
	"log"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

const systemInstruction = `YOUR_SYSTEM_INSTRUCTION_HERE`

var (
	client       *genai.Client
	model        *genai.GenerativeModel
	userSessions = make(map[string]*genai.ChatSession) // Chat history per user ID.
	mu           sync.Mutex                            // To safely handle concurrent access to userSessions.
)

func init() {
	initializeGenAIClient()
}

// Initializes the Generative AI client and model.
func initializeGenAIClient() {
	ctx := context.Background()
	var err error
	client, err = genai.NewClient(ctx, option.WithAPIKey("YOUR_KEY"))
	if err != nil {
		log.Fatalf("Failed to initialize AI client: %v", err)
	}

	model = client.GenerativeModel("gemini-1.5-flash")
	model.SystemInstruction = &genai.Content{
		Parts: []genai.Part{genai.Text(systemInstruction)},
	}

	// Configure model parameters.
	setModelParameters()

	log.Println("GenAI client and model initialized successfully.")
}

// Sets configuration parameters for the AI model.
func setModelParameters() {
	tempValue := float32(1)
	model.Temperature = &tempValue

	topPValue := float32(0.95)
	model.TopP = &topPValue

	topKValue := int32(40)
	model.TopK = &topKValue

	maxTokens := int32(8192)
	model.MaxOutputTokens = &maxTokens

	log.Println("Model parameters set: Temperature=1, TopP=0.95, TopK=40, MaxOutputTokens=8192")
}

// Retrieves or creates a chat session for a user.
func getUserChatSession(userID string) *genai.ChatSession {
	mu.Lock()
	defer mu.Unlock()

	if session, exists := userSessions[userID]; exists {
		log.Printf("Found existing session for user %s", userID)
		return session
	}

	// Create and store a new session for the user.
	newSession := model.StartChat()
	newSession.History = []*genai.Content{}
	userSessions[userID] = newSession

	log.Printf("Created new session for user %s", userID)
	return newSession
}

// Handles AI queries from users through Discord interactions.
func handleAskAI(s *discordgo.Session, i *discordgo.InteractionCreate, query string) {
	userID := i.Member.User.ID
	userName := i.Member.User.Username
	log.Printf("Received query from user %s (ID: %s): %s", userName, userID, query)

	// Get the user's chat session.
	session := getUserChatSession(userID)

	// Generate a response from the AI model.
	response, err := generateAIResponse(context.Background(), query)
	if err != nil {
		log.Printf("Error generating content for user %s (ID: %s): %v", userName, userID, err)
		return
	}

	// Update the chat session history with user query and model response.
	updateChatHistory(session, query, response)

	// Send the AI's response back to the Discord user.
	sendResponseToDiscord(s, i, response)
}

// Generates content based on the user's query using the AI model.
func generateAIResponse(ctx context.Context, query string) (string, error) {
	log.Printf("Generating response for query: %s", query)
	resp, err := model.GenerateContent(ctx, genai.Text(query))
	if err != nil {
		return "", err
	}

	// Extract response content.
	if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
		part := resp.Candidates[0].Content.Parts[0]
		if textPart, ok := part.(genai.Text); ok {
			responseText := string(textPart)
			log.Printf("Generated response: %s", responseText)
			return responseText, nil
		}
	}
	log.Println("No valid response generated.")
	return "No response available.", nil
}

// Updates the chat session history with the user query and AI response.
func updateChatHistory(session *genai.ChatSession, query, response string) {
	session.History = append(session.History, &genai.Content{
		Parts: []genai.Part{genai.Text(query)},
		Role:  "user",
	})
	session.History = append(session.History, &genai.Content{
		Parts: []genai.Part{genai.Text(response)},
		Role:  "model",
	})

	log.Println("Updated chat history with new query and response.")
}

// Sends the AI's response back to the Discord user.
func sendResponseToDiscord(s *discordgo.Session, i *discordgo.InteractionCreate, response string) {
	log.Printf("Sending response to user %s (ID: %s): %s", i.Member.User.Username, i.Member.User.ID, response)
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: response,
		},
	})
	if err != nil {
		log.Printf("Error responding to interaction: %v", err)
	}
}

// Main function to start the Discord bot.
func main() {
	token := flag.String("token", "", "Discord bot token")

	if *token == "" {
		log.Fatal("Please provide a Discord bot token using the -token flag.")
	}

	session, err := discordgo.New("Bot " + *token)
	if err != nil {
		log.Fatalf("Failed to create Discord session: %v", err)
	}

	session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Type == discordgo.InteractionApplicationCommand {
			handleAskAI(s, i, i.ApplicationCommandData().Options[0].StringValue())
		}
	})

	err = session.Open()
	if err != nil {
		log.Fatalf("Failed to open Discord session: %v", err)
	}

	log.Println("Bot is now running. Press CTRL+C to exit.")
	select {}
}
