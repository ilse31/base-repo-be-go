package database

import (
	"context"

	"github.com/google/uuid"
	"github.com/uptrace/bun"

	userdomain "github.com/ilse31/base-repo-be-go/internal/modules/user/domain"
)

type userRepository struct {
	db *bun.DB
}

func NewUserRepository(db *bun.DB) userdomain.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *userdomain.User) error {
	if user.ID == "" {
		user.ID = uuid.New().String()
	}

	_, err := r.db.NewInsert().
		Model(user).
		Exec(ctx)
	return err
}

func (r *userRepository) GetByID(ctx context.Context, id string) (*userdomain.User, error) {
	user := &userdomain.User{}
	err := r.db.NewSelect().
		Model(user).
		Where("id = ?", id).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*userdomain.User, error) {
	user := &userdomain.User{}
	err := r.db.NewSelect().
		Model(user).
		Where("email = ?", email).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *userRepository) Update(ctx context.Context, user *userdomain.User) error {
	_, err := r.db.NewUpdate().
		Model(user).
		Where("id = ?", user.ID).
		Exec(ctx)
	return err
}

func (r *userRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.NewDelete().
		Model(&userdomain.User{}).
		Where("id = ?", id).
		Exec(ctx)
	return err
}

func (r *userRepository) List(ctx context.Context, limit, offset int) ([]*userdomain.User, error) {
	var users []*userdomain.User
	err := r.db.NewSelect().
		Model(&users).
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return users, nil
}
