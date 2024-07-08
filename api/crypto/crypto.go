package crypto

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log"
	"prism/config"
	"prism/models"
	"time"
)

func HashAPIKey(key string) (string, error) {

	hmacSecret := config.AppConfig.Secrets["HMAC_SECRET_KEY"]

	if hmacSecret == "" {
		log.Printf("HMAC_SECRET_KEY is empty")
		return "", errors.New("HMAC_SECRET_KEY is empty")
	}
	h := hmac.New(sha256.New, []byte(hmacSecret))
	h.Write([]byte(key))
	return hex.EncodeToString(h.Sum(nil)), nil
}

func CreateAPIKey(apiKeyModel *models.APIKey) (*models.APIKey, error) {

	apiKey := generateAPIKey()
	if apiKeyModel.ExpiryAt.IsZero() {
		apiKeyModel.ExpiryAt = time.Now().AddDate(0, 0, 90) // 90 days from now
	}

	hashedAPIKey, err := HashAPIKey(apiKey)

	if err != nil {
		return nil, err
	}

	newKey := &models.APIKey{
		Name:         apiKeyModel.Name,
		ExpiryAt:     apiKeyModel.ExpiryAt,
		HashedAPIKey: hashedAPIKey,
		Email:        apiKeyModel.Email,
		ApiKey:       apiKey,
	}

	return newKey, nil
}

const allowedChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

func generateAPIKey() string {
	b := make([]byte, 50)
	for i := range b {
		randomByte := make([]byte, 1)
		_, err := rand.Read(randomByte)
		if err != nil {
			log.Fatalf("failed to generate API key: %v", err)
		}
		b[i] = allowedChars[randomByte[0]%byte(len(allowedChars))]
	}

	return "sk-" + string(b)
}
