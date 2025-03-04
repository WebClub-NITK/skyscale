package auth

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// AuthManager handles authentication and authorization
type AuthManager struct {
	logger  *logrus.Logger
	apiKeys map[string]APIKey
	mu      sync.RWMutex
}

// APIKey represents an API key
type APIKey struct {
	Key       string
	UserID    string
	CreatedAt time.Time
	ExpiresAt time.Time
	Roles     []string
}

// NewAuthManager creates a new authentication manager
func NewAuthManager(logger *logrus.Logger) (*AuthManager, error) {
	return &AuthManager{
		logger:  logger,
		apiKeys: make(map[string]APIKey),
	}, nil
}

// GenerateAPIKey generates a new API key
func (a *AuthManager) GenerateAPIKey(userID string, roles []string, expiresIn time.Duration) (string, error) {
	// Generate random bytes
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	// Encode as base64
	key := base64.StdEncoding.EncodeToString(b)

	// Create API key
	apiKey := APIKey{
		Key:       key,
		UserID:    userID,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(expiresIn),
		Roles:     roles,
	}

	// Store API key
	a.mu.Lock()
	a.apiKeys[key] = apiKey
	a.mu.Unlock()

	return key, nil
}

// ValidateAPIKey validates an API key
func (a *AuthManager) ValidateAPIKey(key string) (APIKey, error) {
	a.mu.RLock()
	apiKey, exists := a.apiKeys[key]
	a.mu.RUnlock()

	if !exists {
		return APIKey{}, errors.New("invalid API key")
	}

	if time.Now().After(apiKey.ExpiresAt) {
		return APIKey{}, errors.New("API key expired")
	}

	return apiKey, nil
}

// RevokeAPIKey revokes an API key
func (a *AuthManager) RevokeAPIKey(key string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if _, exists := a.apiKeys[key]; !exists {
		return errors.New("API key not found")
	}

	delete(a.apiKeys, key)
	return nil
}

// HasRole checks if an API key has a specific role
func (a *AuthManager) HasRole(key string, role string) (bool, error) {
	apiKey, err := a.ValidateAPIKey(key)
	if err != nil {
		return false, err
	}

	for _, r := range apiKey.Roles {
		if r == role {
			return true, nil
		}
	}

	return false, nil
}

// Middleware creates a middleware for authentication
func (a *AuthManager) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get API key from header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Check if it's a Bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid authorization header", http.StatusUnauthorized)
			return
		}

		// Validate API key
		apiKey := parts[1]
		_, err := a.ValidateAPIKey(apiKey)
		if err != nil {
			http.Error(w, fmt.Sprintf("Unauthorized: %v", err), http.StatusUnauthorized)
			return
		}

		// Call next handler
		next.ServeHTTP(w, r)
	})
}

// RoleMiddleware creates a middleware for role-based authorization
func (a *AuthManager) RoleMiddleware(role string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get API key from header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Check if it's a Bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid authorization header", http.StatusUnauthorized)
			return
		}

		// Validate API key
		apiKey := parts[1]
		hasRole, err := a.HasRole(apiKey, role)
		if err != nil {
			http.Error(w, fmt.Sprintf("Unauthorized: %v", err), http.StatusUnauthorized)
			return
		}

		if !hasRole {
			http.Error(w, "Forbidden: insufficient permissions", http.StatusForbidden)
			return
		}

		// Call next handler
		next.ServeHTTP(w, r)
	})
}
