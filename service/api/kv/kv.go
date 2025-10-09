package kv

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	// SessionTTL defines the session expiration time (7 days)
	SessionTTL = 7 * 24 * time.Hour
)

type RedisConfig struct {
	Address  string `json:"address"`
	Password string `json:"password"`
	DB       int    `json:"db"`
}

var client redis.Client

func Connect(config RedisConfig) error {
	client = *redis.NewClient(&redis.Options{
		Addr:     config.Address,
		Password: config.Password,
		DB:       config.DB,
	})

	return nil
}

func Close() {
	client.Close()
}

// generateSessionID creates a secure random session ID
func generateSessionID() (string, error) {
	bytes := make([]byte, 32) // 256 bits
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// CreateSession creates a new session for the given username with 7-day TTL
func CreateSession(username string) (string, error) {
	sessionID, err := generateSessionID()
	if err != nil {
		return "", fmt.Errorf("failed to generate session ID: %v", err)
	}

	ctx := context.Background()
	sessionKey := fmt.Sprintf("session:%s", sessionID)

	// Store username directly as string value with TTL
	err = client.Set(ctx, sessionKey, username, SessionTTL).Err()
	if err != nil {
		return "", fmt.Errorf("failed to store session: %v", err)
	}

	return sessionID, nil
}

// GetSession retrieves the username for a given session ID
func GetSession(sessionID string) (string, error) {
	ctx := context.Background()
	sessionKey := fmt.Sprintf("session:%s", sessionID)

	username, err := client.Get(ctx, sessionKey).Result()
	if err != nil {
		if err == redis.Nil {
			return "", fmt.Errorf("session not found")
		}
		return "", fmt.Errorf("failed to retrieve session: %v", err)
	}

	return username, nil
}

// DeleteSession removes a session from storage
func DeleteSession(sessionID string) error {
	ctx := context.Background()
	sessionKey := fmt.Sprintf("session:%s", sessionID)

	err := client.Del(ctx, sessionKey).Err()
	if err != nil {
		return fmt.Errorf("failed to delete session: %v", err)
	}

	return nil
}

// RefreshSession extends the TTL of an existing session to 7 days from now
func RefreshSession(sessionID string) error {
	ctx := context.Background()
	sessionKey := fmt.Sprintf("session:%s", sessionID)

	result, err := client.Expire(ctx, sessionKey, SessionTTL).Result()
	if err != nil {
		return fmt.Errorf("failed to refresh session: %v", err)
	}

	if !result {
		return fmt.Errorf("session not found")
	}

	return nil
}
