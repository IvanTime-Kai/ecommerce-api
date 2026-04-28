package service

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func generateAccessToken(userID uuid.UUID, secret string, ttlMinutes int) (string, error) {

	claims := jwt.MapClaims{
		"sub": userID.String(),
		"exp": time.Now().Add(time.Duration(ttlMinutes) * time.Minute).Unix(),
		"iat": time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func generateRefreshToken(secret string) (string, string, error) {
	rawBytes := make([]byte, 32)
	_, err := rand.Read(rawBytes)
	if err != nil {
		return "", "", err
	}

	rawToken := hex.EncodeToString(rawBytes)
	hashedToken := hashToken(rawToken, secret)

	return rawToken, hashedToken, nil
}

func hashToken(token string, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(token))
	hashedToken := hex.EncodeToString(mac.Sum(nil))

	return hashedToken
}
