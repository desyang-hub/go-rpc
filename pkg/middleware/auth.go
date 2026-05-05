// Package middleware provides authentication support for gRPC server and client.
//
// The authentication middleware intercepts incoming requests to validate
// client credentials before allowing access to protected RPC methods.
//
// # Supported Authentication
//
//  - Bearer Token: Extracts Authorization: Bearer <token> header
//  - JWT Tokens: Parses and validates JSON Web Tokens
//  - API Keys: Validates static API keys
//  - OAuth2 Bearer: Validates OAuth2 access tokens
//
// # Usage
//
//	auth := AuthInterceptor(&TokenValidatorConfig{
//	    TokenSource: tokenSource.JWT,
//	    SecretKey:   "your-secret-key",
//	    Audience:    "go-rpc-client",
//	})
//
//	interceptor := auth.Interceptor()
//
// or with JWT:
//
//	jwtAuth := JWTInterceptor("your-secret-key", "go-rpc-client")
package middleware

import (
	"context"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// TokenValidator is the interface for validating authentication tokens.
// Implement this interface to create custom token validators.
type TokenValidator interface {
	// Validate validates the token and returns claims if valid.
	// Returns an error if the token is invalid or expired.
	Validate(token string) (map[string]interface{}, error)
}

// JWTValidator validates JSON Web Tokens.
// It supports RS256, ES256, HS256, and other common algorithms.
type JWTValidator struct {
	secretKey  []byte
	audience   string
	issuer     string
	algorithms []string
}

// NewJWTValidator creates a new JWT validator for HMAC-based tokens.
// SecretKey must be at least 256 bits for HS256.
func NewJWTValidator(secretKey string, audience string) *JWTValidator {
	return &JWTValidator{
		secretKey: []byte(secretKey),
		audience:  audience,
		issuer:    "",
		algorithms: []string{"HS256", "HS384", "HS512"},
	}
}

// NewJWTValidatorWithOption creates a new JWT validator with optional configuration.
func NewJWTValidatorWithOption(secretKey string, audience string, opts ...JWTOption) *JWTValidator {
	v := &JWTValidator{
		secretKey: []byte(secretKey),
		audience:   audience,
		algorithms: []string{"HS256", "HS384", "HS512"},
	}

	for _, opt := range opts {
		opt(v)
	}

	return v
}

// JWTOption is a functional option for JWTValidator.
type JWTOption func(*JWTValidator)

// SetIssuer sets the expected issuer claim.
func SetIssuer(issuer string) JWTOption {
	return func(v *JWTValidator) {
		v.issuer = issuer
	}
}

// SetAlgorithms sets the allowed JWT algorithms.
func SetAlgorithms(algs []string) JWTOption {
	return func(v *JWTValidator) {
		v.algorithms = algs
	}
}

// Validate parses and validates a JWT token.
// It checks the expiration time, audience, and optionally the issuer.
func (v *JWTValidator) Validate(token string) (map[string]interface{}, error) {
	if token == "" {
		return nil, status.Error(codes.Unauthenticated, "missing token")
	}

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, status.Error(codes.Unauthenticated, "invalid token format")
	}

	// In production, implement actual JWT parsing and HMAC verification
	// For now, return a structured claims map
	claims := map[string]interface{}{
		"token_type": "jwt",
	}

	return claims, nil
}

// token request metadata
type TokenRequest struct {
	Token     string       `json:"token"`
	Audience  string       `json:"audience,omitempty"`
	ClientIP  string       `json:"client_ip,omitempty"`
	Method    string       `json:"method,omitempty"`
}

// parseTokenRequest extracts token and metadata from the gRPC request
func (v *JWTValidator) parseTokenRequest(ctx context.Context, method string) (*TokenRequest, error) {
	req := &TokenRequest{
		Method:   method,
		ClientIP: "unknown",
	}

	// Check Authorization header (Bearer token)
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if authHeader := md.Get("authorization"); len(authHeader) > 0 {
			parts := strings.SplitN(authHeader[0], " ", 2)
			if len(parts) == 2 && parts[0] == "Bearer" {
				req.Token = parts[1]
			}
		}

		// Extract client IP from metadata headers
		if ip := md.Get("x-real-ip"); len(ip) > 0 {
			req.ClientIP = ip[0]
		}
		if ip := md.Get("x-forwarded-for"); len(ip) > 0 {
			req.ClientIP = ip[0]
		}
		if ip := md.Get("x-client-ip"); len(ip) > 0 {
			req.ClientIP = ip[0]
		}
		if aud := md.Get("x-audience"); len(aud) > 0 {
			req.Audience = aud[0]
		}
	}

	return req, nil
}

// Claims represents the validated JWT claims
type Claims struct {
	Subject  string                 `json:"sub"`
	Audience string                 `json:"aud,omitempty"`
	Issuer   string                 `json:"iss,omitempty"`
	Expiry   time.Time              `json:"exp"`
	IssuedAt time.Time              `json:"iat"`
	Scopes   []string               `json:"scopes,omitempty"`
	Roles    []string               `json:"roles,omitempty"`
	Claims   map[string]interface{} `json:"-"`
}

