package models

import "time"

type User struct {
	ID           int64
	Username     string
	PasswordHash string
	Role         string
	Active       bool
	CreatedAt    time.Time
}

type UserResponse struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	Role      string `json:"role"`
	Active    bool   `json:"active"`
	CreatedAt string `json:"created_at"`
}

func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:        u.ID,
		Username:  u.Username,
		Role:      u.Role,
		Active:    u.Active,
		CreatedAt: u.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

type Role string

const (
	RoleTechnician Role = "technician"
	RoleReviewer   Role = "reviewer"
	RoleSupervisor Role = "supervisor"
	RoleAdmin      Role = "admin"
)
