package auth

import (
	"testing"

	"github.com/google/uuid"
	"seasonschedule/internal/models"
)

func TestHashPassword(t *testing.T) {
	password := "mysecurepassword123"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !CheckPasswordHash(password, hash) {
		t.Errorf("Expected password to match hash")
	}

	if CheckPasswordHash("wrongpassword", hash) {
		t.Errorf("Expected wrong password to fail")
	}
}

func TestGenerateAndValidateToken(t *testing.T) {
	user := models.User{ID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"), Username: "testuser", IsAdmin: false}
	token, err := GenerateToken(user)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	claims, err := ValidateToken(token)
	if err != nil {
		t.Fatalf("Expected token to be valid, got %v", err)
	}

	if claims.Username != user.Username {
		t.Errorf("Expected username %s, got %s", user.Username, claims.Username)
	}
}

func TestValidateToken_Invalid(t *testing.T) {
	_, err := ValidateToken("invalid.token.string")
	if err == nil {
		t.Errorf("Expected error for invalid token")
	}
}
