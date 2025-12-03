package unit_test

import (
	"context"
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
