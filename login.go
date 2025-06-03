package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/quockhanhcao/go-server/internal/auth"
	"github.com/quockhanhcao/go-server/internal/database"
)

func (apiCfg *apiConfig) loginHandler(w http.ResponseWriter, r *http.Request) {
	type loginBody struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	type response struct {
		User
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
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
	accessTokenExpirationTime := time.Hour
	accessToken, err := auth.MakeJWT(user.ID, apiCfg.jwtSecret, time.Duration(accessTokenExpirationTime))
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unable to generate jwt token", err)
		return
	}
	refreshToken, _ := auth.MakeRefreshToken()
	apiCfg.db.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token:  refreshToken,
		UserID: user.ID,
	})

	respondWithJSON(w, http.StatusOK, response{
		User: User{
			ID:        user.ID.String(),
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Email:     user.Email,
		},
		Token:        accessToken,
		RefreshToken: refreshToken,
	})
}

func (apiCfg *apiConfig) refreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	type response struct {
		Token string `json:"token"`
	}
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid refresh token", err)
		return
	}
	refreshTokenDB, err := apiCfg.db.GetRefreshTokenByToken(r.Context(), refreshToken)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid refresh token", err)
		return
	}
	if refreshTokenDB.ExpiresAt.Before(time.Now()) {
		respondWithError(w, http.StatusUnauthorized, "Refresh token expired", nil)
		return
	}
	if refreshTokenDB.RevokedAt.Valid {
		respondWithError(w, http.StatusUnauthorized, "Refresh token has been revoked", nil)
		return
	}
	accessToken, err := auth.MakeJWT(refreshTokenDB.UserID, apiCfg.jwtSecret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to generate jwt token", err)
		return
	}
	respondWithJSON(w, http.StatusOK, response{
		Token: accessToken,
	})
}

func (apiCfg *apiConfig) revokeRefreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid refresh token", err)
		return
	}
	refreshTokenDB, err := apiCfg.db.GetRefreshTokenByToken(r.Context(), refreshToken)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Refresh token not found", err)
		return
	}
	if refreshTokenDB.ExpiresAt.Before(time.Now()) {
		respondWithError(w, http.StatusUnauthorized, "Refresh token expired", nil)
		return
	}
	apiCfg.db.RevokeRefreshToken(r.Context(), refreshTokenDB.Token)
	respondWithJSON(w, http.StatusNoContent, struct{}{})
}
