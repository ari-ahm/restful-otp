package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ari-ahm/restful-otp/internal/models"
	"github.com/ari-ahm/restful-otp/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

func TestInitiateLogin(t *testing.T) {
	mockRepo := new(repository.MockUserRepository)
	authService := NewAuthService(mockRepo, "test-secret")
	ctx := context.Background()
	phone := "+989123456789"

	t.Run("Happy Path - New User", func(t *testing.T) {
		mockRepo.On("FindOTPByPhone", ctx, phone).Return(nil, nil).Once()
		mockRepo.On("UpsertOTP", ctx, phone, mock.AnythingOfType("string"), mock.AnythingOfType("time.Time")).Return(nil).Once()

		err := authService.InitiateLogin(ctx, phone)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Rate Limit Error", func(t *testing.T) {
		recentOTP := &models.OTP{
			CreatedAt: time.Now().Add(-10 * time.Second),
		}
		mockRepo.On("FindOTPByPhone", ctx, phone).Return(recentOTP, nil).Once()

		err := authService.InitiateLogin(ctx, phone)

		assert.Error(t, err)
		var rateLimitErr *RateLimitError
		assert.True(t, errors.As(err, &rateLimitErr))
		mockRepo.AssertExpectations(t)
	})
}

func TestVerifyLogin(t *testing.T) {
	mockRepo := new(repository.MockUserRepository)
	authService := NewAuthService(mockRepo, "test-secret")
	ctx := context.Background()
	phone := "+989123456789"
	correctOTP := "123456"

	hashedOTP, _ := bcrypt.GenerateFromPassword([]byte(correctOTP), bcrypt.DefaultCost)

	t.Run("Happy Path - Sign Up", func(t *testing.T) {
		dbOTP := &models.OTP{
			ID:          "otp-id-123",
			OTPHash:     string(hashedOTP),
			ExpiresAt:   time.Now().Add(5 * time.Minute),
		}
		createdUser := &models.User{ID: "user-id-456", PhoneNumber: phone}

		mockRepo.On("FindOTPByPhone", ctx, phone).Return(dbOTP, nil).Once()
		mockRepo.On("DeleteOTP", ctx, dbOTP.ID).Return(nil).Once()
		mockRepo.On("FindUserByPhone", ctx, phone).Return(nil, nil).Once()
		mockRepo.On("CreateUser", ctx, phone).Return(createdUser, nil).Once()

		token, err := authService.VerifyLogin(ctx, phone, correctOTP)

		assert.NoError(t, err)
		assert.NotEmpty(t, token)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Incorrect OTP", func(t *testing.T) {
		dbOTP := &models.OTP{
			ID:          "otp-id-789",
			OTPHash:     string(hashedOTP),
			ExpiresAt:   time.Now().Add(5 * time.Minute),
		}

		mockRepo.On("FindOTPByPhone", ctx, phone).Return(dbOTP, nil).Once()
		mockRepo.On("IncrementFailedAttempts", ctx, dbOTP.ID).Return(nil).Once()

		_, err := authService.VerifyLogin(ctx, phone, "wrong-otp")

		assert.Error(t, err)
		assert.Equal(t, ErrInvalidOTP, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Too Many Failed Attempts", func(t *testing.T) {
		dbOTP := &models.OTP{
			ID:             "otp-id-101",
			OTPHash:        string(hashedOTP),
			ExpiresAt:      time.Now().Add(5 * time.Minute),
			FailedAttempts: 10,
		}

		mockRepo.On("FindOTPByPhone", ctx, phone).Return(dbOTP, nil).Once()
		mockRepo.On("DeleteOTP", ctx, dbOTP.ID).Return(nil).Once()

		_, err := authService.VerifyLogin(ctx, phone, correctOTP)

		assert.Error(t, err)
		assert.Equal(t, ErrTooManyAttempts, err)
		mockRepo.AssertExpectations(t)
	})
}