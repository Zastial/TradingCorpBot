package commands

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"
	"tradingcorpbot/cache"
	"tradingcorpbot/helpers"
	"tradingcorpbot/nasdaq"
	"tradingcorpbot/types"

	"github.com/bwmarrin/discordgo"
)

const messageLimit = 1800

func (h *Handler) handleTickers(session *discordgo.Session, interaction *discordgo.InteractionCreate, data discordgo.ApplicationCommandInteractionData) {
	prefix, company := parseOptions(data)

	if !validateOptions(session, interaction, prefix, company) {
		return
	}

	if h.fetchStocksFn == nil {
		respondEphemeral(session, interaction, "❌ Commande indisponible: récupération des tickers non configurée.")
		return
	}

	_ = session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})

	// Récupération + filtrage
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()

	filtered, err := h.fetchAndFilter(ctx, prefix, company)
	if err != nil {
		log.Printf("tickers error: %v", err)
		message := "❌ Impossible de récupérer les tickers."
		_, _ = session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{Content: &message})
		return
	}

	if len(filtered) == 0 {
		searchLabel := fmt.Sprintf("le nom « %s »", company)
		if prefix != "" {
			searchLabel = fmt.Sprintf("le préfixe « %s »", prefix)
		}
		message := fmt.Sprintf("Aucun ticker trouvé pour %s.", searchLabel)
		_, _ = session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{Content: &message})
		return
	}

	// Limitation pour recherche par nom
	if company != "" && len(filtered) > 10 {
		filtered = filtered[:10]
	}

	sort.SliceStable(filtered, func(i, j int) bool { return filtered[i].Symbol < filtered[j].Symbol })

	title := buildTitle(prefix, company, len(filtered))

	chunks := helpers.BuildFriendlyChunks(title, filtered)
	if len(chunks) == 0 {
		respondEdit(session, interaction, "❌ Aucun contenu à afficher.")
		return
	}

	sendChunks(session, interaction, chunks)
}

func parseOptions(data discordgo.ApplicationCommandInteractionData) (string, string) {
	return strings.TrimSpace(helpers.OptionString(data.Options, "prefix")), strings.TrimSpace(helpers.OptionString(data.Options, "company"))
}

func validateOptions(session *discordgo.Session, interaction *discordgo.InteractionCreate, prefix, company string) bool {
	if prefix != "" && company != "" {
		respondEphemeral(session, interaction, "❌ Utilise soit `prefix`, soit `company`, mais pas les deux en même temps.")
		return false
	}
	if prefix == "" && company == "" {
		respondEphemeral(session, interaction, "❌ Tu dois fournir au moins un argument: `prefix` ou `company`.")
		return false
	}
	return true
}

func respondEphemeral(session *discordgo.Session, interaction *discordgo.InteractionCreate, msg string) {
	_ = session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Content: msg, Flags: discordgo.MessageFlagsEphemeral},
	})
}

func respondEdit(session *discordgo.Session, interaction *discordgo.InteractionCreate, msg string) {
	_, _ = session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{Content: &msg})
}

func sendChunks(session *discordgo.Session, interaction *discordgo.InteractionCreate, chunks []string) {
	first := chunks[0]
	_, _ = session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{Content: &first})

	for _, chunk := range chunks[1:] {
		_, _ = session.ChannelMessageSendComplex(interaction.ChannelID, &discordgo.MessageSend{
			Content:         chunk,
			AllowedMentions: &discordgo.MessageAllowedMentions{Parse: []discordgo.AllowedMentionType{}},
		})
	}
}

func buildTitle(prefix, company string, count int) string {
	title := fmt.Sprintf("✨ Tickers disponibles (%d)", count)
	if prefix != "" {
		title = fmt.Sprintf("✨ Tickers commençant par « %s » (%d)", prefix, count)
	}
	if company != "" {
		title = fmt.Sprintf("✨ Tickers correspondant à « %s » - Top 10 (%d)", company, count)
	}
	return title
}

func (h *Handler) fetchAndFilter(ctx context.Context, prefix, company string) ([]types.Stock, error) {
	// Try reading from cache first
	stocks, err := cache.GetStocksFromCache(ctx)
	if err != nil {
		log.Printf("Redis unavailable, falling back to Nasdaq: %v", err)
	}

	// If cache miss or error -> fetch from Nasdaq
	if stocks == nil {
		stocks, err = nasdaq.FetchAllStocks(ctx)
		if err != nil {
			return nil, fmt.Errorf("nasdaq fetch failed: %w", err)
		}

		// Populate cache asynchronously
		go func(s []types.Stock) {
			if err := cache.SetStocksInCache(context.Background(), s); err != nil {
				log.Printf("Failed to write to Redis cache: %v", err)
			}
		}(stocks)
	}

	return filterStocks(stocks, prefix, company), nil
}

// filterStocks filtre la liste par préfixe de ticker et/ou nom d'entreprise.
func filterStocks(stocks []types.Stock, prefix, company string) []types.Stock {
	var result []types.Stock
	for _, s := range stocks {
		if prefix != "" && !strings.HasPrefix(s.Symbol, strings.ToUpper(prefix)) {
			continue
		}
		if company != "" && !strings.Contains(strings.ToLower(s.Name), strings.ToLower(company)) {
			continue
		}
		result = append(result, s)
	}
	return result
}
