package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

type CreateUserReq struct {
	Email string `json:"email"`
}

func createUserHandler(cfg *apiConfig) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		var body CreateUserReq
		err := json.NewDecoder(req.Body).Decode(&body)
		if err != nil {
			respondWithError(res, http.StatusBadRequest, "Error decoding body, email field expected")
			return
		}

		user, err := cfg.dbQueries.CreateUser(req.Context(), body.Email)
		if err != nil {
			respondWithError(res, http.StatusInternalServerError, "Error creating user")
			return
		}

		userToReturn := User{
			ID:        user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Email:     user.Email,
		}
		respondWithJSON(res, http.StatusCreated, userToReturn)
	}
}

func createAllUsersHandler(cfg *apiConfig) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		if cfg.platform != "dev" {
			res.WriteHeader(http.StatusForbidden)
			return
		}
		
		err := cfg.dbQueries.DeleteAllUsers(req.Context())
		if err != nil {
			log.Printf("Error deleting users: %v", err)
			respondWithError(res, http.StatusInternalServerError, "An error occurred deleting all the users")
			return
		}
		res.WriteHeader(http.StatusOK)
	}
}