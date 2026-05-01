package nasdaq

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"tradingcorpbot/helpers"
	"tradingcorpbot/types"
)

const nasdaqScreenersURL = "https://api.nasdaq.com/api/screener/stocks?tableonly=true&limit=10000"

type Stock struct {
	Symbol    string
	Name      string
	LastSale  string
	NetChange string
	PctChange string
	MarketCap string
}

type nasdaqResponse struct {
	Data struct {
		Table struct {
			Rows []struct {
				Symbol    string `json:"symbol"`
				Name      string `json:"name"`
				LastSale  string `json:"lastsale"`
				NetChange string `json:"netchange"`
				PctChange string `json:"pctchange"`
				MarketCap string `json:"marketCap"`
				URL       string `json:"url"`
			} `json:"rows"`
		} `json:"table"`
	} `json:"data"`
}

func FetchAllStocks(ctx context.Context) ([]types.Stock, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, nasdaqScreenersURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36")
	req.Header.Set("Referer", "https://www.nasdaq.com/market-activity/stocks/screener")
	req.Header.Set("Origin", "https://www.nasdaq.com")

	client := &http.Client{Timeout: 20 * time.Second}
	response, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("nasdaq api returned %s", response.Status)
	}

	var payload nasdaqResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return nil, err
	}

	stocks := make([]types.Stock, 0, len(payload.Data.Table.Rows))
	for _, row := range payload.Data.Table.Rows {
		symbol := strings.ToUpper(strings.TrimSpace(row.Symbol))
		if symbol == "" {
			continue
		}

		if !helpers.IsValidTickerSymbol(symbol) {
			continue
		}

		stocks = append(stocks, types.Stock{
			Symbol:    symbol,
			Name:      strings.TrimSpace(row.Name),
			LastSale:  strings.TrimSpace(row.LastSale),
			NetChange: strings.TrimSpace(row.NetChange),
			PctChange: strings.TrimSpace(row.PctChange),
			MarketCap: strings.TrimSpace(row.MarketCap),
		})
	}

	sort.SliceStable(stocks, func(i, j int) bool {
		return stocks[i].Symbol < stocks[j].Symbol
	})

	return stocks, nil
}
