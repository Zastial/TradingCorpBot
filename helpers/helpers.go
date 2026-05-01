package helpers

import (
	"fmt"
	"strings"
	"tradingcorpbot/types"

	"github.com/bwmarrin/discordgo"
)

func IsValidTickerSymbol(symbol string) bool {
	if len(symbol) == 0 || len(symbol) > 15 {
		return false
	}

	for _, r := range symbol {
		if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '.' || r == '-' {
			continue
		}

		return false
	}

	return true
}

func normalizeText(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	replacements := []string{
		"à", "a",
		"â", "a",
		"ä", "a",
		"á", "a",
		"ã", "a",
		"å", "a",
		"ç", "c",
		"é", "e",
		"è", "e",
		"ê", "e",
		"ë", "e",
		"î", "i",
		"ï", "i",
		"í", "i",
		"ô", "o",
		"ö", "o",
		"ó", "o",
		"õ", "o",
		"ù", "u",
		"û", "u",
		"ü", "u",
		"ú", "u",
		"ñ", "n",
	}

	replacer := strings.NewReplacer(replacements...)
	return replacer.Replace(value)
}

// func filterStocks(stocks []types.Stock, prefix, company string) []types.Stock {
// 	filtered := make([]types.Stock, 0, len(stocks))
// 	normalizedPrefix := normalizeText(prefix)
// 	normalizedCompany := normalizeText(company)

// 	for _, stock := range stocks {
// 		if prefix != "" && !strings.HasPrefix(normalizeText(stock.Symbol), normalizedPrefix) {
// 			continue
// 		}

// 		if company != "" && !strings.Contains(normalizeText(stock.Name), normalizedCompany) {
// 			continue
// 		}

// 		filtered = append(filtered, stock)
// 	}

// 	return filtered
// }

func BuildFriendlyChunks(title string, stocks []types.Stock) []string {
	chunks := make([]string, 0, 2)
	var current strings.Builder
	current.WriteString(title)

	for _, stock := range stocks {
		entry := formatStockEntry(stock)
		candidate := current.String()
		candidate += "\n\n" + entry

		if len(candidate) > 1800 && current.Len() > len(title) {
			chunks = append(chunks, current.String())
			current.Reset()
			current.WriteString(title)
			current.WriteString("\n\n")
			current.WriteString(entry)
			continue
		}

		if current.Len() == len(title) {
			current.WriteString("\n\n")
		} else {
			current.WriteString("\n\n")
		}
		current.WriteString(entry)
	}

	if current.Len() > 0 {
		chunks = append(chunks, current.String())
	}

	return chunks
}

func formatStockEntry(stock types.Stock) string {
	name := escapeDiscordText(stock.Name)
	if name != "" {
		name = " — " + name
	}

	return fmt.Sprintf(
		"📈 **%s**%s\n💵 Last sale: %s | 📊 Change: %s | 🏦 Market cap: %s",
		stock.Symbol,
		name,
		fallbackText(stock.LastSale),
		fallbackText(stock.PctChange),
		fallbackText(stock.MarketCap),
	)
}

func escapeDiscordText(value string) string {
	replacer := strings.NewReplacer(
		`\\`, `\\\\`,
		`*`, `\\*`,
		`_`, `\\_`,
		`~`, `\\~`,
		"`", "\\`",
		"|", "\\|",
		">", "\\>",
	)
	return replacer.Replace(strings.TrimSpace(value))
}

func fallbackText(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "N/A"
	}

	return value
}

func OptionString(options []*discordgo.ApplicationCommandInteractionDataOption, name string) string {
	for _, option := range options {
		if option == nil || option.Name != name {
			continue
		}

		if value, ok := option.Value.(string); ok {
			return value
		}

		return fmt.Sprint(option.Value)
	}

	return ""
}
