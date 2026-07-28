package application

import (
	"context"
	"errors"
	"testing"

	userdomain "github.com/ilse31/base-repo-be-go/internal/modules/user/domain"
	"github.com/ilse31/base-repo-be-go/internal/shared/domain"
)

type mockUserRepo struct {
	users map[string]*userdomain.User
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{users: make(map[string]*userdomain.User)}
}

func (r *mockUserRepo) Create(ctx context.Context, user *userdomain.User) error {
	if user.ID == "" {
		user.ID = "generated-id"
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

func TestUserService(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewUserService(repo)
	ctx := context.Background()

	// Create User
	u1 := &userdomain.User{
		BaseEntity: domain.BaseEntity{ID: "user-1"},
		Name:       "Alice",
		Email:      "alice@example.com",
	}

	err := svc.CreateUser(ctx, u1)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	// Create User duplicate email
	err = svc.CreateUser(ctx, &userdomain.User{
		BaseEntity: domain.BaseEntity{ID: "user-2"},
		Name:       "Alice2",
		Email:      "alice@example.com",
	})
	if !errors.Is(err, ErrUserExists) {
		t.Errorf("expected ErrUserExists, got %v", err)
	}

	// GetByID
	got, err := svc.GetUserByID(ctx, "user-1")
	if err != nil || got.Name != "Alice" {
		t.Errorf("GetUserByID failed: %v", err)
	}
	_, err = svc.GetUserByID(ctx, "non-existent")
	if !errors.Is(err, ErrUserNotFound) {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}

	// GetByEmail
	gotEmail, err := svc.GetUserByEmail(ctx, "alice@example.com")
	if err != nil || gotEmail.ID != "user-1" {
		t.Errorf("GetUserByEmail failed: %v", err)
	}
	_, err = svc.GetUserByEmail(ctx, "wrong@example.com")
	if !errors.Is(err, ErrUserNotFound) {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}

	// UpdateUser
	u1.Name = "Alice Updated"
	err = svc.UpdateUser(ctx, u1)
	if err != nil {
		t.Fatalf("UpdateUser failed: %v", err)
	}
	err = svc.UpdateUser(ctx, &userdomain.User{BaseEntity: domain.BaseEntity{ID: "non-existent"}})
	if !errors.Is(err, ErrUserNotFound) {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}

	// ListUsers
	list, err := svc.ListUsers(ctx, 10, 0)
	if err != nil || len(list) != 1 {
		t.Errorf("ListUsers failed: %v, len: %d", err, len(list))
	}

	// DeleteUser
	err = svc.DeleteUser(ctx, "non-existent")
	if !errors.Is(err, ErrUserNotFound) {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
	err = svc.DeleteUser(ctx, "user-1")
	if err != nil {
		t.Fatalf("DeleteUser failed: %v", err)
	}
}
