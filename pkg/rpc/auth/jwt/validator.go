package jwt

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/trigg3rX/triggerx-backend/pkg/rpc/auth/config"
)

// Validator handles JWT token validation and parsing
type Validator struct {
	config *config.AuthConfig
}

// NewValidator creates a new JWT validator
func NewValidator(cfg *config.AuthConfig) *Validator {
	return &Validator{
		config: cfg,
	}
}

// ValidateToken validates and parses a JWT token
func (v *Validator) ValidateToken(tokenString string) (*Claims, error) {
	// Remove "Bearer " prefix if present
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	// Parse the token
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(v.config.GetJWTSecretKey()), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	// Extract claims
	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Validate claims
	if err := v.validateClaims(claims); err != nil {
		return nil, err
	}

	return claims, nil
}

// ValidateTokenFromContext validates a token from gRPC metadata
func (v *Validator) ValidateTokenFromContext(ctx context.Context) (*Claims, error) {
	// Extract token from metadata
	token, err := v.extractTokenFromContext(ctx)
	if err != nil {
		return nil, err
	}

	return v.ValidateToken(token)
}

// GenerateToken generates a new JWT token with the given claims
func (v *Validator) GenerateToken(claims *Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(v.config.GetJWTSecretKey()))
}

// GenerateAccessToken generates an access token for a user
func (v *Validator) GenerateAccessToken(userID, email, address, role string, permissions []string) (string, error) {
	claims := NewAccessTokenClaims(userID, email, address, role, permissions, v.config.GetAccessTokenExpiry())
	return v.GenerateToken(claims)
}

// GenerateRefreshToken generates a refresh token for a user
func (v *Validator) GenerateRefreshToken(userID string) (string, error) {
	claims := NewRefreshTokenClaims(userID, v.config.GetRefreshTokenExpiry())
	return v.GenerateToken(claims)
}

// RefreshToken refreshes an access token using a refresh token
func (v *Validator) RefreshToken(refreshTokenString string, userID, email, address, role string, permissions []string) (string, error) {
	// Validate the refresh token
	claims, err := v.ValidateToken(refreshTokenString)
	if err != nil {
		return "", fmt.Errorf("invalid refresh token: %w", err)
	}

	// Check if it's actually a refresh token
	if !claims.IsRefreshToken() {
		return "", fmt.Errorf("token is not a refresh token")
	}

	// Check if the user ID matches
	if claims.UserID != userID {
		return "", fmt.Errorf("refresh token user ID mismatch")
	}

	// Generate new access token
	return v.GenerateAccessToken(userID, email, address, role, permissions)
}

// ValidateClaims validates the JWT claims
func (v *Validator) validateClaims(claims *Claims) error {
	// Check if token is expired
	if claims.IsExpired() {
		return fmt.Errorf("token is expired")
	}

	// Check if token is not yet valid
	if time.Now().Before(claims.NotBefore.Time) {
		return fmt.Errorf("token is not yet valid")
	}

	// Check required fields
	if claims.UserID == "" {
		return fmt.Errorf("missing user ID in token")
	}

	if claims.Subject == "" {
		return fmt.Errorf("missing subject in token")
	}

	// Validate issuer
	if claims.Issuer != "triggerx-backend" {
		return fmt.Errorf("invalid token issuer")
	}

	// Validate audience
	validAudience := false
	for _, audience := range claims.Audience {
		if audience == "triggerx-backend" {
			validAudience = true
			break
		}
	}
	if !validAudience {
		return fmt.Errorf("invalid token audience")
	}

	return nil
}

// ExtractTokenFromMetadata extracts token from gRPC metadata
func (v *Validator) extractTokenFromContext(ctx context.Context) (string, error) {
	// This would be implemented based on your gRPC metadata structure
	// For now, we'll use a placeholder implementation
	return "", fmt.Errorf("extractTokenFromContext not implemented")
}

// IsTokenExpired checks if a token is expired without full validation
func (v *Validator) IsTokenExpired(tokenString string) (bool, error) {
	// Parse token without validation to check expiration
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, &Claims{})
	if err != nil {
		return false, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return false, fmt.Errorf("invalid token claims")
	}

	return claims.IsExpired(), nil
}

// GetTokenExpirationTime gets the expiration time of a token
func (v *Validator) GetTokenExpirationTime(tokenString string) (time.Time, error) {
	// Parse token without validation to get expiration
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, &Claims{})
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return time.Time{}, fmt.Errorf("invalid token claims")
	}

	expTime, err := claims.GetExpirationTime()
	if err != nil {
		return time.Time{}, err
	}
	return expTime.Time, nil
}

// ValidateTokenPermissions validates if a token has the required permissions
func (v *Validator) ValidateTokenPermissions(tokenString string, requiredPermissions ...string) error {
	claims, err := v.ValidateToken(tokenString)
	if err != nil {
		return err
	}

	if !claims.HasAllPermissions(requiredPermissions...) {
		return fmt.Errorf("token does not have required permissions: %v", requiredPermissions)
	}

	return nil
}

// ValidateTokenRole validates if a token has the required role
func (v *Validator) ValidateTokenRole(tokenString string, requiredRole string) error {
	claims, err := v.ValidateToken(tokenString)
	if err != nil {
		return err
	}

	if claims.Role != requiredRole {
		return fmt.Errorf("token does not have required role: %s", requiredRole)
	}

	return nil
}

// GetTokenClaims gets the claims from a token without validation
func (v *Validator) GetTokenClaims(tokenString string) (*Claims, error) {
	// Parse token without validation
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, &Claims{})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}
