package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims represents the JWT claims structure
type Claims struct {
	UserID      string            `json:"user_id"`
	Email       string            `json:"email"`
	Address     string            `json:"address"`
	Role        string            `json:"role"`
	Permissions []string          `json:"permissions"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	jwt.RegisteredClaims
}

// NewClaims creates a new JWT claims instance
func NewClaims(userID, email, address, role string, permissions []string) *Claims {
	now := time.Now()

	return &Claims{
		UserID:      userID,
		Email:       email,
		Address:     address,
		Role:        role,
		Permissions: permissions,
		Metadata:    make(map[string]string),
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "triggerx-backend",
			Subject:   userID,
			Audience:  []string{"triggerx-backend"},
			ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        generateTokenID(),
		},
	}
}

// NewAccessTokenClaims creates claims for access tokens
func NewAccessTokenClaims(userID, email, address, role string, permissions []string, expiry time.Duration) *Claims {
	now := time.Now()

	return &Claims{
		UserID:      userID,
		Email:       email,
		Address:     address,
		Role:        role,
		Permissions: permissions,
		Metadata:    make(map[string]string),
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "triggerx-backend",
			Subject:   userID,
			Audience:  []string{"triggerx-backend"},
			ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        generateTokenID(),
		},
	}
}

// NewRefreshTokenClaims creates claims for refresh tokens
func NewRefreshTokenClaims(userID string, expiry time.Duration) *Claims {
	now := time.Now()

	return &Claims{
		UserID:      userID,
		Email:       "",
		Address:     "",
		Role:        "",
		Permissions: []string{},
		Metadata:    make(map[string]string),
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "triggerx-backend",
			Subject:   userID,
			Audience:  []string{"triggerx-backend"},
			ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        generateTokenID(),
		},
	}
}

// AddMetadata adds metadata to the claims
func (c *Claims) AddMetadata(key, value string) {
	if c.Metadata == nil {
		c.Metadata = make(map[string]string)
	}
	c.Metadata[key] = value
}

// GetMetadata retrieves metadata from the claims
func (c *Claims) GetMetadata(key string) string {
	if c.Metadata == nil {
		return ""
	}
	return c.Metadata[key]
}

// HasPermission checks if the user has a specific permission
func (c *Claims) HasPermission(permission string) bool {
	for _, p := range c.Permissions {
		if p == permission {
			return true
		}
	}
	return false
}

// HasAnyPermission checks if the user has any of the specified permissions
func (c *Claims) HasAnyPermission(permissions ...string) bool {
	for _, required := range permissions {
		if c.HasPermission(required) {
			return true
		}
	}
	return false
}

// HasAllPermissions checks if the user has all of the specified permissions
func (c *Claims) HasAllPermissions(permissions ...string) bool {
	for _, required := range permissions {
		if !c.HasPermission(required) {
			return false
		}
	}
	return true
}

// IsExpired checks if the token is expired
func (c *Claims) IsExpired() bool {
	return time.Now().After(c.ExpiresAt.Time)
}

// IsValid checks if the token is valid (not expired and has required fields)
func (c *Claims) IsValid() bool {
	return !c.IsExpired() && c.UserID != "" && c.Subject != ""
}

// GetExpirationTime returns the expiration time
func (c *Claims) GetExpirationTime() (*jwt.NumericDate, error) {
	return c.ExpiresAt, nil
}

// GetIssuedTime returns the issued time
func (c *Claims) GetIssuedTime() time.Time {
	return c.IssuedAt.Time
}

// GetTimeUntilExpiry returns the time until the token expires
func (c *Claims) GetTimeUntilExpiry() time.Duration {
	return time.Until(c.ExpiresAt.Time)
}

// IsRefreshToken checks if this is a refresh token
func (c *Claims) IsRefreshToken() bool {
	return c.Email == "" && c.Address == "" && c.Role == "" && len(c.Permissions) == 0
}

// IsAccessToken checks if this is an access token
func (c *Claims) IsAccessToken() bool {
	return !c.IsRefreshToken()
}

// generateTokenID generates a unique token ID
func generateTokenID() string {
	return "token_" + time.Now().Format("20060102150405") + "_" + randomString(8)
}

// randomString generates a random string of specified length
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}
