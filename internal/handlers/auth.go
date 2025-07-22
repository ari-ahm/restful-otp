package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"regexp"

	"github.com/ari-ahm/restful-otp/internal/services"
)

type AuthHandler struct {
	Service services.AuthService
}

func NewAuthHandler(service services.AuthService) *AuthHandler {
	return &AuthHandler{Service: service}
}

type InitiateRequest struct {
	PhoneNumber string `json:"phoneNumber"`
}

type VerifyRequest struct {
	PhoneNumber string `json:"phoneNumber"`
	OTP         string `json:"otp"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func writeError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{Error: message})
}

func (h *AuthHandler) InitiateAuthHandler(w http.ResponseWriter, r *http.Request) {
	var req InitiateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	const phoneRegex = `^\+[1-9]\d{1,14}$`
	re := regexp.MustCompile(phoneRegex)
	if !re.MatchString(req.PhoneNumber) {
		writeError(w, http.StatusBadRequest, "Invalid phone number format. Please use E.164 format (e.g., +989123456789).")
		return
	}

	err := h.Service.InitiateLogin(r.Context(), req.PhoneNumber)
	if err != nil {
		log.Printf("Service error in InitiateLogin: %v", err)
		var rateLimitErr *services.RateLimitError
		if errors.As(err, &rateLimitErr) {
			writeError(w, http.StatusTooManyRequests, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "An internal error occurred while initiating login.")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{"message": "OTP has been sent to your phone number. Please verify to continue."})
}

func (h *AuthHandler) VerifyAuthHandler(w http.ResponseWriter, r *http.Request) {
	var req VerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.PhoneNumber == "" || req.OTP == "" {
		writeError(w, http.StatusBadRequest, "Phone number and OTP are required")
		return
	}

	token, err := h.Service.VerifyLogin(r.Context(), req.PhoneNumber, req.OTP)
	if err != nil {
		log.Printf("Service error in VerifyLogin: %v", err)
		if errors.Is(err, services.ErrInvalidOTP) || errors.Is(err, services.ErrOTPExpired) || errors.Is(err, services.ErrNoPendingOTP) || errors.Is(err, services.ErrTooManyAttempts) {
			writeError(w, http.StatusUnauthorized, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "An internal error occurred during verification.")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}