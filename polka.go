package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/ivportilla/chirpy/internal/auth"
)

type UpgradeRequestDTO struct {
	Event string `json:"event"`
	Data  struct {
		UserID string `json:"user_id"`
	} `json:"data"`
}

const USER_UPGRADED = "user.upgraded"

func handleUserUpgrade(cfg *apiConfig) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		defer req.Body.Close()
		var reqData UpgradeRequestDTO

		apiKey, err := auth.GetAPIKey(req.Header)
		if err != nil {
			respondWithError(res, http.StatusUnauthorized, "Unauthorized")
			return
		}

		if apiKey != cfg.apiKey {
			respondWithError(res, http.StatusUnauthorized, "Unauthorized")
			return
		}

		err = json.NewDecoder(req.Body).Decode(&reqData)
		if err != nil {
			respondWithError(res, http.StatusBadRequest, "Error decoding request body")
			return
		}

		if reqData.Event != USER_UPGRADED {
			respondWithJSON(res, http.StatusNoContent, nil)
			return
		}

		userID, err := uuid.Parse(reqData.Data.UserID)
		if err != nil {
			respondWithError(res, http.StatusBadRequest, "Error decoding user id, it must be a uuid")
			return
		}

		_, err = cfg.dbQueries.UpgradeUser(req.Context(), userID)
		if err != nil {
			fmt.Printf("Error upgrading user: %v", err)
			if err == sql.ErrNoRows {
				respondWithError(res, http.StatusNotFound, "User not found")
				return
			}

			respondWithError(res, http.StatusInternalServerError, "Error upgrading user")
			return
		}

		respondWithJSON(res, http.StatusNoContent, nil)
	}
}
