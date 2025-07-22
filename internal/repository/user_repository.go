package repository

import (
	"context"
	"time"

	"github.com/ari-ahm/restful-otp/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository interface {
	FindUserByPhone(ctx context.Context, phone string) (*models.User, error)
	CreateUser(ctx context.Context, phone string) (*models.User, error)
	UpsertOTP(ctx context.Context, phone, otpHash string, expiresAt time.Time) error
	FindOTPByPhone(ctx context.Context, phone string) (*models.OTP, error)
	DeleteOTP(ctx context.Context, otpID string) error
	IncrementFailedAttempts(ctx context.Context, otpID string) error
}

type pgUserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) UserRepository {
	return &pgUserRepository{db: db}
}

func (r *pgUserRepository) FindUserByPhone(ctx context.Context, phone string) (*models.User, error) {
	var user models.User
	err := r.db.QueryRow(ctx, "SELECT id, phone_number, created_at FROM users WHERE phone_number = $1", phone).Scan(&user.ID, &user.PhoneNumber, &user.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return &user, err
}

func (r *pgUserRepository) CreateUser(ctx context.Context, phone string) (*models.User, error) {
	var user models.User
	err := r.db.QueryRow(ctx, "INSERT INTO users (phone_number) VALUES ($1) RETURNING id, phone_number, created_at", phone).Scan(&user.ID, &user.PhoneNumber, &user.CreatedAt)
	return &user, err
}

func (r *pgUserRepository) UpsertOTP(ctx context.Context, phone, otpHash string, expiresAt time.Time) error {
	query := `
        INSERT INTO otps (phone_number, otp_hash, expires_at, created_at, failed_attempts)
        VALUES ($1, $2, $3, NOW(), 0)
        ON CONFLICT (phone_number)
        DO UPDATE SET
            otp_hash = EXCLUDED.otp_hash,
            expires_at = EXCLUDED.expires_at,
            created_at = NOW(),
            failed_attempts = 0;`
	_, err := r.db.Exec(ctx, query, phone, otpHash, expiresAt)
	return err
}

func (r *pgUserRepository) FindOTPByPhone(ctx context.Context, phone string) (*models.OTP, error) {
	var otp models.OTP
	query := "SELECT id, phone_number, otp_hash, expires_at, created_at, failed_attempts FROM otps WHERE phone_number = $1"
	err := r.db.QueryRow(ctx, query, phone).Scan(&otp.ID, &otp.PhoneNumber, &otp.OTPHash, &otp.ExpiresAt, &otp.CreatedAt, &otp.FailedAttempts)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return &otp, err
}

func (r *pgUserRepository) DeleteOTP(ctx context.Context, otpID string) error {
	_, err := r.db.Exec(ctx, "DELETE FROM otps WHERE id = $1", otpID)
	return err
}

func (r *pgUserRepository) IncrementFailedAttempts(ctx context.Context, otpID string) error {
	query := "UPDATE otps SET failed_attempts = failed_attempts + 1 WHERE id = $1"
	_, err := r.db.Exec(ctx, query, otpID)
	return err
}