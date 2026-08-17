package repository

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/jason2071/go-starter-kit-lite/internal/domain"
	"github.com/jason2071/go-starter-kit-lite/internal/usecase"
)

type RefreshTokenRepository struct {
	db *gorm.DB
}

func NewRefreshTokenRepository(db *gorm.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

func (r *RefreshTokenRepository) CreateRefreshToken(
	ctx context.Context,
	token *domain.RefreshToken,
) error {
	model := refreshTokenToModel(token)
	return r.db.WithContext(ctx).Create(&model).Error
}

func (r *RefreshTokenRepository) FindActiveRefreshTokenByHash(
	ctx context.Context,
	hash string,
) (*domain.RefreshToken, error) {
	var model domain.RefreshTokenModel

	err := r.db.WithContext(ctx).
		Where(
			"token_hash = ? AND revoked_at IS NULL AND expires_at > ?",
			hash,
			time.Now().UTC(),
		).
		First(&model).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, usecase.ErrRefreshTokenNotFound
		}
		return nil, err
	}

	return modelToRefreshToken(&model), nil
}

func (r *RefreshTokenRepository) RotateRefreshToken(
	ctx context.Context,
	oldHash string,
	nextToken *domain.RefreshToken,
) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var current domain.RefreshTokenModel

		err := tx.
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where(
				"token_hash = ? AND revoked_at IS NULL AND expires_at > ?",
				oldHash,
				time.Now().UTC(),
			).
			First(&current).
			Error
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

		nextModel := refreshTokenToModel(nextToken)
		return tx.Create(&nextModel).Error
	})
}

func (r *RefreshTokenRepository) RevokeRefreshToken(
	ctx context.Context,
	hash string,
) error {
	now := time.Now().UTC()

	return r.db.WithContext(ctx).
		Model(&domain.RefreshTokenModel{}).
		Where("token_hash = ? AND revoked_at IS NULL", hash).
		Update("revoked_at", now).
		Error
}

func refreshTokenToModel(token *domain.RefreshToken) domain.RefreshTokenModel {
	return domain.RefreshTokenModel{
		ID:        token.ID,
		UserID:    token.UserID,
		TokenHash: token.TokenHash,
		ExpiresAt: token.ExpiresAt,
		RevokedAt: token.RevokedAt,
		CreatedAt: token.CreatedAt,
	}
}

func modelToRefreshToken(model *domain.RefreshTokenModel) *domain.RefreshToken {
	return &domain.RefreshToken{
		ID:        model.ID,
		UserID:    model.UserID,
		TokenHash: model.TokenHash,
		ExpiresAt: model.ExpiresAt,
		RevokedAt: model.RevokedAt,
		CreatedAt: model.CreatedAt,
	}
}
