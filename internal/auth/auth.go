package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("error hashing password")
	}
	return string(hashed), nil
}

func CheckPasswordHash(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	now := time.Now()
	expires := now.Add(expiresIn)
	claims := jwt.RegisteredClaims{
		Issuer:    "chirpy",
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(expires),
		Subject:   userID.String(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(tokenSecret))
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	var claims jwt.RegisteredClaims
	parsed, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (any, error) {
		return []byte(tokenSecret), nil
	})
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("error parsing jwt: %w", err)
	}

	userIdRaw, err := parsed.Claims.GetSubject()
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("error getting user id from token: %w", err)
	}

	userId, err := uuid.Parse(userIdRaw)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("error converting user id to uuid: %w", err)
	}

	return userId, nil
}

func GetBearerToken(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return "", fmt.Errorf("invalid authtorization header")
	}
	return strings.TrimPrefix(authHeader, "Bearer "), nil
}

func MakeRefreshToken() (string, error) {
	output := make([]byte, 32)
	_, err := rand.Read(output)
	if err != nil {
		return "", fmt.Errorf("error generating refresh token")
	}
	return hex.EncodeToString(output), nil
}
