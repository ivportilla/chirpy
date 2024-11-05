package main

import (
	"encoding/json"
	"net/http"
)

type JsonErrorResponse struct {
	Error string `json:"error"`
}

func respondWithError(res http.ResponseWriter, code int, msg string) {
	data, _ := json.Marshal(JsonErrorResponse{Error: msg})
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(code)
	res.Write([]byte(data))
}

func respondWithJSON(res http.ResponseWriter, code int, payload interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.WriteHeader(code)
	res.Header().Set("Content-Type", "application/json")
	res.Write([]byte(data))
}