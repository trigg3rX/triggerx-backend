package jwt

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/trigg3rX/triggerx-backend/pkg/rpc"
	"github.com/trigg3rX/triggerx-backend/pkg/rpc/auth/config"
)

// AuthInterceptor provides JWT authentication for gRPC requests
type AuthInterceptor struct {
	validator *Validator
	config    *config.AuthConfig
}

// NewAuthInterceptor creates a new JWT authentication interceptor
func NewAuthInterceptor(cfg *config.AuthConfig) *AuthInterceptor {
	return &AuthInterceptor{
		validator: NewValidator(cfg),
		config:    cfg,
	}
}

// UnaryServerInterceptor returns a gRPC unary server interceptor for JWT authentication
func (ai *AuthInterceptor) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Skip authentication for certain methods (health checks, etc.)
		if ai.shouldSkipAuth(info.FullMethod) {
			return handler(ctx, req)
		}

		// Extract and validate JWT token
		claims, err := ai.AuthenticateRequest(ctx)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "authentication failed: %v", err)
		}

		// Add claims to context
		ctx = ai.addClaimsToContext(ctx, claims)

		// Continue with the request
		return handler(ctx, req)
	}
}

// UnaryClientInterceptor returns a gRPC unary client interceptor for JWT authentication
func (ai *AuthInterceptor) UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		// Add JWT token to outgoing metadata if available
		ctx = ai.addTokenToContext(ctx)
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// AuthenticateRequest extracts and validates JWT token from gRPC metadata
func (ai *AuthInterceptor) AuthenticateRequest(ctx context.Context) (*Claims, error) {
	// Extract token from metadata
	token, err := ai.extractTokenFromMetadata(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to extract token: %w", err)
	}

	// Validate token
	claims, err := ai.validator.ValidateToken(token)
	if err != nil {
		return nil, fmt.Errorf("token validation failed: %w", err)
	}

	return claims, nil
}

// extractTokenFromMetadata extracts JWT token from gRPC metadata
func (ai *AuthInterceptor) extractTokenFromMetadata(ctx context.Context) (string, error) {
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

// addClaimsToContext adds JWT claims to the context
func (ai *AuthInterceptor) addClaimsToContext(ctx context.Context, claims *Claims) context.Context {
	// Add claims to context using a custom key
	return context.WithValue(ctx, rpc.JWTClaimsKey, claims)
}

// addTokenToContext adds JWT token to outgoing metadata
func (ai *AuthInterceptor) addTokenToContext(ctx context.Context) context.Context {
	// This would typically get the token from the current context or session
	// For now, we'll return the context as-is
	return ctx
}

// shouldSkipAuth determines if authentication should be skipped for a method
func (ai *AuthInterceptor) shouldSkipAuth(method string) bool {
	// Skip authentication for health checks and other public methods
	skipMethods := []string{
		"/grpc.health.v1.Health/Check",
		"/grpc.health.v1.Health/Watch",
		"/grpc.reflection.v1alpha.ServerReflection/ServerReflectionInfo",
		"/triggerx.backend.AuthService/Login",
		"/triggerx.backend.AuthService/Register",
		"/triggerx.backend.AuthService/RefreshToken",
	}

	for _, skipMethod := range skipMethods {
		if method == skipMethod {
			return true
		}
	}

	return false
}

// GetClaimsFromContext extracts JWT claims from context
func (ai *AuthInterceptor) GetClaimsFromContext(ctx context.Context) (*Claims, bool) {
	claims, ok := ctx.Value(rpc.JWTClaimsKey).(*Claims)
	return claims, ok
}

// RequirePermission creates an interceptor that requires specific permissions
func (ai *AuthInterceptor) RequirePermission(permissions ...string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Get claims from context
		claims, ok := ai.GetClaimsFromContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.Unauthenticated, "no authentication claims found")
		}

		// Check permissions
		if !claims.HasAllPermissions(permissions...) {
			return nil, status.Errorf(codes.PermissionDenied, "insufficient permissions: required %v", permissions)
		}

		return handler(ctx, req)
	}
}

// RequireRole creates an interceptor that requires a specific role
func (ai *AuthInterceptor) RequireRole(role string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Get claims from context
		claims, ok := ai.GetClaimsFromContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.Unauthenticated, "no authentication claims found")
		}

		// Check role
		if claims.Role != role {
			return nil, status.Errorf(codes.PermissionDenied, "insufficient role: required %s, got %s", role, claims.Role)
		}

		return handler(ctx, req)
	}
}

// OptionalAuth creates an interceptor that optionally authenticates requests
func (ai *AuthInterceptor) OptionalAuth() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Try to authenticate, but don't fail if no token is provided
		claims, err := ai.AuthenticateRequest(ctx)
		if err == nil {
			// Authentication successful, add claims to context
			ctx = ai.addClaimsToContext(ctx, claims)
		}
		// If authentication fails, continue without claims

		return handler(ctx, req)
	}
}

// GenerateToken generates a new JWT token for a user
func (ai *AuthInterceptor) GenerateToken(userID, email, address, role string, permissions []string) (string, error) {
	return ai.validator.GenerateAccessToken(userID, email, address, role, permissions)
}

// GenerateRefreshToken generates a refresh token for a user
func (ai *AuthInterceptor) GenerateRefreshToken(userID string) (string, error) {
	return ai.validator.GenerateRefreshToken(userID)
}

// ValidateToken validates a JWT token
func (ai *AuthInterceptor) ValidateToken(token string) (*Claims, error) {
	return ai.validator.ValidateToken(token)
}
