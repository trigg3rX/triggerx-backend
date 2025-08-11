package middleware

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/trigg3rX/triggerx-backend/pkg/rpc"
	"github.com/trigg3rX/triggerx-backend/pkg/rpc/auth/config"
	"github.com/trigg3rX/triggerx-backend/pkg/rpc/auth/jwt"
)

// AuthMiddleware provides authentication middleware for gRPC services
type AuthMiddleware struct {
	jwtInterceptor *jwt.AuthInterceptor
	config         *config.AuthConfig
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(cfg *config.AuthConfig) *AuthMiddleware {
	return &AuthMiddleware{
		jwtInterceptor: jwt.NewAuthInterceptor(cfg),
		config:         cfg,
	}
}

// AuthenticateRequest authenticates a gRPC request
func (am *AuthMiddleware) AuthenticateRequest(ctx context.Context, method string) (*jwt.Claims, error) {
	// Skip authentication for certain methods
	if am.shouldSkipAuth(method) {
		return nil, nil
	}

	// Extract and validate JWT token
	claims, err := am.jwtInterceptor.AuthenticateRequest(ctx)
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	return claims, nil
}

// RequireAuth creates an interceptor that requires authentication
func (am *AuthMiddleware) RequireAuth() grpc.UnaryServerInterceptor {
	return am.jwtInterceptor.UnaryServerInterceptor()
}

// RequirePermission creates an interceptor that requires specific permissions
func (am *AuthMiddleware) RequirePermission(permissions ...string) grpc.UnaryServerInterceptor {
	return am.jwtInterceptor.RequirePermission(permissions...)
}

// RequireRole creates an interceptor that requires a specific role
func (am *AuthMiddleware) RequireRole(role string) grpc.UnaryServerInterceptor {
	return am.jwtInterceptor.RequireRole(role)
}

// OptionalAuth creates an interceptor that optionally authenticates requests
func (am *AuthMiddleware) OptionalAuth() grpc.UnaryServerInterceptor {
	return am.jwtInterceptor.OptionalAuth()
}

// GetClaimsFromContext extracts JWT claims from context
func (am *AuthMiddleware) GetClaimsFromContext(ctx context.Context) (*jwt.Claims, bool) {
	return am.jwtInterceptor.GetClaimsFromContext(ctx)
}

// GenerateToken generates a new JWT token for a user
func (am *AuthMiddleware) GenerateToken(userID, email, address, role string, permissions []string) (string, error) {
	return am.jwtInterceptor.GenerateToken(userID, email, address, role, permissions)
}

// GenerateRefreshToken generates a refresh token for a user
func (am *AuthMiddleware) GenerateRefreshToken(userID string) (string, error) {
	return am.jwtInterceptor.GenerateRefreshToken(userID)
}

// ValidateToken validates a JWT token
func (am *AuthMiddleware) ValidateToken(token string) (*jwt.Claims, error) {
	return am.jwtInterceptor.ValidateToken(token)
}

// AddTokenToMetadata adds a JWT token to gRPC metadata
func (am *AuthMiddleware) AddTokenToMetadata(ctx context.Context, token string) context.Context {
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		md = metadata.New(nil)
	}

	md.Set("authorization", "Bearer "+token)
	return metadata.NewOutgoingContext(ctx, md)
}

// ExtractTokenFromMetadata extracts JWT token from gRPC metadata
func (am *AuthMiddleware) ExtractTokenFromMetadata(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", fmt.Errorf("no metadata found")
	}

	// Try to get token from Authorization header
	authHeaders := md.Get("authorization")
	if len(authHeaders) > 0 {
		token := authHeaders[0]
		if token != "" {
			return token, nil
		}
	}

	// Try to get token from x-auth-token header
	authTokens := md.Get("x-auth-token")
	if len(authTokens) > 0 {
		token := authTokens[0]
		if token != "" {
			return token, nil
		}
	}

	// Try to get token from x-jwt-token header
	jwtTokens := md.Get("x-jwt-token")
	if len(jwtTokens) > 0 {
		token := jwtTokens[0]
		if token != "" {
			return token, nil
		}
	}

	return "", fmt.Errorf("no authentication token found in metadata")
}

// shouldSkipAuth determines if authentication should be skipped for a method
func (am *AuthMiddleware) shouldSkipAuth(method string) bool {
	// Skip authentication for health checks and other public methods
	skipMethods := []string{
		"/grpc.health.v1.Health/Check",
		"/grpc.health.v1.Health/Watch",
		"/grpc.reflection.v1alpha.ServerReflection/ServerReflectionInfo",
		"/triggerx.backend.AuthService/Login",
		"/triggerx.backend.AuthService/Register",
		"/triggerx.backend.AuthService/RefreshToken",
		"/triggerx.backend.AuthService/ValidateToken",
	}

	for _, skipMethod := range skipMethods {
		if method == skipMethod {
			return true
		}
	}

	return false
}

// CreateAuthContext creates a context with authentication information
func (am *AuthMiddleware) CreateAuthContext(ctx context.Context, claims *jwt.Claims) context.Context {
	return context.WithValue(ctx, rpc.JWTClaimsKey, claims)
}

// IsAuthenticated checks if the context contains valid authentication claims
func (am *AuthMiddleware) IsAuthenticated(ctx context.Context) bool {
	claims, ok := am.GetClaimsFromContext(ctx)
	return ok && claims != nil && claims.IsValid()
}

// GetUserID extracts user ID from context
func (am *AuthMiddleware) GetUserID(ctx context.Context) (string, error) {
	claims, ok := am.GetClaimsFromContext(ctx)
	if !ok {
		return "", status.Errorf(codes.Unauthenticated, "no authentication claims found")
	}
	return claims.UserID, nil
}

// GetUserRole extracts user role from context
func (am *AuthMiddleware) GetUserRole(ctx context.Context) (string, error) {
	claims, ok := am.GetClaimsFromContext(ctx)
	if !ok {
		return "", status.Errorf(codes.Unauthenticated, "no authentication claims found")
	}
	return claims.Role, nil
}

// GetUserPermissions extracts user permissions from context
func (am *AuthMiddleware) GetUserPermissions(ctx context.Context) ([]string, error) {
	claims, ok := am.GetClaimsFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "no authentication claims found")
	}
	return claims.Permissions, nil
}

// HasPermission checks if the user has a specific permission
func (am *AuthMiddleware) HasPermission(ctx context.Context, permission string) bool {
	claims, ok := am.GetClaimsFromContext(ctx)
	if !ok {
		return false
	}
	return claims.HasPermission(permission)
}

// HasAnyPermission checks if the user has any of the specified permissions
func (am *AuthMiddleware) HasAnyPermission(ctx context.Context, permissions ...string) bool {
	claims, ok := am.GetClaimsFromContext(ctx)
	if !ok {
		return false
	}
	return claims.HasAnyPermission(permissions...)
}

// HasAllPermissions checks if the user has all of the specified permissions
func (am *AuthMiddleware) HasAllPermissions(ctx context.Context, permissions ...string) bool {
	claims, ok := am.GetClaimsFromContext(ctx)
	if !ok {
		return false
	}
	return claims.HasAllPermissions(permissions...)
}

// HasRole checks if the user has a specific role
func (am *AuthMiddleware) HasRole(ctx context.Context, role string) bool {
	claims, ok := am.GetClaimsFromContext(ctx)
	if !ok {
		return false
	}
	return claims.Role == role
}
