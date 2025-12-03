package ports

import (
	"context"
)

type AuthService interface {
	Login(ctx context.Context, email, password string) (string, string, error) // AccessToken, RefreshToken, Error
	Register(ctx context.Context, email, password string) error
	RefreshToken(ctx context.Context, refreshToken string) (string, string, error)
	LoginWithOAuth(ctx context.Context, provider, code string) (string, string, error)
}