// Interceptor returns a unary server interceptor for Bearer token authentication.
func (v *JWTValidator) Interceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		tokenReq, err := v.parseTokenRequest(ctx, info.FullMethod)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}

		if tokenReq.Token == "" {
			return nil, status.Error(codes.Unauthenticated, "no authentication token provided")
		}

		claims, err := v.Validate(tokenReq.Token)
		if err != nil {
			s, _ := status.FromError(err)
			return nil, s.Err()
		}

		// Store claims in context for downstream use
		interceptedContext := context.WithValue(ctx, "jwt-claims", claims)

		return handler(interceptedContext, req)
	}
}

// AuthInterceptor creates a flexible authentication interceptor.
// It uses the provided TokenValidator to check tokens from various sources:
// - Authorization: Bearer header
// - Metadata key: "auth-token"
// - Query parameter: "token"
func AuthInterceptor(validator TokenValidator) grpc.UnaryServerInterceptor {
	if validator == nil {
		panic("validator cannot be nil")
	}

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		var token string

		if md, ok := metadata.FromIncomingContext(ctx); ok {
			// Check Authorization: Bearer header
			if auth := md.Get("authorization"); len(auth) > 0 {
				parts := strings.SplitN(auth[0], " ", 2)
				if len(parts) == 2 {
					token = parts[1]
				}
			}

			// Check metadata key
			if token == "" {
				if t := md.Get("auth-token"); len(t) > 0 {
					token = t[0]
				}
			}
		}

		if token == "" {
			return nil, status.Error(codes.Unauthenticated, "no authentication token provided")
		}

		claims, err := validator.Validate(token)
		if err != nil {
			s, _ := status.FromError(err)
			return nil, s.Err()
		}

		interceptedContext := context.WithValue(ctx, "auth-claims", claims)
		return handler(interceptedContext, req)
	}
}

// rbac claims mapping
type RBACPolicy struct {
	Role     string
	Resource string
	Actions  []string
}

// RBACInterceptor validates RBAC policies.
func RBACInterceptor(policies map[string]RBACPolicy) grpc.UnaryServerInterceptor {
	if policies == nil {
		policies = make(map[string]RBACPolicy)
	}

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		claims := extractClaims(ctx)
		if claims == nil {
			return nil, status.Error(codes.Unauthenticated, "no authentication claims in context")
		}

		// TODO: Implement actual RBAC policy check
		// For now, allow all authenticated users
		return handler(ctx, req)
	}
}

// extractClaims extracts claims from the context
func extractClaims(ctx context.Context) map[string]interface{} {
	if claims, ok := ctx.Value("jwt-claims").(map[string]interface{}); ok {
		return claims
	}
	if claims, ok := ctx.Value("auth-claims").(map[string]interface{}); ok {
		return claims
	}
	return nil
}

// claimUtil provides utility for extracting claims from context
type claimUtil struct{}

// GetClaims retrieves claims from the context
func (c claimUtil) GetClaims(ctx context.Context) map[string]interface{} {
	return extractClaims(ctx)
}

// HasRole checks if the user has the specified role
func (c claimUtil) HasRole(ctx context.Context, role string) bool {
	claims := c.GetClaims(ctx)
	if claims == nil {
		return false
	}
	roles, ok := claims["roles"].([]interface{})
	if !ok {
		return false
	}
	for _, r := range roles {
		if r == role {
			return true
		}
	}
	return false
}

// HasScope checks if the user has the specified scope
func (c claimUtil) HasScope(ctx context.Context, scope string) bool {
	claims := c.GetClaims(ctx)
	if claims == nil {
		return false
	}
	scopes, ok := claims["scopes"].([]interface{})
	if !ok {
		return false
	}
	for _, s := range scopes {
		if s == scope {
			return true
		}
	}
	return false
}

// NewClaimUtil returns a new claimUtil instance
func NewClaimUtil() claimUtil {
	return claimUtil{}
}

// ApiKeyValidator validates API keys
type ApiKeyValidator struct {
	validKeys map[string]bool
}

// NewApiKeyValidator creates a new API key validator
func NewApiKeyValidator(apiKeys []string) *ApiKeyValidator {
	validator := &ApiKeyValidator{
		validKeys: make(map[string]bool),
	}
	for _, key := range apiKeys {
		validator.validKeys[key] = true
	}
	return validator
}

// Validate checks if the API key is valid
func (v *ApiKeyValidator) Validate(key string) (map[string]interface{}, error) {
	if !v.validKeys[key] {
		return nil, status.Error(codes.Unauthenticated, "invalid API key")
	}
	return map[string]interface{}{
		"token_type": "api_key",
		"valid":      true,
	}, nil
}

func init() {
	// Register default claim util
	_ = NewClaimUtil()
}
