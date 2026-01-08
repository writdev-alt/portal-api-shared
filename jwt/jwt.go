package jwt

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	jwtSecret          []byte
	accessTokenExpiry  time.Duration
	refreshTokenExpiry time.Duration
)

// Claims represents JWT claims
type Claims struct {
	UUID  string `json:"uuid"`
	Email string `json:"email"`
	Name  string `json:"name"`
	jwt.RegisteredClaims
}

// Init initializes JWT with secret from environment variables.
//
// Required:
// - JWT_SECRET
//
// Optional:
// - JWT_ACCESS_TOKEN_EXPIRY (Go duration string, e.g. "1h", "30m") OR JWT_ACCESS_TOKEN_EXPIRY_HOURS (int)
// - JWT_REFRESH_TOKEN_EXPIRY (Go duration string, e.g. "168h") OR JWT_REFRESH_TOKEN_EXPIRY_DAYS (int)
func Init() {
	secret := os.Getenv("JWT_SECRET")
	if strings.TrimSpace(secret) == "" {
		panic("JWT_SECRET is not configured")
	}
	jwtSecret = []byte(secret)

	accessTokenExpiry = parseDurationOrHours("JWT_ACCESS_TOKEN_EXPIRY", "JWT_ACCESS_TOKEN_EXPIRY_HOURS", time.Hour)
	refreshTokenExpiry = parseDurationOrDays("JWT_REFRESH_TOKEN_EXPIRY", "JWT_REFRESH_TOKEN_EXPIRY_DAYS", 7*24*time.Hour)
}

// GetSecret returns the JWT secret for debugging/verification purposes
// WARNING: Only use this for debugging. Never expose in production responses.
func GetSecret() string {
	if len(jwtSecret) == 0 {
		Init()
	}
	return string(jwtSecret)
}

// GetAccessTokenExpiry returns the access token expiry duration
func GetAccessTokenExpiry() time.Duration {
	if accessTokenExpiry == 0 {
		Init()
	}
	return accessTokenExpiry
}

// GenerateToken generates a JWT token for admin
func GenerateToken(id uuid.UUID, email, name string) (string, error) {
	return GenerateTokenWithExpiry(id, email, name, accessTokenExpiry) // Default: 1 hour
}

// GenerateTokenWithExpiry generates a JWT token with custom expiration time
func GenerateTokenWithExpiry(id uuid.UUID, email, name string, expiry time.Duration) (string, error) {
	if len(jwtSecret) == 0 {
		Init()
	}

	claims := Claims{
		UUID:  id.String(),
		Email: email,
		Name:  name,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now().Add(-30 * time.Second)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// GenerateRefreshToken generates a refresh token with longer expiration
func GenerateRefreshToken(id uuid.UUID, email, name string) (string, error) {
	if refreshTokenExpiry == 0 {
		Init()
	}
	return GenerateTokenWithExpiry(id, email, name, refreshTokenExpiry)
}

// ValidateToken validates a JWT token and returns the claims
func ValidateToken(tokenString string) (*Claims, error) {
	if len(jwtSecret) == 0 {
		Init()
	}

	token, err := jwt.ParseWithClaims(
		tokenString,
		&Claims{},
		func(token *jwt.Token) (interface{}, error) {
			// Verify the signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			// Return the secret key for verification
			return jwtSecret, nil
		},
		jwt.WithValidMethods([]string{"HS256"}),
		// jwt.WithSkipClaimsValidation(true),
		jwt.WithLeeway(30*time.Second),
	)

	if err != nil {
		return nil, err
	}

	// Check if token is valid
	if !token.Valid {
		return nil, errors.New("token is not valid")
	}

	if claims, ok := token.Claims.(*Claims); ok {
		return claims, nil
	}

	return nil, errors.New("invalid token claims")
}

// ComparePassword compares a hashed password with a plain password
func ComparePassword(hashedPassword, plainPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
	return err == nil
}

func parseDurationOrHours(durationKey, hoursKey string, defaultValue time.Duration) time.Duration {
	if v := strings.TrimSpace(os.Getenv(durationKey)); v != "" {
		if d, err := time.ParseDuration(v); err == nil && d > 0 {
			return d
		}
	}
	if v := strings.TrimSpace(os.Getenv(hoursKey)); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return time.Duration(n) * time.Hour
		}
	}
	return defaultValue
}

func parseDurationOrDays(durationKey, daysKey string, defaultValue time.Duration) time.Duration {
	if v := strings.TrimSpace(os.Getenv(durationKey)); v != "" {
		if d, err := time.ParseDuration(v); err == nil && d > 0 {
			return d
		}
	}
	if v := strings.TrimSpace(os.Getenv(daysKey)); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return time.Duration(n) * 24 * time.Hour
		}
	}
	return defaultValue
}
