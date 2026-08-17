package domain

import (
	"time"

	"github.com/google/uuid"
)

type UserModel struct {
	ID           uuid.UUID   `gorm:"type:uuid;primaryKey"`
	Email        string      `gorm:"size:255;uniqueIndex;not null"`
	PasswordHash string      `gorm:"column:password_hash;not null"`
	Name         string      `gorm:"size:255;not null"`
	IsActive     bool        `gorm:"column:is_active;not null"`
	Roles        []RoleModel `gorm:"many2many:user_roles;foreignKey:ID;joinForeignKey:UserID;References:ID;joinReferences:RoleID"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (UserModel) TableName() string { return "users" }

type RoleModel struct {
	ID   uint16 `gorm:"primaryKey"`
	Name string `gorm:"size:50;uniqueIndex;not null"`
}

func (RoleModel) TableName() string { return "roles" }

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

func (u User) RoleNames() []string {
	roles := make([]string, 0, len(u.Roles))
	for _, role := range u.Roles {
		roles = append(roles, role.Name)
	}
	return roles
}
