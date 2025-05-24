package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

type parameters struct {
	Body string `json:"body"`
}

type returnVals struct {
	CleanedBody string `json:"cleaned_body"`
}

func validateChirpHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	jsonBody := parameters{}
	err := decoder.Decode(&jsonBody)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Something went wrong", err)
		return
	}
	const maxChirpLength = 140
	if len(jsonBody.Body) > maxChirpLength {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}
	cleanedBody := replaceBadWords(jsonBody.Body)
	respondWithJSON(w, http.StatusOK, returnVals{CleanedBody: cleanedBody})
}

func replaceBadWords(body string) string {
	profaneWords := map[string]interface{}{
		"kerfuffle": nil,
		"sharbert":  nil,
		"fornax":    nil,
	}

	words := strings.Split(body, " ")
	for i, word := range words {
		loweredWord := strings.ToLower(word)
		if _, exists := profaneWords[loweredWord]; exists {
			words[i] = "****"
		}
	}
	cleanedBody := strings.Join(words, " ")
	return cleanedBody
}
