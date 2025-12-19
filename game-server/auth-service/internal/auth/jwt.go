package auth

import (
	"errors"
	"os"
	"time"

	"github.com/darkphotonKN/cosmic-void-server/auth-service/internal/models"
	commonconstants "github.com/darkphotonKN/cosmic-void-server/common/constants"
	"github.com/golang-jwt/jwt/v5"
)

/**
* Generates and signs a JWT token with claims of either the "access" or "refresh" types.
**/
func GenerateJWT(user models.Member, tokenType commonconstants.TokenType, expiration time.Duration) (string, error) {
	JWTSecret := []byte(os.Getenv("JWT_SECRET"))

	// Define the custom claims for the token
	claims := jwt.MapClaims{
		"sub":       user.ID.String(),
		"exp":       time.Now().Add(expiration).Unix(),
		"iat":       time.Now().Unix(),
		"tokenType": tokenType,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(JWTSecret)
}

/**
* RefreshToken validates a refresh token and generates a new access token if
* valid.
**/
func RefreshToken(refreshToken string, user models.Member) (string, int, error) {
	// Parse the refresh token
	token, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil || !token.Valid {
		return "", 0, errors.New("invalid refresh token")
	}

	// Check if the token type is "refresh"
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || claims["tokenType"] != string(commonconstants.Refresh) {
		return "", 0, errors.New("invalid token type")
	}

	// Generate a new access token with a 15-minute expiration
	newAccessToken, err := GenerateJWT(user, commonconstants.Access, 15*time.Minute)
	if err != nil {
		return "", 0, errors.New("could not generate new access token")
	}

	// Return the new access token and expiration time (in seconds)
	return newAccessToken, int(15 * 60), nil
}

// ValidateJWT validates a JWT token and returns the claims
func ValidateJWT(tokenString string) (jwt.RegisteredClaims, error) {
	JWTSecret := []byte(os.Getenv("JWT_SECRET"))

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return JWTSecret, nil
	})

	if err != nil {
		return jwt.RegisteredClaims{}, err
	}

	if !token.Valid {
		return jwt.RegisteredClaims{}, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return jwt.RegisteredClaims{}, errors.New("invalid claims")
	}

	// Extract sub (user ID) from MapClaims
	sub, _ := claims["sub"].(string)

	return jwt.RegisteredClaims{
		ID: sub,
	}, nil
}
