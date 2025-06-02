package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/quockhanhcao/go-server/internal/auth"
)

func (apiCfg *apiConfig) loginHandler(w http.ResponseWriter, r *http.Request) {
	type loginBody struct {
		Email            string `json:"email"`
		Password         string `json:"password"`
		ExpiresInSeconds int    `json:"expires_in_seconds"`
	}
	decoder := json.NewDecoder(r.Body)
	params := loginBody{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't decode params", err)
		return
	}
	user, err := apiCfg.db.GetUserByEmail(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", err)
		return
	}
	if err := auth.CheckPasswordHash(user.Password, params.Password); err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", err)
		return
	}

	// set expire time for token
	expirationTime := time.Hour
	if params.ExpiresInSeconds > 0 && params.ExpiresInSeconds < 3600 {
		expirationTime = time.Duration(params.ExpiresInSeconds) * time.Second
	}
	accessToken, err := auth.MakeJWT(user.ID, apiCfg.jwtSecret, time.Duration(expirationTime))
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unable to generate jwt token", err)
		return
	}

	response := struct {
		ID        string    `json:"id"`
		Email     string    `json:"email"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Token     string    `json:"token"`
	}{
		ID:        user.ID.String(),
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Token:     accessToken,
	}
	respondWithJSON(w, http.StatusOK, response)
}
