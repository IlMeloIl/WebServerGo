package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

func handlerChirpValidate(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}
	type respVals struct {
		CleanedBody string `json:"cleaned_body"`
	}
	params := parameters{}
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}
	const maxChirpLen = 140
	if len(params.Body) > maxChirpLen {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}
	cleanedBody := cleanBody(params.Body)
	respondWithJSON(w, http.StatusOK, respVals{
		CleanedBody: cleanedBody,
	})
}

func cleanBody(s string) string {
	bannedWords := map[string]bool{
		"kerfuffle": true,
		"sharbert":  true,
		"fornax":    true,
	}
	words := strings.Split(s, " ")
	for i, word := range words {
		if bannedWords[strings.ToLower(word)] {
			words[i] = "****"
		}
	}
	return strings.Join(words, " ")
}
