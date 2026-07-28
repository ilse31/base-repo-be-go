package jwt

import (
	"testing"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
)

func TestJWTManager_GenerateAndValidate(t *testing.T) {
	mgr := NewJWTManager("secret123", 1, 168)

	// Access Token
	token, err := mgr.GenerateAccessToken("user-1", "user@example.com")
	if err != nil {
		t.Fatalf("GenerateAccessToken failed: %v", err)
	}

	claims, err := mgr.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}
	if claims.UserID != "user-1" || claims.Email != "user@example.com" {
		t.Errorf("unexpected claims: %+v", claims)
	}

	// Refresh Token
	refToken, err := mgr.GenerateRefreshToken("user-1", "user@example.com")
	if err != nil {
		t.Fatalf("GenerateRefreshToken failed: %v", err)
	}

	claimsRef, err := mgr.ValidateToken(refToken)
	if err != nil {
		t.Fatalf("ValidateToken (refresh) failed: %v", err)
	}
	if claimsRef.UserID != "user-1" || claimsRef.Email != "user@example.com" {
		t.Errorf("unexpected claims: %+v", claimsRef)
	}

	// GenerateToken alias
	genToken, err := mgr.GenerateToken("user-1", "user@example.com")
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}
	if genToken == "" {
		t.Error("expected non-empty token")
	}
}

func TestJWTManager_InvalidToken(t *testing.T) {
	mgr := NewJWTManager("secret123", 1, 168)
	otherMgr := NewJWTManager("wrongsecret", 1, 168)

	token, _ := mgr.GenerateAccessToken("user-1", "user@example.com")

	// Validate with wrong secret
	_, err := otherMgr.ValidateToken(token)
	if err == nil {
		t.Error("expected error for wrong secret, got nil")
	}

	// Validate garbage string
	_, err = mgr.ValidateToken("not.a.valid.token")
	if err == nil {
		t.Error("expected error for malformed token, got nil")
	}
}

func TestJWTManager_ExpiredToken(t *testing.T) {
	mgr := NewJWTManager("secret123", -1, -1) // negative expiration = expired

	token, err := mgr.GenerateAccessToken("user-1", "user@example.com")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	_, err = mgr.ValidateToken(token)
	if err == nil {
		t.Error("expected error for expired token, got nil")
	}
}

func TestJWTManager_InvalidSigningMethod(t *testing.T) {
	claims := &Claims{
		UserID: "user-1",
		Email:  "test@example.com",
		RegisteredClaims: jwtlib.RegisteredClaims{
			ExpiresAt: jwtlib.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	// Sign with None or unsupported method
	token := jwtlib.NewWithClaims(jwtlib.SigningMethodNone, claims)
	tokenStr, _ := token.SignedString(jwtlib.UnsafeAllowNoneSignatureType)

	mgr := NewJWTManager("secret123", 1, 168)
	_, err := mgr.ValidateToken(tokenStr)
	if err == nil {
		t.Error("expected error for unsupported signing method, got nil")
	}
}
