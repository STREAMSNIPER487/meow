package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

type LogMessage struct {
	Identifier    string
	URL           string
	At            time.Time
	Method        string
	StateExpected uint16
	StateResult   uint16
	Success       bool
	TookSecs      float64
}

func (l LogMessage) String() string {
	return fmt.Sprintf("accessed endpoint %s [URL: %s] at %v with HTTP method %s; "+
		"expected status %d, got status %d (success: %v) in %.4f seconds.",
		l.Identifier, l.URL, l.At, l.Method,
		l.StateExpected, l.StateResult, l.Success, l.TookSecs)
}

func getRedisClient() *redis.Client {
	url := os.Getenv("REDIS_URL")
	if url == "" {
		url = "localhost:6379"
	}

	return redis.NewClient(&redis.Options{
		Addr:     url,
		Password: "",
		DB:       0,
	})
}

func saveLogMessageToRedis(rdb *redis.Client, log LogMessage) error {
	ctx := context.Background()
	key := fmt.Sprintf("log:%s", log.Identifier)
	data := map[string]interface{}{
		"Identifier":    log.Identifier,
		"URL":           log.URL,
		"At":            log.At.Format(time.RFC3339),
		"Method":        log.Method,
		"StateExpected": log.StateExpected,
		"StateResult":   log.StateResult,
		"Success":       log.Success,
		"TookSecs":      log.TookSecs,
	}
	return rdb.HSet(ctx, key, data).Err()
}

func main() {
	rdb := getRedisClient()
	defer rdb.Close()

	frickelbude := LogMessage{
		Identifier:    "frickelbude",
		URL:           "https://code.frickelbude.ch/api/v1/version",
		At:            time.Now(),
		Method:        "GET",
		StateExpected: 200,
		StateResult:   404,
		Success:       false,
		TookSecs:      0.132,
	}
	amazon := LogMessage{
		Identifier:    "amazon.de",
		URL:           "https://www.amazon.de/",
		At:            time.Now(),
		Method:        "GET",
		StateExpected: 200,
		StateResult:   200,
		Success:       true,
		TookSecs:      0.0012,
	}

	fmt.Println(frickelbude)
	fmt.Println(amazon)

	if err := saveLogMessageToRedis(rdb, frickelbude); err != nil {
		fmt.Printf("Failed to save frickelbude log: %v\n", err)
	}
	if err := saveLogMessageToRedis(rdb, amazon); err != nil {
		fmt.Printf("Failed to save amazon log: %v\n", err)
	}
}
