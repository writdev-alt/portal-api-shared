package jwt

import (
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
)

func setupJWTTest() {
	// Set up test environment variables
	os.Setenv("JWT_SECRET", "test-secret-key-for-unit-testing")
	os.Setenv("JWT_ACCESS_TOKEN_EXPIRY_HOURS", "1")
	os.Setenv("JWT_REFRESH_TOKEN_EXPIRY_DAYS", "7")

	// Reset global variables
	jwtSecret = nil
	accessTokenExpiry = 0
	refreshTokenExpiry = 0

	Init()
}

func teardownJWTTest() {
	os.Unsetenv("JWT_SECRET")
	os.Unsetenv("JWT_ACCESS_TOKEN_EXPIRY_HOURS")
	os.Unsetenv("JWT_REFRESH_TOKEN_EXPIRY_DAYS")
	os.Unsetenv("JWT_ACCESS_TOKEN_EXPIRY")
	os.Unsetenv("JWT_REFRESH_TOKEN_EXPIRY")
}

func TestInit(t *testing.T) {
	teardownJWTTest()
	defer teardownJWTTest()

	os.Setenv("JWT_SECRET", "test-secret")
	os.Setenv("JWT_ACCESS_TOKEN_EXPIRY_HOURS", "2")

	jwtSecret = nil
	accessTokenExpiry = 0

	// Should not panic
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Init panicked: %v", r)
			}
		}()
		Init()
	}()

	if len(jwtSecret) == 0 {
		t.Error("Init did not set jwtSecret")
	}

	if accessTokenExpiry == 0 {
		t.Error("Init did not set accessTokenExpiry")
	}
}

func TestInit_PanicsOnMissingSecret(t *testing.T) {
	teardownJWTTest()
	defer teardownJWTTest()

	os.Unsetenv("JWT_SECRET")
	jwtSecret = nil

	defer func() {
		if r := recover(); r == nil {
			t.Error("Init should panic when JWT_SECRET is not set")
		}
	}()

	Init()
}

func TestGenerateToken(t *testing.T) {
	setupJWTTest()
	defer teardownJWTTest()

	id := uuid.New()
	email := "test@example.com"
	name := "Test User"

	token, err := GenerateToken(id, email, name)
	if err != nil {
		t.Errorf("GenerateToken returned error: %v", err)
	}

	if token == "" {
		t.Error("GenerateToken returned empty token")
	}
}

func TestGenerateTokenWithExpiry(t *testing.T) {
	setupJWTTest()
	defer teardownJWTTest()

	id := uuid.New()
	email := "test@example.com"
	name := "Test User"
	expiry := time.Hour * 2

	token, err := GenerateTokenWithExpiry(id, email, name, expiry)
	if err != nil {
		t.Errorf("GenerateTokenWithExpiry returned error: %v", err)
	}

	if token == "" {
		t.Error("GenerateTokenWithExpiry returned empty token")
	}

	// Validate the token
	claims, err := ValidateToken(token)
	if err != nil {
		t.Errorf("ValidateToken returned error: %v", err)
	}

	if claims.UUID != id.String() {
		t.Errorf("Claims UUID = %s, expected %s", claims.UUID, id.String())
	}

	if claims.Email != email {
		t.Errorf("Claims Email = %s, expected %s", claims.Email, email)
	}

	if claims.Name != name {
		t.Errorf("Claims Name = %s, expected %s", claims.Name, name)
	}
}

func TestGenerateRefreshToken(t *testing.T) {
	setupJWTTest()
	defer teardownJWTTest()

	id := uuid.New()
	email := "test@example.com"
	name := "Test User"

	token, err := GenerateRefreshToken(id, email, name)
	if err != nil {
		t.Errorf("GenerateRefreshToken returned error: %v", err)
	}

	if token == "" {
		t.Error("GenerateRefreshToken returned empty token")
	}

	// Validate the token
	claims, err := ValidateToken(token)
	if err != nil {
		t.Errorf("ValidateToken returned error: %v", err)
	}

	if claims.UUID != id.String() {
		t.Errorf("Claims UUID = %s, expected %s", claims.UUID, id.String())
	}
}

