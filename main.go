package main

import (
	"fmt"
	"log"
	"math"
	"os"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

// Constants
const (
	ExchangeRate = 1.2195 // GBP to USD exchange rate
	MarkupRate   = 0.30   // 30% markup rate for 'a/t' type
)

// PriceType represents the type of Robux price calculation
type PriceType string

const (
	BT PriceType = "b/t"
	AT PriceType = "a/t"
)

// PricePerRobux stores the price per Robux for different types
var PricePerRobux = map[PriceType]float64{
	BT: 0.0045,  // GBP per Robux for b/t
	AT: 0.00675, // GBP per Robux for a/t
}

// ConvertGBPToUSD converts GBP to USD
func ConvertGBPToUSD(gbpAmount float64) float64 {
	return gbpAmount * ExchangeRate
}

// HandleInteraction handles Discord slash commands
func HandleInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}

	data := i.ApplicationCommandData()
	if data.Name != "price" {
		return
	}

	priceType, amount, err := ParseCommandOptions(data.Options)
	if err != nil {
		RespondWithError(s, i.Interaction, fmt.Sprintf("Error: %v", err))
		return
	}

	rate, ok := PricePerRobux[priceType]
	if !ok {
		RespondWithError(s, i.Interaction, "Invalid type. Use 'b/t' or 'a/t'.")
		return
	}

	gbpAmount := float64(amount) * rate
	if priceType == AT {
		gbpAmount += gbpAmount * MarkupRate
	}
	usdAmount := ConvertGBPToUSD(gbpAmount)

	// Create an embed for the response
	embed := &discordgo.MessageEmbed{
		Title:       "Price Calculation",
		Description: fmt.Sprintf("**Conversion Type:** %s\n**Amount of Robux:** %d", priceType, amount),
		Color:       0x00FF00, // Green color
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Amount in GBP",
				Value:  fmt.Sprintf("£%.2f", gbpAmount),
				Inline: true,
			},
			{
				Name:   "Amount in USD",
				Value:  fmt.Sprintf("$%.2f", usdAmount),
				Inline: true,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Powered by your friendly Discord bot",
		},
	}

	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	}); err != nil {
		log.Printf("Failed to send response: %v", err)
	}
}

// ParseCommandOptions parses the command options from the interaction data
func ParseCommandOptions(options []*discordgo.ApplicationCommandInteractionDataOption) (PriceType, int64, error) {
	if len(options) < 2 {
		return "", 0, fmt.Errorf("insufficient command options")
	}

	priceType := PriceType(options[0].StringValue())

	var amount int64
	switch v := options[1].Value.(type) {
	case int64:
		amount = v
	case float64:
		amount = int64(math.Round(v))
	default:
		return "", 0, fmt.Errorf("failed to parse amount: unexpected type %T", options[1].Value)
	}

	return priceType, amount, nil
}

// RespondWithError sends an error message as a response to the interaction
func RespondWithError(s *discordgo.Session, interaction *discordgo.Interaction, message string) {
	if err := s.InteractionRespond(interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: message,
		},
	}); err != nil {
		log.Printf("Failed to send error response: %v", err)
	}
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		log.Fatal("DISCORD_TOKEN environment variable is required")
	}

	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatalf("Error creating Discord session: %v", err)
	}

	if err := dg.Open(); err != nil {
		log.Fatalf("Error opening connection: %v", err)
	}
	defer dg.Close()

	if err := RegisterSlashCommands(dg); err != nil {
		log.Fatalf("Error registering slash commands: %v", err)
	}

	dg.AddHandler(HandleInteraction)

	fmt.Println("Bot is now running. Press CTRL+C to exit.")
	select {}
}

// RegisterSlashCommands registers the slash commands for the bot
func RegisterSlashCommands(dg *discordgo.Session) error {
	_, err := dg.ApplicationCommandCreate(dg.State.User.ID, "", &discordgo.ApplicationCommand{
		Name:        "price",
		Description: "Calculate the price in GBP and USD for a given amount of Robux",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "type",
				Description: "The type of conversion (b/t or a/t)",
				Required:    true,
				Choices: []*discordgo.ApplicationCommandOptionChoice{
					{Name: "b/t", Value: "b/t"},
					{Name: "a/t", Value: "a/t"},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "amount",
				Description: "The amount of Robux",
				Required:    true,
			},
		},
	})
	return err
}
