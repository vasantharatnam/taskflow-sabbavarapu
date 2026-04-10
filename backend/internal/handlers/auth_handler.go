package handlers

import (
	"context"

	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/jackc/pgx"
	"github.com/vasantharatnam/taskflow-sabbavarapu/backend/internal/auth"
	"github.com/vasantharatnam/taskflow-sabbavarapu/backend/internal/models"
	"github.com/vasantharatnam/taskflow-sabbavarapu/backend/internal/repository"
	"github.com/vasantharatnam/taskflow-sabbavarapu/backend/internal/response"
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

type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string      `json:"token"`
	User  AuthUserDTO `json:"user"`
	Message string    `json:"message,omitempty"`
}

type AuthUserDTO struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	fields := validateRegisterRequest(req)
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

	response.WriteJSON(w, http.StatusCreated, AuthResponse{
		Token: token,
		User: AuthUserDTO{
			ID:    user.ID,
			Name:  user.Name,
			Email: user.Email,
		},
		Message: "user registered successfully",
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	fields := validateLoginRequest(req)
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

	response.WriteJSON(w, http.StatusOK, AuthResponse{
		Token: token,
		User: AuthUserDTO{
			ID:    user.ID,
			Name:  user.Name,
			Email: user.Email,
		},
		Message: "user logged in successfully",
	})

}

func validateRegisterRequest(req RegisterRequest) map[string]string {
	fields := make(map[string]string)

	if strings.TrimSpace(req.Name) == "" {
		fields["name"] = "name is required"
	}

	if strings.TrimSpace(req.Email) == "" {
		fields["email"] = "email is required"
	} else if !isValidEmail(req.Email) {
		fields["email"] = "invalid email format"
	}

	if strings.TrimSpace(req.Password) == "" {
		fields["password"] = "is required"
	} else if len(req.Password) < 6 {
		fields["password"] = "must be at least 6 characters"
	}

	return fields
}

func validateLoginRequest(req LoginRequest) map[string]string {
	fields := make(map[string]string)

	if strings.TrimSpace(req.Email) == "" {
		fields["email"] = "is required"
	} else if !isValidEmail(req.Email) {
		fields["email"] = "is invalid"
	}

	if strings.TrimSpace(req.Password) == "" {
		fields["password"] = "is required"
	}
	return fields
}

func isValidEmail(email string) bool {
	email = strings.TrimSpace(email)
	pattern := `^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`
	return regexp.MustCompile(pattern).MatchString(email)
}
