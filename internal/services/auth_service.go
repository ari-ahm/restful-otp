package services

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/ari-ahm/restful-otp/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidOTP      = errors.New("invalid otp")
	ErrOTPExpired      = errors.New("otp has expired")
	ErrNoPendingOTP    = errors.New("no pending otp found for this number")
	ErrInternal        = errors.New("an internal error occurred")
	ErrTooManyAttempts = errors.New("too many failed attempts, please request a new otp")
)

type RateLimitError struct{ WaitTime time.Duration }

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("You must wait %d seconds before requesting another OTP.", int(math.Ceil(e.WaitTime.Seconds())))
}

type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

type AuthService interface {
	InitiateLogin(ctx context.Context, phone string) error
	VerifyLogin(ctx context.Context, phone, otp string) (string, error)
}

type authService struct {
	repo      repository.UserRepository
	jwtSecret []byte
}

func NewAuthService(repo repository.UserRepository, jwtSecret string) AuthService {
	return &authService{repo: repo, jwtSecret: []byte(jwtSecret)}
}

func (s *authService) InitiateLogin(ctx context.Context, phone string) error {
	const otpRateLimit = 60 * time.Second
	existingOTP, err := s.repo.FindOTPByPhone(ctx, phone)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		log.Printf("repository.FindOTPByPhone error: %v", err)
		return ErrInternal
	}

	if existingOTP != nil {
		if time.Since(existingOTP.CreatedAt) < otpRateLimit {
			return &RateLimitError{WaitTime: otpRateLimit - time.Since(existingOTP.CreatedAt)}
		}
	}

	otp, err := generateOTP()
	if err != nil {
		log.Printf("generateOTP error: %v", err)
		return ErrInternal
	}
	otpHash, err := bcrypt.GenerateFromPassword([]byte(otp), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("bcrypt.GenerateFromPassword error: %v", err)
		return ErrInternal
	}
	expiresAt := time.Now().Add(5 * time.Minute)

	if err := s.repo.UpsertOTP(ctx, phone, string(otpHash), expiresAt); err != nil {
		log.Printf("repository.UpsertOTP error: %v", err)
		return ErrInternal
	}

	log.Printf("Simulating sending OTP %s to %s", otp, phone)
	return nil
}

func (s *authService) VerifyLogin(ctx context.Context, phone, otp string) (string, error) {
	const maxAttempts = 10
	dbOTP, err := s.repo.FindOTPByPhone(ctx, phone)
	if err != nil {
		log.Printf("repository.FindOTPByPhone error: %v", err)
		return "", ErrInternal
	}
	if dbOTP == nil {
		return "", ErrNoPendingOTP
	}

	if time.Now().After(dbOTP.ExpiresAt) {
		return "", ErrOTPExpired
	}

	if dbOTP.FailedAttempts >= maxAttempts {
		return "", ErrTooManyAttempts
	}

	if err := bcrypt.CompareHashAndPassword([]byte(dbOTP.OTPHash), []byte(otp)); err != nil {
		if err := s.repo.IncrementFailedAttempts(ctx, dbOTP.ID); err != nil {
			log.Printf("repository.IncrementFailedAttempts error: %v", err)
			return "", ErrInternal
		}
		return "", ErrInvalidOTP
	}

	if err := s.repo.DeleteOTP(ctx, dbOTP.ID); err != nil {
		log.Printf("repository.DeleteOTP error: %v", err)
		return "", ErrInternal
	}

	user, err := s.repo.FindUserByPhone(ctx, phone)
	if err != nil {
		log.Printf("repository.FindUserByPhone error: %v", err)
		return "", ErrInternal
	}
	if user == nil {
		user, err = s.repo.CreateUser(ctx, phone)
		if err != nil {
			log.Printf("repository.CreateUser error: %v", err)
			return "", ErrInternal
		}
		log.Printf("New user created with ID %s for phone number %s", user.ID, phone)
	}

	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		UserID:         user.ID,
		RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(expirationTime)},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		log.Printf("token.SignedString error: %v", err)
		return "", ErrInternal
	}
	return tokenString, nil
}

func generateOTP() (string, error) {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	val := (int(b[0])<<24 | int(b[1])<<16 | int(b[2])<<8 | int(b[3])) % 900000
	return fmt.Sprintf("%06d", 100000+val), nil
}
