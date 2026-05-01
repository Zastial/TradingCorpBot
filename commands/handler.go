package commands

import (
	"context"
	"log"
	"time"
	"tradingcorpbot/types"

	"github.com/bwmarrin/discordgo"

	"tradingcorpbot/cache"
)

type AnalyzeFunc func(ctx context.Context, ticker string, user *discordgo.User) error
type FetchStocksFunc func(ctx context.Context) ([]types.Stock, error)

type Handler struct {
	analyzeFn     AnalyzeFunc
	fetchStocksFn FetchStocksFunc
}

func NewHandler(analyzeFn AnalyzeFunc, fetchStocksFn FetchStocksFunc) *Handler {
	return &Handler{
		analyzeFn:     analyzeFn,
		fetchStocksFn: fetchStocksFn,
	}
}

func (h *Handler) Definitions() []*discordgo.ApplicationCommand {
	return []*discordgo.ApplicationCommand{
		{
			Name:        "analyse",
			Description: "Analyse une action",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "ticker",
					Description: "Nom de l'action",
					Required:    true,
				},
			},
		},
		{
			Name:        "tickers",
			Description: "Affiche les tickers disponibles",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "prefix",
					Description: "Filtre les tickers qui commencent par ce texte",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "company",
					Description: "Recherche par nom d'entreprise",
					Required:    false,
				},
			},
		},
	}
}

func (h *Handler) HandleInteraction(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	if interaction.Type != discordgo.InteractionApplicationCommand {
		return
	}

	locked, err := cache.AcquireInteractionLock(context.Background(), interaction.ID, 2*time.Minute)
	if err != nil {
		log.Printf("interaction lock error: %v", err)
	} else if !locked {
		return
	}

	data := interaction.ApplicationCommandData()

	switch data.Name {
	case "analyse":
		h.handleAnalyse(session, interaction, data)
	case "tickers":
		h.handleTickers(session, interaction, data)
	default:
		log.Printf("unknown command: %s", data.Name)
	}
}
