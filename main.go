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
	ExchangeRate = 1.22 // GBP to USD exchange rate (1 GBP = 1.22 USD)
	MarkupRate   = 0.30 // 30% markup rate for 'a/t' type
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

// HandleInteraction processes Discord slash commands
func HandleInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand || i.ApplicationCommandData().Name != "price" {
		return
	}

	priceType, amount, err := ParseCommandOptions(i.ApplicationCommandData().Options)
	if err != nil {
		RespondWithError(s, i.Interaction, fmt.Sprintf("Error: %v", err))
		return
	}

	rate, exists := PricePerRobux[priceType]
	if !exists {
		RespondWithError(s, i.Interaction, "Invalid type. Use 'b/t' or 'a/t'.")
		return
	}

	var gbpAmount float64
	var gamepassPrice int64

	if priceType == BT {
		gbpAmount = float64(amount) * rate
		gamepassPrice = amount
	} else {
		// For 'a/t', we include the 30% markup
		gbpAmount = float64(amount) * rate
		gamepassPrice = int64(math.Round(float64(amount) / (1 - MarkupRate)))
	}

	usdAmount := ConvertGBPToUSD(gbpAmount)

	embed := &discordgo.MessageEmbed{
		Title:       "Price Calculation",
		Description: fmt.Sprintf("**Conversion Type:** %s\n**Amount of Robux:** %d", priceType, amount),
		Color:       0x00FF00, // Green color
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Gamepass Price", Value: fmt.Sprintf("%d R$", gamepassPrice+1), Inline: true},
			{Name: "Amount in GBP", Value: fmt.Sprintf("£%.2f", gbpAmount), Inline: true},
			{Name: "Amount in USD", Value: fmt.Sprintf("$%.2f", usdAmount), Inline: true},
		},
		Footer: &discordgo.MessageEmbedFooter{Text: "Powered by your friendly Discord bot"},
	}

	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Embeds: []*discordgo.MessageEmbed{embed}},
	}); err != nil {
		log.Printf("Failed to send response: %v", err)
	}
}

// ParseCommandOptions extracts PriceType and amount from interaction options
func ParseCommandOptions(options []*discordgo.ApplicationCommandInteractionDataOption) (PriceType, int64, error) {
	if len(options) < 2 {
		return "", 0, fmt.Errorf("insufficient command options")
	}

	priceType := PriceType(options[0].StringValue())
	amount, err := parseAmount(options[1].Value)
	if err != nil {
		return "", 0, err
	}

	return priceType, amount, nil
}

// parseAmount converts value to int64
func parseAmount(value interface{}) (int64, error) {
	switch v := value.(type) {
	case int64:
		return v, nil
	case float64:
		return int64(math.Round(v)), nil
	default:
		return 0, fmt.Errorf("unexpected type %T for amount", value)
	}
}

// RespondWithError sends an error message as a response to the interaction
func RespondWithError(s *discordgo.Session, interaction *discordgo.Interaction, message string) {
	if err := s.InteractionRespond(interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Content: message},
	}); err != nil {
		log.Printf("Failed to send error response: %v", err)
	}
}

// main initializes the bot, registers commands, and starts listening
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
