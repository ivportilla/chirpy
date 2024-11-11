package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/ivportilla/chirpy/internal/auth"
	"github.com/ivportilla/chirpy/internal/database"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	Token        string    `json:"token,omitempty"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	IsChirpyRed  bool      `json:"is_chirpy_red"`
}

type CreateUserReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UpdateUserReq struct {
	CreateUserReq
}

func ToResponseUser(dbUser database.User) User {
	return User{
		ID:          dbUser.ID,
		CreatedAt:   dbUser.CreatedAt,
		UpdatedAt:   dbUser.UpdatedAt,
		Email:       dbUser.Email,
		IsChirpyRed: dbUser.IsChirpyRed,
	}
}

func createUserHandler(cfg *apiConfig) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		defer req.Body.Close()
		var body CreateUserReq
		err := json.NewDecoder(req.Body).Decode(&body)
		if err != nil {
			respondWithError(res, http.StatusBadRequest, "Error decoding body, email field expected")
			return
		}

		pwd, err := auth.HashPassword(body.Password)
		if err != nil {
			fmt.Printf("Error generating password hash: %v\n", err)
			respondWithError(res, http.StatusBadRequest, "Error generating password hash")
			return
		}
		user, err := cfg.dbQueries.CreateUser(req.Context(), database.CreateUserParams{Email: body.Email, HashedPassword: pwd})
		if err != nil {
			respondWithError(res, http.StatusInternalServerError, "Error creating user")
			return
		}

		respondWithJSON(res, http.StatusCreated, ToResponseUser(user))
	}
}

func updateUserHandler(cfg *apiConfig) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		defer req.Body.Close()
		var body UpdateUserReq
		err := json.NewDecoder(req.Body).Decode(&body)
		if err != nil {
			respondWithError(res, http.StatusBadRequest, "Error decoding body, email and password fields expected")
			return
		}

		userID := uuid.MustParse(req.Context().Value("user_id").(string))
		pwd, err := auth.HashPassword(body.Password)
		if err != nil {
			fmt.Printf("Error generating password hash: %v\n", err)
			respondWithError(res, http.StatusBadRequest, "Error generating password hash")
			return
		}
		user, err := cfg.dbQueries.UpdateUser(req.Context(), database.UpdateUserParams{Email: body.Email, HashedPassword: pwd, ID: userID})
		if err != nil {
			respondWithError(res, http.StatusInternalServerError, "Error updating user")
			return
		}

		respondWithJSON(res, http.StatusOK, ToResponseUser(user))
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
