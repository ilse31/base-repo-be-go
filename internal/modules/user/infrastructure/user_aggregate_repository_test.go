package database

import (
	"context"
	"errors"
	"testing"

	userdomain "github.com/ilse31/base-repo-be-go/internal/modules/user/domain"
	"github.com/ilse31/base-repo-be-go/internal/shared/domain"
	"github.com/ilse31/base-repo-be-go/internal/shared/domain/valueobjects"
)

type mockUserRepo struct {
	users map[string]*userdomain.User
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{users: make(map[string]*userdomain.User)}
}

func (r *mockUserRepo) Create(ctx context.Context, user *userdomain.User) error {
	if user.ID == "" {
		user.ID = valueobjects.GenerateUserID().String()
	}
	userCopy := *user
	r.users[user.ID] = &userCopy
	return nil
}

func (r *mockUserRepo) GetByID(ctx context.Context, id string) (*userdomain.User, error) {
	u, ok := r.users[id]
	if !ok {
		return nil, errors.New("not found")
	}
	userCopy := *u
	return &userCopy, nil
}

func (r *mockUserRepo) GetByEmail(ctx context.Context, email string) (*userdomain.User, error) {
	for _, u := range r.users {
		if u.Email == email {
			userCopy := *u
			return &userCopy, nil
		}
	}
	return nil, errors.New("not found")
}

func (r *mockUserRepo) Update(ctx context.Context, user *userdomain.User) error {
	if _, ok := r.users[user.ID]; !ok {
		return errors.New("not found")
	}
	userCopy := *user
	r.users[user.ID] = &userCopy
	return nil
}

func (r *mockUserRepo) Delete(ctx context.Context, id string) error {
	if _, ok := r.users[id]; !ok {
		return errors.New("not found")
	}
	delete(r.users, id)
	return nil
}

func (r *mockUserRepo) List(ctx context.Context, limit, offset int) ([]*userdomain.User, error) {
	res := make([]*userdomain.User, 0, len(r.users))
	for _, u := range r.users {
		userCopy := *u
		res = append(res, &userCopy)
	}
	return res, nil
}

func TestUserAggregateRepository(t *testing.T) {
	mockRepo := newMockUserRepo()
	aggRepo := NewUserAggregateRepository(mockRepo)
	ctx := context.Background()

	email, _ := valueobjects.NewEmail("user@example.com")
	pwd, _ := valueobjects.NewPassword("Password123!")
	agg, err := userdomain.NewUserAggregate(email, pwd, "Alice")
	if err != nil {
		t.Fatalf("failed to create user aggregate: %v", err)
	}

	// Save new user aggregate
	err = aggRepo.Save(ctx, agg)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// GetByEmail
	foundEmail, err := aggRepo.GetByEmail(ctx, "user@example.com")
	if err != nil || foundEmail.Name() != "Alice" {
		t.Fatalf("GetByEmail failed: %v", err)
	}

	// GetByID
	foundID, err := aggRepo.GetByID(ctx, agg.ID().String())
	if err != nil || foundID.Name() != "Alice" {
		t.Fatalf("GetByID failed: %v", err)
	}

	// Save existing user aggregate (Update)
	_ = agg.UpdateName("Alice Updated")
	err = aggRepo.Save(ctx, agg)
	if err != nil {
		t.Fatalf("Save update failed: %v", err)
	}

	// Delete
	err = aggRepo.Delete(ctx, agg.ID().String())
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err = aggRepo.GetByID(ctx, agg.ID().String())
	if err == nil {
		t.Error("expected error after delete, got nil")
	}
	_, err = aggRepo.GetByEmail(ctx, "user@example.com")
	if err == nil {
		t.Error("expected error after delete, got nil")
	}
}

func TestEntityToAggregateFailures(t *testing.T) {
	mockRepo := newMockUserRepo()
	aggRepo := &UserAggregateRepository{userRepository: mockRepo}

	// Invalid ID
	badEntity1 := &userdomain.User{BaseEntity: domain.BaseEntity{ID: "invalid-uuid"}, Email: "a@b.com", Password: "hash"}
	_, err := aggRepo.entityToAggregate(badEntity1)
	if err == nil {
		t.Error("expected error for invalid user id")
	}

	// Invalid Email
	validID := valueobjects.GenerateUserID().String()
	badEntity2 := &userdomain.User{BaseEntity: domain.BaseEntity{ID: validID}, Email: "invalid-email", Password: "hash"}
	_, err = aggRepo.entityToAggregate(badEntity2)
	if err == nil {
		t.Error("expected error for invalid email")
	}

	// Empty Password Hash
	badEntity3 := &userdomain.User{BaseEntity: domain.BaseEntity{ID: validID}, Email: "a@b.com", Password: ""}
	_, err = aggRepo.entityToAggregate(badEntity3)
	if err == nil {
		t.Error("expected error for empty password hash")
	}
}
