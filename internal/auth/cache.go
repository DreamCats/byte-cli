package auth

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/DreamCats/byte-cli/internal/config"
)

const (
	tokenTTL    = time.Hour
	tokenBuffer = 5 * time.Minute
)

type cachedToken struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

func cachePath(region string) string {
	return filepath.Join(config.ConfigDir(), "token_cache", region+".json")
}

func getCached(region string) *cachedToken {
	data, err := os.ReadFile(cachePath(region))
	if err != nil {
		return nil
	}
	var token cachedToken
	if err := json.Unmarshal(data, &token); err != nil {
		return nil
	}
	return &token
}

func setCached(region, token string) error {
	dir := filepath.Dir(cachePath(region))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	payload := cachedToken{Token: token, ExpiresAt: time.Now().Add(tokenTTL)}
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(cachePath(region), data, 0o600)
}

func IsTokenValid(region string) bool {
	cached := getCached(region)
	return cached != nil && time.Now().Before(cached.ExpiresAt.Add(-tokenBuffer))
}
