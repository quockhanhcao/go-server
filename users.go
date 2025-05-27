package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/quockhanhcao/go-server/internal/auth"
	"github.com/quockhanhcao/go-server/internal/database"
)

type createUserBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (apiCfg *apiConfig) createUsersHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	jsonBody := createUserBody{}
	err := decoder.Decode(&jsonBody)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't decode params", err)
		return
	}
	hashPassword, err := auth.HashPassword(jsonBody.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to hash password", err)
	}
	user, err := apiCfg.db.CreateUser(r.Context(), database.CreateUserParams{Email: jsonBody.Email, Password: hashPassword})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create user", err)
		return
	}
	response := User{
		ID:        user.ID.String(),
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
	respondWithJSON(w, http.StatusCreated, response)
}

func (apiCfg *apiConfig) deleteAllUsersHandler(w http.ResponseWriter, r *http.Request) {
	if apiCfg.platform != "dev" {
		respondWithError(w, http.StatusForbidden, "This endpoint is not available in production", nil)
	}
	err := apiCfg.db.DeleteUsers(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to delete users", err)
		return
	}
}
