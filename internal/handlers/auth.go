package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"confluent/internal/auth"
	"confluent/internal/models"

	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	db        *sql.DB
	jwtSecret string
}

func NewAuthHandler(db *sql.DB, jwtSecret string) *AuthHandler {
	return &AuthHandler{db: db, jwtSecret: jwtSecret}
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type authResponse struct {
	Token string       `json:"token"`
	User  *models.User `json:"user"`
}

// Register creates a new team member account.
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.FullName = strings.TrimSpace(req.FullName)
	if req.Email == "" || len(req.Password) < 6 || req.FullName == "" {
		Error(w, http.StatusBadRequest, "Name, valid email, and password (min 6 chars) are required")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Failed to hash password")
		return
	}

	res, err := h.db.Exec(`
		INSERT INTO users (email, password_hash, full_name, role)
		VALUES (?, ?, ?, 'editor')
	`, req.Email, string(hash), req.FullName)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			Error(w, http.StatusConflict, "An account with this email already exists")
			return
		}
		Error(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	id, _ := res.LastInsertId()
	user := &models.User{
		ID:        id,
		Email:     req.Email,
		FullName:  req.FullName,
		Role:      models.RoleEditor,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	token, err := auth.GenerateToken(user, h.jwtSecret)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Failed to generate session token")
		return
	}

	JSON(w, http.StatusCreated, authResponse{Token: token, User: user})
}

// Login verifies credentials and issues a JWT session token.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	user := &models.User{}
	err := h.db.QueryRow(`
		SELECT id, email, password_hash, full_name, role, created_at, updated_at
		FROM users WHERE email = ?
	`, req.Email).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.FullName, &user.Role, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			Error(w, http.StatusUnauthorized, "Invalid email or password")
			return
		}
		Error(w, http.StatusInternalServerError, "Database error")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		Error(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	token, err := auth.GenerateToken(user, h.jwtSecret)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	JSON(w, http.StatusOK, authResponse{Token: token, User: user})
}

// Me returns the profile of the current authenticated user.
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	if user == nil {
		Error(w, http.StatusUnauthorized, "Not authenticated")
		return
	}
	JSON(w, http.StatusOK, user)
}
