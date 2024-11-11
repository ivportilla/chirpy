package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ivportilla/chirpy/internal/auth"
	"github.com/ivportilla/chirpy/internal/database"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (cfg *apiConfig) withAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		token, err := auth.GetBearerToken(req.Header)
		if err != nil {
			respondWithError(res, http.StatusUnauthorized, "Unauthorized")
			return
		}

		userID, err := auth.ValidateJWT(token, cfg.authSecret)
		if err != nil {
			respondWithError(res, http.StatusUnauthorized, "Unauthorized")
			return
		}

		ctx := context.WithValue(req.Context(), "user_id", userID.String())

		next.ServeHTTP(res, req.WithContext(ctx))
	})
}

func loginHandler(cfg *apiConfig) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		var reqBody LoginRequest
		err := json.NewDecoder(req.Body).Decode(&reqBody)
		if err != nil {
			respondWithError(res, http.StatusBadRequest, "Error decoding body, it should contain email and password")
			return
		}

		user, err := cfg.dbQueries.GetUser(req.Context(), reqBody.Email)
		if err != nil {
			if err == sql.ErrNoRows {
				respondWithError(res, http.StatusNotFound, "User not found")
				return
			}
			respondWithError(res, http.StatusInternalServerError, "Error getting user")
			return
		}

		err = auth.CheckPasswordHash(reqBody.Password, user.HashedPassword)
		if err != nil {
			respondWithError(res, http.StatusUnauthorized, "Incorrect email or password")
			return
		}

		expiresIn := 3600 * time.Second
		token, err := auth.MakeJWT(user.ID, cfg.authSecret, expiresIn)
		if err != nil {
			fmt.Printf("Error creating JWT token: %v", err)
			respondWithError(res, http.StatusInternalServerError, "Error creating JWT token")
			return
		}

		refreshToken, err := auth.MakeRefreshToken()
		if err != nil {
			fmt.Printf("Error creating refresh token: %v", err)
			respondWithError(res, http.StatusInternalServerError, "Error creating refresh token")
			return
		}
		err = cfg.dbQueries.CreateRefreshToken(req.Context(), database.CreateRefreshTokenParams{Token: refreshToken, UserID: user.ID, ExpiresAt: time.Now().Add(60 * 24 * time.Hour)})
		if err != nil {
			fmt.Printf("Error creating saving refresh token: %v", err)
			respondWithError(res, http.StatusInternalServerError, "Error creating refresh token")
			return
		}

		userResponse := ToResponseUser(user)
		userResponse.Token = token
		userResponse.RefreshToken = refreshToken

		respondWithJSON(res, http.StatusOK, userResponse)
	}
}

type RefreshTokenResponse struct {
	Token string `json:"token"`
}

func refreshTokenHandler(cfg *apiConfig) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		token, err := auth.GetBearerToken(req.Header)
		if err != nil {
			fmt.Printf("Error extracting refresh token from auth header: %v", err)
			respondWithError(res, http.StatusUnauthorized, "Unauthorized")
			return
		}

		refreshToken, err := cfg.dbQueries.GetRefreshToken(req.Context(), token)
		if err != nil {
			fmt.Printf("Error fetching refresh token: %v", err)
			respondWithError(res, http.StatusUnauthorized, "Unauthorized")
			return
		}

		expired := time.Since(refreshToken.ExpiresAt) >= 0 || refreshToken.RevokedAt.Valid
		if expired {
			respondWithError(res, http.StatusUnauthorized, "Unauthorized")
			return
		}

		newToken, err := auth.MakeJWT(refreshToken.UserID, cfg.authSecret, 3600*time.Second)
		if err != nil {
			respondWithError(res, http.StatusInternalServerError, "Error creating JWT token")
			return
		}
		respondWithJSON(res, http.StatusOK, RefreshTokenResponse{Token: newToken})
	}
}

func revokeRefreshToken(cfg *apiConfig) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		token, err := auth.GetBearerToken(req.Header)
		if err != nil {
			fmt.Printf("Error extracting refresh token from auth header: %v", err)
			respondWithError(res, http.StatusUnauthorized, "Unauthorized")
			return
		}

		err = cfg.dbQueries.RevokeRefreshToken(req.Context(), token)
		if err != nil {
			fmt.Printf("Error revoking token: %v", err)
			respondWithError(res, http.StatusUnauthorized, "Unauthorized")
			return
		}

		respondWithJSON(res, http.StatusNoContent, nil)
	}
}
