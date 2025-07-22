package repository

import (
	"context"
	"time"

	"github.com/ari-ahm/restful-otp/internal/models"
	"github.com/stretchr/testify/mock"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) FindUserByPhone(ctx context.Context, phone string) (*models.User, error) {
	args := m.Called(ctx, phone)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) CreateUser(ctx context.Context, phone string) (*models.User, error) {
	args := m.Called(ctx, phone)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) UpsertOTP(ctx context.Context, phone, otpHash string, expiresAt time.Time) error {
	args := m.Called(ctx, phone, otpHash, expiresAt)
	return args.Error(0)
}

func (m *MockUserRepository) FindOTPByPhone(ctx context.Context, phone string) (*models.OTP, error) {
	args := m.Called(ctx, phone)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.OTP), args.Error(1)
}

func (m *MockUserRepository) DeleteOTP(ctx context.Context, otpID string) error {
	args := m.Called(ctx, otpID)
	return args.Error(0)
}

func (m *MockUserRepository) IncrementFailedAttempts(ctx context.Context, otpID string) error {
	args := m.Called(ctx, otpID)
	return args.Error(0)
}