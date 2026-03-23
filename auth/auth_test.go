package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const testSecret = "test-secret-key"

func TestGenerateToken_Success(t *testing.T) {
	ta := NewTokenAuth(testSecret)

	token, err := ta.GenerateToken(42)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}
}

func TestGenerateToken_DifferentUsers(t *testing.T) {
	ta := NewTokenAuth(testSecret)

	t1, _ := ta.GenerateToken(1)
	t2, _ := ta.GenerateToken(2)

	if t1 == t2 {
		t.Fatal("tokens for different users should be different")
	}
}

func TestValidateToken_Success(t *testing.T) {
	ta := NewTokenAuth(testSecret)

	token, _ := ta.GenerateToken(42)
	userID, err := ta.ValidateToken(token)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if userID != 42 {
		t.Fatalf("expected userID=42, got %d", userID)
	}
}

func TestValidateToken_RoundTrip(t *testing.T) {
	ta := NewTokenAuth(testSecret)

	for _, id := range []int{1, 100, 999, 0} {
		token, err := ta.GenerateToken(id)
		if err != nil {
			t.Fatalf("generate failed for id=%d: %v", id, err)
		}
		got, err := ta.ValidateToken(token)
		if err != nil {
			t.Fatalf("validate failed for id=%d: %v", id, err)
		}
		if got != id {
			t.Fatalf("expected %d, got %d", id, got)
		}
	}
}

func TestValidateToken_InvalidToken(t *testing.T) {
	ta := NewTokenAuth(testSecret)

	_, err := ta.ValidateToken("totally.invalid.token")
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
}

func TestValidateToken_EmptyToken(t *testing.T) {
	ta := NewTokenAuth(testSecret)

	_, err := ta.ValidateToken("")
	if err == nil {
		t.Fatal("expected error for empty token")
	}
}

func TestValidateToken_WrongSecret(t *testing.T) {
	ta1 := NewTokenAuth("secret-one")
	ta2 := NewTokenAuth("secret-two")

	token, _ := ta1.GenerateToken(1)
	_, err := ta2.ValidateToken(token)

	if err == nil {
		t.Fatal("expected error when validating with wrong secret")
	}
}

func TestValidateToken_ExpiredToken(t *testing.T) {
	ta := NewTokenAuth(testSecret)

	claims := jwt.MapClaims{
		"user_id": float64(1),
		"exp":     time.Now().Add(-1 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(testSecret))

	_, err := ta.ValidateToken(tokenString)
	if err == nil {
		t.Fatal("expected error for expired token")
	}
}

func TestValidateToken_MissingUserID(t *testing.T) {
	ta := NewTokenAuth(testSecret)

	claims := jwt.MapClaims{
		"exp": time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(testSecret))

	_, err := ta.ValidateToken(tokenString)
	if err == nil {
		t.Fatal("expected error for token without user_id")
	}
}

func TestValidateToken_WrongSigningMethod(t *testing.T) {
	claims := jwt.MapClaims{
		"user_id": float64(1),
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	tokenString, _ := token.SignedString(jwt.UnsafeAllowNoneSignatureType)

	ta := NewTokenAuth(testSecret)
	_, err := ta.ValidateToken(tokenString)
	if err == nil {
		t.Fatal("expected error for none signing method")
	}
}

func TestNewTokenAuth(t *testing.T) {
	ta := NewTokenAuth("my-secret")
	if ta == nil {
		t.Fatal("expected non-nil TokenAuth")
	}
	if string(ta.secret) != "my-secret" {
		t.Fatalf("expected secret my-secret, got %s", string(ta.secret))
	}
}
