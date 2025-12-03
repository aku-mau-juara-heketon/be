package unit_test

import (
	"context"
	"errors"
	"hackathon-be/internal/core/domain"
	"hackathon-be/internal/core/services"
	"hackathon-be/pkg/config"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserRepository is a mock implementation of ports.UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id uint) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func TestRegister_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	cfg := &config.Config{JWTSecret: "secret"}
	svc := services.NewAuthService(mockRepo, cfg)

	ctx := context.Background()
	email := "test@example.com"
	password := "password123"

	// Expect GetByEmail to return nil (user does not exist)
	mockRepo.On("GetByEmail", ctx, email).Return(nil, nil)
	// Expect Create to be called
	mockRepo.On("Create", ctx, mock.AnythingOfType("*domain.User")).Return(nil)

	err := svc.Register(ctx, email, password)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestRegister_UserAlreadyExists(t *testing.T) {
	mockRepo := new(MockUserRepository)
	cfg := &config.Config{JWTSecret: "secret"}
	svc := services.NewAuthService(mockRepo, cfg)

	ctx := context.Background()
	email := "existing@example.com"
	password := "password123"

	existingUser := &domain.User{Email: email}

	// Expect GetByEmail to return a user
	mockRepo.On("GetByEmail", ctx, email).Return(existingUser, nil)

	err := svc.Register(ctx, email, password)

	assert.Error(t, err)
	assert.Equal(t, "user already exists", err.Error())
	mockRepo.AssertNotCalled(t, "Create")
}

func TestLoginWithOAuth_Success_NewUser(t *testing.T) {
	mockRepo := new(MockUserRepository)
	cfg := &config.Config{JWTSecret: "secret"}
	svc := services.NewAuthService(mockRepo, cfg)

	ctx := context.Background()
	provider := "google"
	code := "valid_google_code"
	email := "google_user@example.com"

	// Expect GetByEmail to return error (user does not exist)
	mockRepo.On("GetByEmail", ctx, email).Return(nil, errors.New("user not found"))
	// Expect Create to be called
	mockRepo.On("Create", ctx, mock.MatchedBy(func(u *domain.User) bool {
		return u.Email == email && u.Provider == provider && u.ProviderID == "google_123"
	})).Return(nil)

	accessToken, refreshToken, err := svc.LoginWithOAuth(ctx, provider, code)

	assert.NoError(t, err)
	assert.NotEmpty(t, accessToken)
	assert.NotEmpty(t, refreshToken)
	mockRepo.AssertExpectations(t)
}

func TestLoginWithOAuth_Success_ExistingUser(t *testing.T) {
	mockRepo := new(MockUserRepository)
	cfg := &config.Config{JWTSecret: "secret"}
	svc := services.NewAuthService(mockRepo, cfg)

	ctx := context.Background()
	provider := "github"
	code := "valid_github_code"
	email := "github_user@example.com"

	existingUser := &domain.User{
		ID:    1,
		Email: email,
	}

	// Expect GetByEmail to return existing user
	mockRepo.On("GetByEmail", ctx, email).Return(existingUser, nil)
	// Create should NOT be called
	// In our current implementation, we don't call Update on repo, we just modify the struct and generate token
	// If we added Update to repo, we would mock it here

	accessToken, refreshToken, err := svc.LoginWithOAuth(ctx, provider, code)

	assert.NoError(t, err)
	assert.NotEmpty(t, accessToken)
	assert.NotEmpty(t, refreshToken)
	mockRepo.AssertExpectations(t)
}

func TestLoginWithOAuth_InvalidCode(t *testing.T) {
	mockRepo := new(MockUserRepository)
	cfg := &config.Config{JWTSecret: "secret"}
	svc := services.NewAuthService(mockRepo, cfg)

	ctx := context.Background()
	provider := "google"
	code := "invalid_code"

	_, _, err := svc.LoginWithOAuth(ctx, provider, code)

	assert.Error(t, err)
	assert.Equal(t, "failed to fetch google user", err.Error())
}
