package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type RequestParams struct {
	Body string `json:"body"`
}

type ValidationResponse struct {
	CleanedBody string `json:"cleaned_body"`
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

func validateChirpHandler(res http.ResponseWriter, req *http.Request) {
	var body RequestParams
	defer req.Body.Close()

	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&body)
	if err != nil {
		fmt.Printf("Error decoding chirp request body: %v", err)
		respondWithError(res, http.StatusBadRequest, "Something went wrong")
		return
	}

	if !isChirpValid(body.Body) {
		respondWithError(res, http.StatusBadRequest, "Chirp is too long")
		return
	}

	respondWithJSON(res, http.StatusOK, ValidationResponse{ CleanedBody: sanitizeChirp(body.Body) })
}
