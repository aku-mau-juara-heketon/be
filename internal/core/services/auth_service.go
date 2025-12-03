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
