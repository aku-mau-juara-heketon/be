package ports

import (
	"context"
	"hackathon-be/internal/core/domain"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByID(ctx context.Context, id uint) (*domain.User, error)
}

type AuthService interface {
	Login(ctx context.Context, email, password string) (string, string, error) // AccessToken, RefreshToken, Error
	Register(ctx context.Context, email, password string) error
	RefreshToken(ctx context.Context, refreshToken string) (string, string, error)
}
