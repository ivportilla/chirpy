package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/ivportilla/chirpy/internal/database"
)

type RequestParams struct {
	Body   string `json:"body"`
	UserID string `json:"user_id"`
}

type ValidationResponse struct {
	CleanedBody string `json:"cleaned_body"`
}

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func isChirpValid(content string) bool {
	return len(content) <= 140
}

var forbiddenWords = []string{"sharbert", "kerfuffle", "fornax"}

func isForbiddenWord(target string) bool {
	for _, fWord := range forbiddenWords {
		if strings.EqualFold(target, fWord) {
			return true
		}
	}
	return false
}

func sanitizeChirp(content string) string {
	words := strings.Split(content, " ")
	resultWords := []string{}

	for _, word := range words {
		if isForbiddenWord(word) {
			resultWords = append(resultWords, "****")
		} else {
			resultWords = append(resultWords, word)
		}
	}

	return strings.Join(resultWords, " ")
}

func toChirp(target database.Chirp) Chirp {
	return Chirp{
		ID:        target.ID,
		CreatedAt: target.CreatedAt,
		UpdatedAt: target.UpdatedAt,
		Body:      target.Body,
		UserID:    target.UserID,
	}
}

func createChirpHandler(cfg *apiConfig) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		var reqBody RequestParams
		defer req.Body.Close()

		decoder := json.NewDecoder(req.Body)
		err := decoder.Decode(&reqBody)
		if err != nil {
			fmt.Printf("Error decoding chirp request body: %v", err)
			respondWithError(res, http.StatusBadRequest, "Error decoding request")
			return
		}

		if !isChirpValid(reqBody.Body) {
			respondWithError(res, http.StatusBadRequest, "Chirp is too long")
			return
		}

		userId, err := uuid.Parse(reqBody.UserID)
		if err != nil {
			fmt.Printf("Error converting user_id to uuid: %v", err)
			respondWithError(res, http.StatusBadRequest, "Error decoding user_id, it is not a valid UUID")
			return
		}

		chirpBody := sanitizeChirp(reqBody.Body)
		chirp, err := cfg.dbQueries.CreateChirp(req.Context(), database.CreateChirpParams{Body: chirpBody, UserID: userId})
		if err != nil {
			fmt.Printf("Error creating chirp: %v", err)
			respondWithError(res, http.StatusInternalServerError, "Error creating chirp")
			return
		}

		respondWithJSON(res, http.StatusCreated, toChirp(chirp))
	}
}

func getAllChirpsHandler(cfg *apiConfig) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		chirps, err := cfg.dbQueries.GetChirps(req.Context())
		if err != nil {
			fmt.Printf("Error fetching chirps: %v", err)
			respondWithError(res, http.StatusInternalServerError, "Error getting chirps")
		}

		response := make([]Chirp, len(chirps))
		for i, chirp := range chirps {
			response[i] = toChirp(chirp)
		}

		respondWithJSON(res, http.StatusOK, response)
	}
}


func getChirp(cfg *apiConfig) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		id := req.PathValue("chirpID")
		parsedId, err := uuid.Parse(id)
		if err != nil {
			fmt.Printf("Error converting chirp id to uuid: %v", err)
			respondWithError(res, http.StatusBadRequest, "Invalid chirp ID, it must be a UUID")
			return
		}
		chirp, err := cfg.dbQueries.GetChirp(req.Context(), parsedId)
		if err != nil {
			if err == sql.ErrNoRows {
				respondWithError(res, http.StatusNotFound, "Chirp not found")
				return
			}
			
			fmt.Printf("Error fetching chirp %s: %v", id, err)
			respondWithError(res, http.StatusInternalServerError, "Error getting chirp")
			return
		}

		respondWithJSON(res, http.StatusOK, toChirp(chirp))
	}
}