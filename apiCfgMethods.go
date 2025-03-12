package main

import (
	"GoServer/internal/database"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	auth "GoServer/internal/auth"

	"github.com/google/uuid"
)

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) middlewareMetricsReset(w http.ResponseWriter, r *http.Request) {
	if cfg.Platform != "dev" {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	if cfg.Platform != "dev" {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	cfg.fileserverHits.Store(0)

	if err := cfg.DB.DeleteAllUsers(r.Context()); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error deleting all users", err)
		return
	}

	if err := cfg.DB.DeleteAllUsers(r.Context()); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error deleting all users", err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `
	<html>
		<body>
			<h1>Welcome, Chirpy Admin</h1>
			<p>Chirpy has been visited %d times!</p>
		</body>
	</html>
	`, cfg.fileserverHits.Load())
}

func (cfg *apiConfig) createUser(w http.ResponseWriter, r *http.Request) {
	userUnmarshallInto := User{}
	if err := json.NewDecoder(r.Body).Decode(&userUnmarshallInto); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error unmarshalling User", err)
	}

	hashedPassword, err := auth.HashPassword(userUnmarshallInto.HashedPassword)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error hashing password", err)
		return
	}

	user, err := cfg.DB.CreateUser(r.Context(), database.CreateUserParams{Email: userUnmarshallInto.Email, HashedPassword: hashedPassword})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error creating User", err)
	}

	respondWithJSON(w, http.StatusCreated, User{ID: user.ID, CreatedAt: user.CreatedAt, UpdatedAt: user.UpdatedAt, Email: user.Email})
}

func (cfg *apiConfig) createChirp(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Error getting token", err)
		return
	}

	uuid, err := auth.ValidateJWT(token, cfg.Secret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Error validating token", err)
		return
	}

	chirpUnmarshallInto := Chirp{}
	if err := json.NewDecoder(r.Body).Decode(&chirpUnmarshallInto); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error unmarshalling Chirp", err)
	}
	const maxChirpLen = 140
	if len(chirpUnmarshallInto.Body) > maxChirpLen {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}

	chirp, err := cfg.DB.CreateChirp(r.Context(), database.CreateChirpParams{Body: chirpUnmarshallInto.Body, UserID: uuid})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error creating chirp", err)
	}

	chirpResponse := Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserId:    chirp.UserID,
	}

	respondWithJSON(w, http.StatusCreated, chirpResponse)
}

func (cfg *apiConfig) getChirps(w http.ResponseWriter, r *http.Request) {
	chirps, err := cfg.DB.GetChirps(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't get all chirps", err)
	}
	chirpsResponse := make([]Chirp, len(chirps))
	for i, chirp := range chirps {
		chirpsResponse[i] = Chirp{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserId:    chirp.UserID,
		}
	}
	respondWithJSON(w, http.StatusOK, chirpsResponse)
}

func (cfg *apiConfig) getChirpByID(w http.ResponseWriter, r *http.Request) {

	chirpIDStr := r.PathValue("chirpID")

	chirpID, err := uuid.Parse(chirpIDStr)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Invalid chirp ID format", err)
		return
	}

	chirp, err := cfg.DB.GetChirpByID(r.Context(), chirpID)
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusNotFound, "Chirp not found", nil)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Error retrieving chirp", err)
		return
	}

	chirpResponse := Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserId:    chirp.UserID,
	}

	respondWithJSON(w, http.StatusOK, chirpResponse)
}

func (cfg *apiConfig) login(w http.ResponseWriter, r *http.Request) {
	loginParams := struct {
		Email          string `json:"email"`
		HashedPassword string `json:"password"`
	}{
		Email:          "",
		HashedPassword: "",
	}
	if err := json.NewDecoder(r.Body).Decode(&loginParams); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error unmarshalling login parameters", err)
		return
	}

	user, err := cfg.DB.GetUserByEmail(r.Context(), loginParams.Email)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Email not found in DB", err)
		return
	}

	if err = auth.CheckPasswordHash(loginParams.HashedPassword, user.HashedPassword); err != nil {
		respondWithError(w, http.StatusUnauthorized, "WRONG", err)
		return
	}

	token, err := auth.MakeJWT(user.ID, cfg.Secret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error making JWT", err)
		return
	}

	refresh_token := auth.MakeRefreshToken()

	cfg.DB.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{Token: refresh_token, UserID: user.ID, ExpiresAt: time.Now().Add(time.Hour * 1440)})

	respondWithJSON(w, http.StatusOK, User{ID: user.ID, CreatedAt: user.CreatedAt, UpdatedAt: user.UpdatedAt, Email: user.Email, AccessToken: token, RefreshToken: refresh_token})

}

func (cfg *apiConfig) refresh(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Error getting token", err)
		return
	}

	refreshTokenFromDB, err := cfg.DB.GetRefreshToken(r.Context(), refreshToken)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusUnauthorized, "Invalid refresh token", err)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Error getting refresh token", err)
		return
	}

	if time.Now().After(refreshTokenFromDB.ExpiresAt) {
		respondWithError(w, http.StatusUnauthorized, "Refresh token expired", nil)
		return
	}

	if refreshTokenFromDB.RevokedAt.Valid {
		respondWithError(w, http.StatusUnauthorized, "Refresh token revoked", nil)
		return
	}

	userID := refreshTokenFromDB.UserID
	accessToken, err := auth.MakeJWT(userID, cfg.Secret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error making JWT", err)
		return
	}

	respondWithJSON(w, http.StatusOK, struct {
		Token string `json:"token"`
	}{Token: accessToken})

}

func (cfg *apiConfig) revoke(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Error getting token", err)
		return
	}

	_, err = cfg.DB.GetRefreshToken(r.Context(), refreshToken)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusUnauthorized, "Invalid refresh token", err)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Error getting refresh token", err)
		return
	}

	if err := cfg.DB.RevokeRefreshToken(r.Context(), refreshToken); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error deleting refresh token", err)
		return
	}

	respondWithJSON(w, http.StatusNoContent, nil)

}
