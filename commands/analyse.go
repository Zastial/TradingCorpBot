package commands

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"
	"tradingcorpbot/helpers"

	"github.com/bwmarrin/discordgo"
)

func (h *Handler) handleAnalyse(session *discordgo.Session, interaction *discordgo.InteractionCreate, data discordgo.ApplicationCommandInteractionData) {
	ticker := strings.ToUpper(strings.TrimSpace(helpers.OptionString(data.Options, "ticker")))
	if !helpers.IsValidTickerSymbol(ticker) || len(ticker) > 5 {
		_ = session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ Ticker invalide. Utilise uniquement des lettres (ex: AAPL, TSLA).",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	if h.analyzeFn == nil {
		message := "❌ Commande indisponible: analyse non configurée."
		_ = session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: message,
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	initialMessage := fmt.Sprintf("Analyse en cours pour **%s**... ⏳", ticker)
	_ = session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Content: initialMessage},
	})

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	user := interaction.User
	if user == nil && interaction.Member != nil {
		user = interaction.Member.User
	}

	statusMessage := fmt.Sprintf("✅ Analyse en cours pour **%s**. Merci de patienter.", ticker)
	if err := h.analyzeFn(ctx, ticker, user); err != nil {
		log.Printf("analyse webhook error: %v", err)
		statusMessage = "❌ Erreur lors de l'envoi vers n8n."
	}

	_, _ = session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
		Content: &statusMessage,
	})
}
