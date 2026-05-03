package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type SessionClaims struct {
	UserID uint  `json:"user_id"`
	Exp    int64 `json:"exp"`
}

func sessionSecret() []byte {
	secret := strings.TrimSpace(os.Getenv("SESSION_SECRET"))
	if secret == "" {
		secret = "dev-only-session-secret"
	}
	return []byte(secret)
}

func SessionTTL() time.Duration {
	hours, err := strconv.Atoi(strings.TrimSpace(os.Getenv("SESSION_TTL_HOURS")))
	if err != nil || hours <= 0 {
		hours = 168
	}
	return time.Duration(hours) * time.Hour
}

func GenerateSessionToken(userID uint) (string, error) {
	claims := SessionClaims{
		UserID: userID,
		Exp:    time.Now().Add(SessionTTL()).Unix(),
	}
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	encodedPayload := base64.RawURLEncoding.EncodeToString(payload)
	signature := signPayload(encodedPayload)
	return fmt.Sprintf("%s.%s", encodedPayload, signature), nil
}

func ValidateSessionToken(token string) (uint, error) {
	token = strings.TrimSpace(token)
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return 0, errors.New("invalid session token format")
	}

	expectedSignature := signPayload(parts[0])
	if !hmac.Equal([]byte(expectedSignature), []byte(parts[1])) {
		return 0, errors.New("invalid session token signature")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return 0, err
	}

	var claims SessionClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return 0, err
	}
	if claims.UserID == 0 {
		return 0, errors.New("empty session user")
	}
	if time.Now().Unix() > claims.Exp {
		return 0, errors.New("session token expired")
	}

	return claims.UserID, nil
}

func signPayload(payload string) string {
	mac := hmac.New(sha256.New, sessionSecret())
	mac.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
