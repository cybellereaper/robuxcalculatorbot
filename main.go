package main

import (
	"fmt"
	"log"
	"math"
	"os"

	"github.com/bwmarrin/discordgo"
)

// Constants
const (
	ExchangeRate = 1.22 // GBP to USD exchange rate
	MarkupRate   = 0.30 // 30% markup for 'a/t' type
)

// PriceType represents the type of Robux price calculation
type PriceType string

const (
	BT PriceType = "b/t"
	AT PriceType = "a/t"
)

// PricePerRobux maps PriceType to GBP per Robux
var PricePerRobux = map[PriceType]float64{
	BT: 0.0045,  // GBP per Robux for 'b/t'
	AT: 0.00675, // GBP per Robux for 'a/t'
}

// ConvertGBPToUSD converts GBP to USD using the exchange rate
func ConvertGBPToUSD(gbp float64) float64 {
	return gbp * ExchangeRate
}

// ConvertUSDToGBP converts USD to GBP using the exchange rate
func ConvertUSDToGBP(usd float64) float64 {
	return usd / ExchangeRate
}

// HandleInteraction processes Discord slash commands concurrently
func HandleInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.ApplicationCommandData().Name {
	case "price":
		go handlePriceCommand(s, i)
	case "convert":
		go handleConvertCommand(s, i)
	case "robux":
		go handleRobuxCommand(s, i)
	case "help":
		go handleHelpCommand(s, i)
	}
}

func handlePriceCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
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

	gbpAmount := float64(amount) * rate
	gamepassPrice := calculateGamepassPrice(priceType, amount)
	botUser := s.State.User

	embed := createEmbed("Price Calculation", fmt.Sprintf("**Conversion Type:** %s\n**Amount of Robux:** %d", priceType, amount), botUser)
	embed.Fields = []*discordgo.MessageEmbedField{
		{Name: "Gamepass Price", Value: fmt.Sprintf("%d R$", gamepassPrice), Inline: true},
		{Name: "Amount in GBP", Value: fmt.Sprintf("£%.2f", gbpAmount), Inline: true},
		{Name: "Amount in USD", Value: fmt.Sprintf("$%.2f", ConvertGBPToUSD(gbpAmount)), Inline: true},
	}

	sendEmbedResponse(s, i.Interaction, embed)
}

func handleHelpCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	embed := createEmbed("Available Commands", "Here are the available commands and their usage:\n"+
		"`/price`: Calculate the price in GBP and USD for a given amount of Robux\n"+
		"`/convert`: Convert between GBP and USD\n"+
		"`/robux`: Convert GBP or USD to the amount of Robux", s.State.User)

	sendEmbedResponse(s, i.Interaction, embed)
}

func handleConvertCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options
	if len(options) < 2 {
		RespondWithError(s, i.Interaction, "Insufficient command options")
		return
	}

	currency := options[0].StringValue()
	amount := options[1].FloatValue()
	botUser := s.State.User

	embed := createEmbed("Currency Conversion", "", botUser)
	switch currency {
	case "GBP":
		convertedAmount := ConvertGBPToUSD(amount)
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "Amount in GBP",
			Inline: true,
			Value:  fmt.Sprintf("£%.2f", amount),
		}, &discordgo.MessageEmbedField{
			Name:   "Amount in USD",
			Inline: true,
			Value:  fmt.Sprintf("$%.2f", convertedAmount),
		})
	case "USD":
		convertedAmount := ConvertUSDToGBP(amount)
		embed.Fields = append(embed.Fields,
			&discordgo.MessageEmbedField{
				Name:   "Amount in USD",
				Inline: true,
				Value:  fmt.Sprintf("$%.2f", amount),
			},
			&discordgo.MessageEmbedField{
				Name:   "Amount in GBP",
				Inline: true,
				Value:  fmt.Sprintf("£%.2f", convertedAmount),
			})
	default:
		RespondWithError(s, i.Interaction, "Invalid currency. Use 'GBP' or 'USD'.")
		return
	}

	sendEmbedResponse(s, i.Interaction, embed)
}

func handleRobuxCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options
	if len(options) < 2 {
		RespondWithError(s, i.Interaction, "Insufficient command options")
		return
	}

	currency := options[0].StringValue()
	amount := options[1].FloatValue()
	botUser := s.State.User

	embed := createEmbed("Robux Calculation", "", botUser)
	switch currency {
	case "GBP":
		robuxAmount := int64(amount / PricePerRobux[BT])
		embed.Description = fmt.Sprintf("£%.2f affords %d R$ ($%.2f)", amount, robuxAmount, ConvertGBPToUSD(amount))

	case "USD":
		gbpAmount := ConvertUSDToGBP(amount)
		robuxAmount := int64(gbpAmount / PricePerRobux[BT])
		embed.Description = fmt.Sprintf("$%.2f affords %d R$ (£%.2f)", amount, robuxAmount, gbpAmount)

	default:
		RespondWithError(s, i.Interaction, "Invalid currency. Use 'GBP' or 'USD'.")
		return
	}

	sendEmbedResponse(s, i.Interaction, embed)
}

// ParseCommandOptions parses the command options and returns the price type and amount
func ParseCommandOptions(options []*discordgo.ApplicationCommandInteractionDataOption) (PriceType, int64, error) {
	if len(options) < 2 {
		return "", 0, fmt.Errorf("insufficient command options")
	}
	priceType := PriceType(options[0].StringValue())
	amount, err := parseAmount(options[1].Value)
	return priceType, amount, err
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

// calculateGamepassPrice calculates the gamepass price based on the price type and amount
func calculateGamepassPrice(priceType PriceType, amount int64) int64 {
	if priceType == AT {
		return int64(math.Round(float64(amount)/(1-MarkupRate))) + 1
	}
	return amount
}

// createEmbed creates a Discord embed message
func createEmbed(title, description string, botUser *discordgo.User) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title:       title,
		Description: description,
		Color:       0x0096FF,
		Footer: &discordgo.MessageEmbedFooter{
			Text:    fmt.Sprint("Powered by ", botUser.Username),
			IconURL: botUser.AvatarURL("2048"),
		},
	}
}

// sendEmbedResponse sends an embed response to the interaction
func sendEmbedResponse(s *discordgo.Session, interaction *discordgo.Interaction, embed *discordgo.MessageEmbed) {
	if err := s.InteractionRespond(interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Embeds: []*discordgo.MessageEmbed{embed}, Flags: 64},
	}); err != nil {
		log.Printf("Failed to send response: %v", err)
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
	commands := []*discordgo.ApplicationCommand{
		{
			Name:        "help",
			Description: "Display the available commands and their usage",
		},
		{
			Name:        "price",
			Description: "Calculate the price in GBP and USD for a given amount of Robux",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "type",
					Description: "Conversion type (b/t or a/t)",
					Required:    true,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{Name: "b/t", Value: "b/t"},
						{Name: "a/t", Value: "a/t"},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "amount",
					Description: "Amount of Robux",
					Required:    true,
				},
			},
		},
		{
			Name:        "convert",
			Description: "Convert between GBP and USD",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "currency",
					Description: "Currency to convert from (GBP or USD)",
					Required:    true,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{Name: "GBP", Value: "GBP"},
						{Name: "USD", Value: "USD"},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionNumber,
					Name:        "amount",
					Description: "Amount to convert",
					Required:    true,
				},
			},
		},
		{
			Name:        "robux",
			Description: "Convert GBP or USD to the amount of Robux",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "currency",
					Description: "Currency to convert from (GBP or USD)",
					Required:    true,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{Name: "GBP", Value: "GBP"},
						{Name: "USD", Value: "USD"},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionNumber,
					Name:        "amount",
					Description: "Amount to convert",
					Required:    true,
				},
			},
		},
	}

	for _, cmd := range commands {
		_, err := dg.ApplicationCommandCreate(dg.State.User.ID, "", cmd)
		if err != nil {
			return err
		}
	}
	return nil
}
