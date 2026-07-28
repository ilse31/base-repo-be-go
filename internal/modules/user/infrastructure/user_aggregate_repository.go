package database

import (
	"context"

	userdomain "github.com/ilse31/base-repo-be-go/internal/modules/user/domain"
	"github.com/ilse31/base-repo-be-go/internal/shared/domain/valueobjects"
)

// UserAggregateRepository implements the UserAggregateRepository interface
// It adapts the existing UserRepository to work with aggregates
type UserAggregateRepository struct {
	userRepository userdomain.UserRepository
}

func NewUserAggregateRepository(userRepository userdomain.UserRepository) userdomain.UserAggregateRepository {
	return &UserAggregateRepository{
		userRepository: userRepository,
	}
}

// Save saves a user aggregate by converting it to entity and using the user repository
func (r *UserAggregateRepository) Save(ctx context.Context, aggregate *userdomain.UserAggregate) error {
	// Convert aggregate to entity
	entity := aggregate.ToEntity()

	// Check if user already exists
	existingUser, err := r.userRepository.GetByEmail(ctx, entity.Email)
	if err == nil && existingUser != nil {
		// Update existing user
		entity.ID = existingUser.ID
		entity.CreatedAt = existingUser.CreatedAt
		return r.userRepository.Update(ctx, &entity)
	}

	// Create new user
	return r.userRepository.Create(ctx, &entity)
}

// GetByID retrieves a user by ID and converts it to aggregate
func (r *UserAggregateRepository) GetByID(ctx context.Context, id string) (*userdomain.UserAggregate, error) {
	entity, err := r.userRepository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return r.entityToAggregate(entity)
}

// GetByEmail retrieves a user by email and converts it to aggregate
func (r *UserAggregateRepository) GetByEmail(ctx context.Context, email string) (*userdomain.UserAggregate, error) {
	entity, err := r.userRepository.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	return r.entityToAggregate(entity)
}

// Delete deletes a user by ID
func (r *UserAggregateRepository) Delete(ctx context.Context, id string) error {
	return r.userRepository.Delete(ctx, id)
}

// entityToAggregate converts a User entity to UserAggregate
func (r *UserAggregateRepository) entityToAggregate(entity *userdomain.User) (*userdomain.UserAggregate, error) {
	// Create value objects
	userID, err := valueobjects.NewUserID(entity.ID)
	if err != nil {
		return nil, err
	}

	email, err := valueobjects.NewEmail(entity.Email)
	if err != nil {
		return nil, err
	}

	password, err := valueobjects.NewPasswordFromHash(entity.Password)
	if err != nil {
		return nil, err
	}

	// Reconstruct aggregate
	aggregate := userdomain.ReconstructUserAggregate(
		userID,
		email,
		password,
		entity.Name,
		entity.CreatedAt,
		entity.UpdatedAt,
	)

	return aggregate, nil
}
