package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ivportilla/chirpy/internal/auth"
)

type LoginRequest struct {
	Email            string `json:"email"`
	Password         string `json:"password"`
	ExpiresInSeconds *int   `json:"expires_in_seconds"`
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
		if reqBody.ExpiresInSeconds != nil {
			expiresIn = time.Duration(*reqBody.ExpiresInSeconds) * time.Second
		}
		token, err := auth.MakeJWT(user.ID, cfg.authSecret, expiresIn)
		if err != nil {
			fmt.Printf("Error creating JWT token: %v", err)
			respondWithError(res, http.StatusInternalServerError, "Error creating JWT token")
			return
		}

		userResponse := ToResponseUser(user)
		userResponse.Token = token

		respondWithJSON(res, http.StatusOK, userResponse)
	}
}
