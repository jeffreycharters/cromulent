package handlers

import (
	"cromulent/db"
	"cromulent/models"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	sessionUser    *models.User
	lastActivityAt time.Time
}

const sessionTimeout = 30 * time.Minute

func NewAuthHandler() *AuthHandler {
	return &AuthHandler{}
}

func (h *AuthHandler) Login(username, password string) (*models.UserResponse, error) {
	row := db.DB.QueryRow(
		`SELECT id, username, password_hash, role, created_at FROM users WHERE username = ? AND active = 1`,
		username,
	)

	var u models.User
	if err := row.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Role, &u.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("invalid credentials")
		}
		return nil, fmt.Errorf("login: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	h.sessionUser = &u
	h.lastActivityAt = time.Now()
	r := u.ToResponse()
	return &r, nil
}

func (h *AuthHandler) Logout() {
	h.sessionUser = nil
}

func (h *AuthHandler) CurrentUser() *models.UserResponse {
	if h.sessionUser == nil {
		return nil
	}
	if time.Since(h.lastActivityAt) > sessionTimeout {
		h.sessionUser = nil
		return nil
	}
	h.lastActivityAt = time.Now()
	r := h.sessionUser.ToResponse()
	return &r
}

func (h *AuthHandler) IsAuthenticated() bool {
	return h.CurrentUser() != nil
}

func CreateUser(username, password string, role models.Role) (int64, error) {
	if len(password) < 6 {
		return 0, fmt.Errorf("password must be at least 6 characters")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, fmt.Errorf("hash password: %w", err)
	}
	result, err := db.DB.Exec(
		`INSERT INTO users (username, password_hash, role) VALUES (?, ?, ?)`,
		username, string(hash), role,
	)
	if err != nil {
		return 0, fmt.Errorf("create user: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("last insert id: %w", err)
	}
	return id, nil
}

func (h *AuthHandler) ListUsers() ([]models.UserResponse, error) {
	rows, err := db.DB.Query(
		`SELECT id, username, role, active, created_at FROM users ORDER BY username`,
	)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	var users []models.UserResponse
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Username, &u.Role, &u.Active, &u.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, u.ToResponse())
	}
	return users, nil
}

func (h *AuthHandler) DeactivateUser(id int64) error {
	_, err := db.DB.Exec(`UPDATE users SET active = 0 WHERE id = ?`, id)
	return err
}

func (h *AuthHandler) ActivateUser(id int64) error {
	_, err := db.DB.Exec(`UPDATE users SET active = 1 WHERE id = ?`, id)
	return err
}

func (h *AuthHandler) CreateUser(username, password, role string) error {
	_, err := CreateUser(username, password, models.Role(role))
	return err
}
