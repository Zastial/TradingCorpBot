package cache

import (
	"context"
	"encoding/json"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"

	"tradingcorpbot/types"
)

var client *redis.Client

func Init() {
	db, _ := strconv.Atoi(os.Getenv("REDIS_DB"))
	client = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_URL"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       db,
	})
}

// GetStocksFromCache lit la liste depuis Redis.
// Retourne (nil, nil) si le cache est vide (miss).
// Retourne (nil, err) si Redis est indisponible.
func GetStocksFromCache(ctx context.Context) ([]types.Stock, error) {
	key := os.Getenv("CACHE_KEY_STOCKS_ALL")

	val, err := client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var stocks []types.Stock
	if err := json.Unmarshal([]byte(val), &stocks); err != nil {
		return nil, err
	}
	return stocks, nil
}

// SetStocksInCache écrit la liste dans Redis avec le TTL configuré.
func SetStocksInCache(ctx context.Context, stocks []types.Stock) error {
	key := os.Getenv("CACHE_KEY_STOCKS_ALL")
	ttl := os.Getenv("REDIS_TTL")

	seconds, _ := strconv.Atoi(ttl)
	data, err := json.Marshal(stocks)
	if err != nil {
		return err
	}

	return client.Set(ctx, key, data, time.Duration(seconds)*time.Second).Err()
}

// AcquireInteractionLock tente de réserver le traitement d'une interaction pour un seul replica.
func AcquireInteractionLock(ctx context.Context, interactionID string, ttl time.Duration) (bool, error) {
	return client.SetNX(ctx, "interaction:lock:"+interactionID, "1", ttl).Result()
}
