package model

import (
	"time"

	"github.com/google/uuid"
)

// UserModel and RoleModel are GORM persistence models. They must not leak
// into the domain layer.
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
