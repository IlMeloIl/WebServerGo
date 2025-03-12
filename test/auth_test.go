package auth

import (
	auth "GoServer/internal/auth"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestMakeJWT(t *testing.T) {
	userID := uuid.New()
	tokenSecret := "mysecret"
	expiresIn := time.Hour

	token, err := auth.MakeJWT(userID, tokenSecret, expiresIn)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if token == "" {
		t.Fatalf("expected a token, got an empty string")
	}
}

func TestValidateJWT(t *testing.T) {
	userID := uuid.New()
	tokenSecret := "mysecret"
	expiresIn := time.Hour

	token, err := auth.MakeJWT(userID, tokenSecret, expiresIn)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	validatedUserID, err := auth.ValidateJWT(token, tokenSecret)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if validatedUserID != userID {
		t.Fatalf("expected userID %v, got %v", userID, validatedUserID)
	}
}

func TestValidateExpiredJWT(t *testing.T) {
	userID := uuid.New()
	tokenSecret := "mysecret"
	expiresIn := -time.Hour

	token, err := auth.MakeJWT(userID, tokenSecret, expiresIn)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	_, err = auth.ValidateJWT(token, tokenSecret)
	if err == nil {
		t.Fatalf("expected an error, got none")
	}
}

func TestValidateJWTWithWrongSecret(t *testing.T) {
	userID := uuid.New()
	tokenSecret := "mysecret"
	wrongSecret := "wrongsecret"
	expiresIn := time.Hour

	token, err := auth.MakeJWT(userID, tokenSecret, expiresIn)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	_, err = auth.ValidateJWT(token, wrongSecret)
	if err == nil {
		t.Fatalf("expected an error, got none")
	}
}
