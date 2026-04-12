package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/jackc/pgx"
	"github.com/vasantharatnam/taskflow-sabbavarapu/backend/internal/api"
	"github.com/vasantharatnam/taskflow-sabbavarapu/backend/internal/auth"
	"github.com/vasantharatnam/taskflow-sabbavarapu/backend/internal/models"
	"github.com/vasantharatnam/taskflow-sabbavarapu/backend/internal/repository"
	"github.com/vasantharatnam/taskflow-sabbavarapu/backend/internal/response"
	"github.com/vasantharatnam/taskflow-sabbavarapu/backend/internal/utils"
)

type AuthHandler struct {
	userRepo       *repository.UserRepository
	jwtSecret      string
	jwtExpiryHours int
}

func NewAuthHandler(userRepo *repository.UserRepository, jwtSecret string, jwtExpiryHours int) *AuthHandler {
	return &AuthHandler{
		userRepo:       userRepo,
		jwtSecret:      jwtSecret,
		jwtExpiryHours: jwtExpiryHours,
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req api.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	fields := utils.ValidateRegisterRequest(req)
	if len(fields) > 0 {
		response.WriteValidationError(w, fields)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	email := strings.ToLower(strings.TrimSpace(req.Email))

	userExists, err := h.userRepo.IsUserExists(ctx, email)

	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	if userExists {
		response.WriteValidationError(w, map[string]string{
			"email": "already exists",
		})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "failed to hashpassword")
		return
	}

	user := &models.User{
		Name:         req.Name,
		Email:        email,
		PasswordHash: string(hashedPassword),
	}

	if err := h.userRepo.CreateUser(ctx, user); err != nil {
		response.WriteError(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	token, err := auth.GenerateToken(h.jwtSecret, user.ID, user.Email, h.jwtExpiryHours)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	response.WriteJSON(w, http.StatusCreated, api.AuthResponse{
		Token: token,
		User: api.AuthUserDTO{
			ID:    user.ID,
			Name:  user.Name,
			Email: user.Email,
		},
		Message: "user registered successfully",
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req api.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	fields := utils.ValidateLoginRequest(req)
	if len(fields) > 0 {
		response.WriteValidationError(w, fields)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	user, err := h.userRepo.GetUserByEmail(ctx, strings.ToLower(strings.TrimSpace(req.Email)))

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			response.WriteError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		response.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	token, err := auth.GenerateToken(h.jwtSecret, user.ID, user.Email, h.jwtExpiryHours)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	response.WriteJSON(w, http.StatusOK, api.AuthResponse{
		Token: token,
		User: api.AuthUserDTO{
			ID:    user.ID,
			Name:  user.Name,
			Email: user.Email,
		},
		Message: "user logged in successfully",
	})

}

