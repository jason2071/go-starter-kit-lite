package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/jason2071/go-starter-kit-lite/internal/domain"
	"github.com/jason2071/go-starter-kit-lite/internal/usecase"
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

type RefreshTokenModel struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey"`
	UserID    uuid.UUID  `gorm:"type:uuid;column:user_id;index;not null"`
	TokenHash string     `gorm:"column:token_hash;size:64;uniqueIndex;not null"`
	ExpiresAt time.Time  `gorm:"column:expires_at;index;not null"`
	RevokedAt *time.Time `gorm:"column:revoked_at"`
	CreatedAt time.Time  `gorm:"column:created_at;autoCreateTime"`
}

func (RefreshTokenModel) TableName() string { return "refresh_tokens" }

type Postgres struct{ db *gorm.DB }

func NewPostgres(db *gorm.DB) *Postgres { return &Postgres{db: db} }

func (r *Postgres) CreateWithRole(ctx context.Context, user *domain.User, role string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var roleModel RoleModel
		if err := tx.Where("name = ?", role).First(&roleModel).Error; err != nil {
			return err
		}
		model := userToModel(user)
		model.Roles = []RoleModel{roleModel}
		if err := tx.Create(&model).Error; err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return usecase.ErrEmailExists
			}
			return err
		}
		return nil
	})
}

func (r *Postgres) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	var model UserModel
	if err := r.db.WithContext(ctx).Preload("Roles").First(&model, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, usecase.ErrUserNotFound
		}
		return nil, err
	}
	return modelToUser(&model), nil
}

func (r *Postgres) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var model UserModel
	if err := r.db.WithContext(ctx).Preload("Roles").First(&model, "email = ?", email).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, usecase.ErrUserNotFound
		}
		return nil, err
	}
	return modelToUser(&model), nil
}

func (r *Postgres) List(ctx context.Context, opts usecase.UserListOptions) ([]domain.User, int64, error) {
	query := r.db.WithContext(ctx).Model(&UserModel{})
	if opts.Search != "" {
		q := "%" + opts.Search + "%"
		query = query.Where("email ILIKE ? OR name ILIKE ?", q, q)
	}
	if opts.IsActive != nil {
		query = query.Where("is_active = ?", *opts.IsActive)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	sortColumns := map[string]string{"created_at": "created_at", "updated_at": "updated_at", "email": "email", "name": "name"}
	sortColumn := sortColumns[opts.Sort]
	if sortColumn == "" {
		sortColumn = "created_at"
	}
	order := "DESC"
	if opts.Order == "asc" {
		order = "ASC"
	}

	var models []UserModel
	if err := query.Preload("Roles").Order(fmt.Sprintf("%s %s", sortColumn, order)).Limit(opts.Page.PageSize).Offset(opts.Page.Offset()).Find(&models).Error; err != nil {
		return nil, 0, err
	}
	users := make([]domain.User, 0, len(models))
	for i := range models {
		users = append(users, *modelToUser(&models[i]))
	}
	return users, total, nil
}

func (r *Postgres) Create(ctx context.Context, token *domain.RefreshToken) error {
	model := refreshToModel(token)
	return r.db.WithContext(ctx).Create(&model).Error
}

func (r *Postgres) FindActiveByHash(ctx context.Context, hash string) (*domain.RefreshToken, error) {
	var model RefreshTokenModel
	err := r.db.WithContext(ctx).
		Where("token_hash = ? AND revoked_at IS NULL AND expires_at > ?", hash, time.Now().UTC()).
		First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, usecase.ErrRefreshNotFound
		}
		return nil, err
	}
	return modelToRefresh(&model), nil
}

func (r *Postgres) Rotate(ctx context.Context, oldHash string, next *domain.RefreshToken) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var current RefreshTokenModel
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("token_hash = ? AND revoked_at IS NULL AND expires_at > ?", oldHash, time.Now().UTC()).
			First(&current).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return usecase.ErrRefreshTokenUsed
			}
			return err
		}
		now := time.Now().UTC()
		if err := tx.Model(&current).Update("revoked_at", now).Error; err != nil {
			return err
		}
		model := refreshToModel(next)
		return tx.Create(&model).Error
	})
}

func (r *Postgres) Revoke(ctx context.Context, hash string) error {
	now := time.Now().UTC()
	return r.db.WithContext(ctx).Model(&RefreshTokenModel{}).
		Where("token_hash = ? AND revoked_at IS NULL", hash).
		Update("revoked_at", now).Error
}

func userToModel(u *domain.User) UserModel {
	return UserModel{ID: u.ID, Email: u.Email, PasswordHash: u.PasswordHash, Name: u.Name, IsActive: u.IsActive, CreatedAt: u.CreatedAt, UpdatedAt: u.UpdatedAt}
}

func modelToUser(m *UserModel) *domain.User {
	roles := make([]domain.Role, 0, len(m.Roles))
	for _, role := range m.Roles {
		roles = append(roles, domain.Role{ID: role.ID, Name: role.Name})
	}
	return &domain.User{ID: m.ID, Email: m.Email, PasswordHash: m.PasswordHash, Name: m.Name, IsActive: m.IsActive, Roles: roles, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt}
}

func refreshToModel(t *domain.RefreshToken) RefreshTokenModel {
	return RefreshTokenModel{ID: t.ID, UserID: t.UserID, TokenHash: t.TokenHash, ExpiresAt: t.ExpiresAt, RevokedAt: t.RevokedAt, CreatedAt: t.CreatedAt}
}

func modelToRefresh(m *RefreshTokenModel) *domain.RefreshToken {
	return &domain.RefreshToken{ID: m.ID, UserID: m.UserID, TokenHash: m.TokenHash, ExpiresAt: m.ExpiresAt, RevokedAt: m.RevokedAt, CreatedAt: m.CreatedAt}
}
