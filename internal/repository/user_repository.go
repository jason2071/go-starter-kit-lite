package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

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

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) CreateUserWithRole(
	ctx context.Context,
	user *domain.User,
	roleName string,
) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var role RoleModel
		if err := tx.Where("name = ?", roleName).First(&role).Error; err != nil {
			return err
		}

		model := userToModel(user)
		model.Roles = []RoleModel{role}

		if err := tx.Create(&model).Error; err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return usecase.ErrEmailExists
			}
			return err
		}
		return nil
	})
}

func (r *UserRepository) FindUserByID(
	ctx context.Context,
	id uuid.UUID,
) (*domain.User, error) {
	var model UserModel
	err := r.db.WithContext(ctx).
		Preload("Roles").
		First(&model, "id = ?", id).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, usecase.ErrUserNotFound
		}
		return nil, err
	}
	return modelToUser(&model), nil
}

func (r *UserRepository) FindUserByEmail(
	ctx context.Context,
	email string,
) (*domain.User, error) {
	var model UserModel
	err := r.db.WithContext(ctx).
		Preload("Roles").
		First(&model, "email = ?", email).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, usecase.ErrUserNotFound
		}
		return nil, err
	}
	return modelToUser(&model), nil
}

func (r *UserRepository) ListUsers(
	ctx context.Context,
	opts usecase.ListUsersOptions,
) ([]domain.User, int64, error) {
	query := r.db.WithContext(ctx).Model(&UserModel{})

	if opts.Search != "" {
		search := "%" + opts.Search + "%"
		query = query.Where("email ILIKE ? OR name ILIKE ?", search, search)
	}

	if opts.IsActive != nil {
		query = query.Where("is_active = ?", *opts.IsActive)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	sortColumns := map[string]string{
		"created_at": "created_at",
		"updated_at": "updated_at",
		"email":      "email",
		"name":       "name",
	}
	sortColumn := sortColumns[opts.Sort]
	if sortColumn == "" {
		sortColumn = "created_at"
	}

	order := "DESC"
	if opts.Order == "asc" {
		order = "ASC"
	}

	var models []UserModel
	err := query.
		Preload("Roles").
		Order(fmt.Sprintf("%s %s", sortColumn, order)).
		Limit(opts.Page.PageSize).
		Offset(opts.Page.Offset()).
		Find(&models).
		Error
	if err != nil {
		return nil, 0, err
	}

	users := make([]domain.User, 0, len(models))
	for i := range models {
		users = append(users, *modelToUser(&models[i]))
	}
	return users, total, nil
}

func userToModel(user *domain.User) UserModel {
	return UserModel{
		ID:           user.ID,
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		Name:         user.Name,
		IsActive:     user.IsActive,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
	}
}

func modelToUser(model *UserModel) *domain.User {
	roles := make([]domain.Role, 0, len(model.Roles))
	for _, role := range model.Roles {
		roles = append(roles, domain.Role{ID: role.ID, Name: role.Name})
	}

	return &domain.User{
		ID:           model.ID,
		Email:        model.Email,
		PasswordHash: model.PasswordHash,
		Name:         model.Name,
		IsActive:     model.IsActive,
		Roles:        roles,
		CreatedAt:    model.CreatedAt,
		UpdatedAt:    model.UpdatedAt,
	}
}
