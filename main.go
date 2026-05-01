package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"

	"tradingcorpbot/cache"
	botcommands "tradingcorpbot/commands"
	n8n "tradingcorpbot/n8n"
	"tradingcorpbot/nasdaq"
	"tradingcorpbot/types"
)

func main() {
	cfg, err := LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	cache.Init()

	commandHandler := botcommands.NewHandler(
		func(ctx context.Context, ticker string, user *discordgo.User) error {
			return n8n.SendRequest(ctx, cfg, ticker, user)
		},
		nasdaq.FetchAllStocks,
	)

	session, err := discordgo.New("Bot " + cfg.DiscordToken)
	if err != nil {
		log.Fatal(err)
	}

	session.AddHandler(func(s *discordgo.Session, interaction *discordgo.InteractionCreate) {
		commandHandler.HandleInteraction(s, interaction)
	})
	session.AddHandlerOnce(func(s *discordgo.Session, ready *discordgo.Ready) {
		log.Printf("Connecté en tant que %s", ready.User.Username)
		log.Printf("N8N mode prod: %t", cfg.Prod)
		log.Printf("N8N webhook URL: %s", cfg.N8NWebhookURL())
	})

	if err := registerCommands(session, cfg, commandHandler); err != nil {
		log.Fatal(err)
	}

	if err := session.Open(); err != nil {
		log.Fatal(err)
	}
	defer session.Close()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
}

func registerCommands(session *discordgo.Session, cfg types.Config, handler *botcommands.Handler) error {
	commands := handler.Definitions()

	_, err := session.ApplicationCommandBulkOverwrite(cfg.DiscordAppID, "", commands)
	return err
}
