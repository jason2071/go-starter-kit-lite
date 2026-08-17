package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/jason2071/go-starter-kit-lite/internal/domain"
	model "github.com/jason2071/go-starter-kit-lite/internal/models"
)

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
		var role model.RoleModel
		if err := tx.Where("name = ?", roleName).First(&role).Error; err != nil {
			return err
		}

		userModel := userToModel(user)
		userModel.Roles = []model.RoleModel{role}

		if err := tx.Create(&userModel).Error; err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return domain.ErrEmailExists
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
	var model model.UserModel
	err := r.db.WithContext(ctx).
		Preload("Roles").
		First(&model, "id = ?", id).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return modelToUser(&model), nil
}

func (r *UserRepository) FindUserByEmail(
	ctx context.Context,
	email string,
) (*domain.User, error) {
	var model model.UserModel
	err := r.db.WithContext(ctx).
		Preload("Roles").
		First(&model, "email = ?", email).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return modelToUser(&model), nil
}

func (r *UserRepository) ListUsers(
	ctx context.Context,
	opts domain.ListUsersOptions,
) ([]domain.User, int64, error) {
	query := r.db.WithContext(ctx).Model(&model.UserModel{})

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

	var models []model.UserModel
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

func userToModel(user *domain.User) model.UserModel {
	return model.UserModel{
		ID:           user.ID,
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		Name:         user.Name,
		IsActive:     user.IsActive,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
	}
}

func modelToUser(model *model.UserModel) *domain.User {
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
