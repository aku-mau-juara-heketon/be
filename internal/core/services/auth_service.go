package services

import (
	"context"
	"errors"
	"hackathon-be/internal/core/domain"
	"hackathon-be/internal/core/ports"
	"hackathon-be/pkg/auth"
	"hackathon-be/pkg/config"
	"hackathon-be/pkg/logger"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	repo   ports.UserRepository
	config *config.Config
}

func NewAuthService(repo ports.UserRepository, cfg *config.Config) *AuthService {
	return &AuthService{
		repo:   repo,
		config: cfg,
	}
}

func (s *AuthService) Register(ctx context.Context, email, password string) error {
	existingUser, _ := s.repo.GetByEmail(ctx, email)
	if existingUser != nil {
		return errors.New("user already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		logger.Log.Error("Failed to hash password", "error", err)
		return err
	}

	user := &domain.User{
		Email:        email,
		PasswordHash: string(hashedPassword),
	}

	return s.repo.Create(ctx, user)
}

func (s *AuthService) Login(ctx context.Context, email, password string) (string, string, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return "", "", errors.New("invalid credentials")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return "", "", errors.New("invalid credentials")
	}

	return auth.GenerateTokens(user.ID, s.config.JWTSecret)
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (string, string, error) {
	claims, err := auth.ValidateToken(refreshToken, s.config.JWTSecret)
	if err != nil {
		return "", "", errors.New("invalid refresh token")
	}

	// Ideally, we should check if the user still exists or if the token has been revoked
	// For now, we just generate a new pair
	return auth.GenerateTokens(claims.UserID, s.config.JWTSecret)
}

func (s *AuthService) LoginWithOAuth(ctx context.Context, provider, code string) (string, string, error) {
	// 1. Exchange code for token & fetch user info
	var email, name, avatar, providerID string
	var err error

	switch provider {
	case "google":
		email, name, avatar, providerID, err = s.fetchGoogleUser(ctx, code)
	case "github":
		email, name, avatar, providerID, err = s.fetchGithubUser(ctx, code)
	default:
		return "", "", errors.New("unsupported provider")
	}

	if err != nil {
		return "", "", err
	}

	// 2. Check if user exists
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		// Create new user
		user = &domain.User{
			Email:      email,
			Name:       name,
			AvatarURL:  avatar,
			Provider:   provider,
			ProviderID: providerID,
		}
		if err := s.repo.Create(ctx, user); err != nil {
			return "", "", err
		}
	} else {
		// Update existing user info
		user.Name = name
		user.AvatarURL = avatar
		user.Provider = provider
		user.ProviderID = providerID
		// In a real app, we'd have an Update method in repo
		// For hackathon, we assume Create handles upsert or we ignore update for now to keep it simple
		// Or we can just proceed with existing user
	}

	// 3. Generate Tokens
	return auth.GenerateTokens(user.ID, s.config.JWTSecret)
}

// Helper methods (Mock implementation for now or real implementation if we want)
// Real implementation requires importing oauth2 packages
func (s *AuthService) fetchGoogleUser(ctx context.Context, code string) (string, string, string, string, error) {
	// TODO: Implement actual Google OAuth exchange
	// For now, returning dummy data to simulate success if code is "valid_google_code"
	if code == "valid_google_code" {
		return "google_user@example.com", "Google User", "https://avatar.google.com/123", "google_123", nil
	}
	return "", "", "", "", errors.New("failed to fetch google user")
}

func (s *AuthService) fetchGithubUser(ctx context.Context, code string) (string, string, string, string, error) {
	// TODO: Implement actual Github OAuth exchange
	if code == "valid_github_code" {
		return "github_user@example.com", "Github User", "https://avatar.github.com/123", "github_123", nil
	}
	return "", "", "", "", errors.New("failed to fetch github user")
}
