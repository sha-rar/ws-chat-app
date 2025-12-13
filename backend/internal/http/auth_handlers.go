package http

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/sha-rar/ws-chat-app/dev/backend/internal/auth"
	"github.com/sha-rar/ws-chat-app/dev/backend/internal/db"
	"github.com/sha-rar/ws-chat-app/dev/backend/internal/models"
)

type authRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type authResponse struct {
	Token string      `json:"token"`
	User  models.User `json:"user"`
}

func handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	username := strings.TrimSpace(req.Username)
	password := strings.TrimSpace(req.Password)

	if username == "" || password == "" {
		http.Error(w, "username and password required", http.StatusBadRequest)
		return
	}

	if len(password) < 6 {
		http.Error(w, "password too short (min 6)", http.StatusBadRequest)
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "failed to hash password", http.StatusInternalServerError)
		return
	}

	query := `
		INSERT INTO users (username, password_hash)
		VALUES ($1, $2)
		RETURNING id, username, created_at
	`
	row := db.DB.QueryRowContext(r.Context(), query, username, string(hashed))

	var user models.User
	if err := row.Scan(&user.ID, &user.Username, &user.CreatedAt); err != nil {
		// unique violation etc.
		http.Error(w, "could not create user (username may be taken)", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	// For simplicity we don't auto-login here; we just return the created user
	_ = json.NewEncoder(w).Encode(user)
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	username := strings.TrimSpace(req.Username)
	password := strings.TrimSpace(req.Password)

	if username == "" || password == "" {
		http.Error(w, "username and password required", http.StatusBadRequest)
		return
	}

	var user models.User
	var passwordHash string

	query := `
		SELECT id, username, password_hash, created_at
		FROM users
		WHERE username = $1
	`
	err := db.DB.QueryRowContext(r.Context(), query, username).
		Scan(&user.ID, &user.Username, &passwordHash, &user.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}
		http.Error(w, "failed to query user", http.StatusInternalServerError)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)); err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	token, err := auth.GenerateToken(user.ID, user.Username)
	if err != nil {
		http.Error(w, "failed to generate token", http.StatusInternalServerError)
		return
	}

	resp := authResponse{
		Token: token,
		User:  user,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}
