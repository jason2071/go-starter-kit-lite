package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID
	Email        string
	PasswordHash string
	Name         string
	IsActive     bool
	Roles        []Role
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Role struct {
	ID   uint16
	Name string
}

type RefreshToken struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	TokenHash string
	ExpiresAt time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
}

func (u User) RoleNames() []string {
	out := make([]string, 0, len(u.Roles))
	for _, role := range u.Roles {
		out = append(out, role.Name)
	}
	return out
}
