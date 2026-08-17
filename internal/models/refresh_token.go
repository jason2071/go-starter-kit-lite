package model

import (
	"time"

	"github.com/google/uuid"
)

// RefreshTokenModel is the GORM persistence model for refresh tokens.
type RefreshTokenModel struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey"`
	UserID    uuid.UUID  `gorm:"type:uuid;column:user_id;index;not null"`
	TokenHash string     `gorm:"column:token_hash;size:64;uniqueIndex;not null"`
	ExpiresAt time.Time  `gorm:"column:expires_at;index;not null"`
	RevokedAt *time.Time `gorm:"column:revoked_at"`
	CreatedAt time.Time  `gorm:"column:created_at;autoCreateTime"`
}

func (RefreshTokenModel) TableName() string { return "refresh_tokens" }
