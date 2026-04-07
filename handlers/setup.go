package handlers

import (
	"cromulent/db"
	"cromulent/models"
	"database/sql"
	"errors"
)

type SetupHandler struct{}

func NewSetupHandler() *SetupHandler {
	return &SetupHandler{}
}

func (h *SetupHandler) NeedsSetup() bool {
	var count int
	err := db.DB.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&count)
	if err != nil {
		return true
	}
	return count == 0
}

func (h *SetupHandler) CreateAdminUser(username, password string) error {
	id, err := CreateUser(username, password, models.RoleAdmin)
	if err != nil {
		return err
	}
	return db.EnsureGlobalRuleSet(id)
}

func (h *SetupHandler) UserExists(username string) bool {
	var id int64
	err := db.DB.QueryRow(`SELECT id FROM users WHERE username = ?`, username).Scan(&id)
	return !errors.Is(err, sql.ErrNoRows)
}
