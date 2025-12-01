package utils

import (
	"Avocycle/models"
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecretKey = []byte(os.Getenv("JWT_SECRET"))

type Claims struct {
	UserID       uint   `json:"user_id"`
    Email        string `json:"email"`
    Role         string `json:"role"`
    AuthProvider string `json:"auth_provider"`
    jwt.RegisteredClaims
}

func GenerateJWT(user *models.User) (string, error) {
	expirationTime := time.Now().Add(24*time.Hour)

	claims := &Claims{
		UserID: user.ID,
		Email: user.Email,
		Role: user.Role,
		AuthProvider: user.AuthProvider,
		RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(expirationTime),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            NotBefore: jwt.NewNumericDate(time.Now()),
            Issuer:    "Avocycle",
            Subject:   user.Email,
        },
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(jwtSecretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateJWT validates and parses JWT token
func ValidateJWT(tokenString string) (*Claims, error) {
    // Check if JWT secret is set
    if len(tokenString) == 0 {
        return nil, errors.New("JWT secret key is not set")
    }

    // Parse token
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        // Validate signing method
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, errors.New("invalid signing method")
        }
        return jwtSecretKey, nil
    })

    if err != nil {
        return nil, err
    }

    // Extract claims
    if claims, ok := token.Claims.(*Claims); ok && token.Valid {
        return claims, nil
    }

    return nil, errors.New("invalid token")
}

// RefreshJWT generates a new JWT token with extended expiration
func RefreshJWT(oldTokenString string) (string, error) {
    // Validate old token
    claims, err := ValidateJWT(oldTokenString)
    if err != nil {
        return "", err
    }

    // Check if token is expired
    if time.Until(claims.ExpiresAt.Time) < 0 {
        return "", errors.New("token has expired")
    }

    // Create new token with extended expiration
    newExpirationTime := time.Now().Add(24 * time.Hour)
    
    newClaims := &Claims{
        UserID:       claims.UserID,
        Email:        claims.Email,
        Role:         claims.Role,
        AuthProvider: claims.AuthProvider,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(newExpirationTime),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            NotBefore: jwt.NewNumericDate(time.Now()),
            Issuer:    "Avocycle",
            Subject:   claims.Email,
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, newClaims)
    return token.SignedString(jwtSecretKey)
}

// ==== Temp token khusus flow Google ====

// TempGoogleClaims dipakai untuk menyimpan data sementara user Google
type TempGoogleClaims struct {
    Provider   string `json:"provider"`
    ProviderID string `json:"provider_id"`
    Email      string `json:"email"`
    FullName   string `json:"full_name"`
    jwt.RegisteredClaims
}

// GenerateTempGoogleToken membuat JWT jangka pendek untuk carry data Google user ke FE
func GenerateTempGoogleToken(provider, providerID, email, fullName string) (string, error) {
    expirationTime := time.Now().Add(10 * time.Minute)

    claims := &TempGoogleClaims{
        Provider:   provider,
        ProviderID: providerID,
        Email:      email,
        FullName:   fullName,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(expirationTime),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            NotBefore: jwt.NewNumericDate(time.Now()),
            Issuer:    "Avocycle-Google-Temp",
            Subject:   email,
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(jwtSecretKey)
}

// ParseTempGoogleToken mem-parse tempToken dari FE
func ParseTempGoogleToken(tokenString string) (*TempGoogleClaims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &TempGoogleClaims{}, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, errors.New("invalid signing method")
        }
        return jwtSecretKey, nil
    })

    if err != nil {
        return nil, err
    }

    if claims, ok := token.Claims.(*TempGoogleClaims); ok && token.Valid {
        return claims, nil
    }

    return nil, errors.New("invalid temp token")
}
