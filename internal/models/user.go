package models

import (
	"time"
)

type User struct {
	ID          string    `json:"id"`
	PhoneNumber string    `json:"phoneNumber"`
	CreatedAt   time.Time `json:"createdAt"`
}

type OTP struct {
	ID             string    `json:"id"`
	PhoneNumber    string    `json:"phoneNumber"`
	OTPHash        string    `json:"-"`
	ExpiresAt      time.Time `json:"expiresAt"`
	FailedAttempts int       `json:"-"`
	CreatedAt      time.Time `json:"createdAt"`
}