func TestValidateToken(t *testing.T) {
	setupJWTTest()
	defer teardownJWTTest()

	id := uuid.New()
	email := "test@example.com"
	name := "Test User"

	token, err := GenerateToken(id, email, name)
	if err != nil {
		t.Fatalf("GenerateToken returned error: %v", err)
	}

	claims, err := ValidateToken(token)
	if err != nil {
		t.Errorf("ValidateToken returned error: %v", err)
	}

	if claims == nil {
		t.Error("ValidateToken returned nil claims")
		return
	}

	if claims.UUID != id.String() {
		t.Errorf("Claims UUID = %s, expected %s", claims.UUID, id.String())
	}

	if claims.Email != email {
		t.Errorf("Claims Email = %s, expected %s", claims.Email, email)
	}

	if claims.Name != name {
		t.Errorf("Claims Name = %s, expected %s", claims.Name, name)
	}
}

func TestValidateToken_InvalidToken(t *testing.T) {
	setupJWTTest()
	defer teardownJWTTest()

	invalidToken := "invalid.token.string"

	_, err := ValidateToken(invalidToken)
	if err == nil {
		t.Error("ValidateToken should return error for invalid token")
	}
}

func TestValidateToken_ExpiredToken(t *testing.T) {
	setupJWTTest()
	defer teardownJWTTest()

	id := uuid.New()
	email := "test@example.com"
	name := "Test User"

	// Generate token with very short expiry
	token, err := GenerateTokenWithExpiry(id, email, name, -time.Hour)
	if err != nil {
		t.Fatalf("GenerateTokenWithExpiry returned error: %v", err)
	}

	// Try to validate expired token
	_, err = ValidateToken(token)
	if err == nil {
		t.Error("ValidateToken should return error for expired token")
	}
}

func TestComparePassword(t *testing.T) {
	plainPassword := "test-password-123"
	hashedPassword := "$2a$04$testhashhere" // This is a mock - actual bcrypt hash would be longer

	// Test with invalid hash (this will fail but should not panic)
	result := ComparePassword(hashedPassword, plainPassword)
	// We expect false since the hash is invalid, but function should not panic
	_ = result

	// Test with empty strings
	result = ComparePassword("", "")
	if result {
		t.Error("ComparePassword should return false for empty strings")
	}
}

func TestGetSecret(t *testing.T) {
	setupJWTTest()
	defer teardownJWTTest()

	secret := GetSecret()
	if secret == "" {
		t.Error("GetSecret returned empty string")
	}

	if secret != "test-secret-key-for-unit-testing" {
		t.Errorf("GetSecret = %s, expected 'test-secret-key-for-unit-testing'", secret)
	}
}

func TestGetAccessTokenExpiry(t *testing.T) {
	setupJWTTest()
	defer teardownJWTTest()

	expiry := GetAccessTokenExpiry()
	if expiry == 0 {
		t.Error("GetAccessTokenExpiry returned zero duration")
	}

	expected := time.Hour
	if expiry != expected {
		t.Errorf("GetAccessTokenExpiry = %v, expected %v", expiry, expected)
	}
}

func BenchmarkGenerateToken(b *testing.B) {
	setupJWTTest()
	defer teardownJWTTest()

	id := uuid.New()
	email := "test@example.com"
	name := "Test User"

	for i := 0; i < b.N; i++ {
		GenerateToken(id, email, name)
	}
}

func BenchmarkValidateToken(b *testing.B) {
	setupJWTTest()
	defer teardownJWTTest()

	id := uuid.New()
	email := "test@example.com"
	name := "Test User"

	token, _ := GenerateToken(id, email, name)

	for i := 0; i < b.N; i++ {
		ValidateToken(token)
	}
}
