package memory

import (
	"context"
	"sync"

	"HaruhiServer/internal/domain"
)

type UserRepository struct {
	mu         sync.RWMutex
	byID       map[domain.UserID]*domain.User
	usernameID map[string]domain.UserID
}

func NewUserRepository() *UserRepository {
	return &UserRepository{
		byID:       make(map[domain.UserID]*domain.User),
		usernameID: make(map[string]domain.UserID),
	}
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	if err := checkContext(ctx); err != nil {
		return err
	}
	if user == nil {
		return domain.NewDomainError(domain.ErrInvalidArgument, "user is required")
	}
	if err := user.Validate(); err != nil {
		return err
	}

	usernameKey := indexKey(user.Username)

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.byID[user.ID]; ok {
		return domain.NewDomainError(domain.ErrConflict, "user id already exists")
	}
	if _, ok := r.usernameID[usernameKey]; ok {
		return domain.NewDomainError(domain.ErrConflict, "username already exists")
	}

	r.byID[user.ID] = cloneUser(user)
	r.usernameID[usernameKey] = user.ID
	return nil
}

func (r *UserRepository) Update(ctx context.Context, user *domain.User) error {
	if err := checkContext(ctx); err != nil {
		return err
	}
	if user == nil {
		return domain.NewDomainError(domain.ErrInvalidArgument, "user is required")
	}
	if err := user.Validate(); err != nil {
		return err
	}

	newUsernameKey := indexKey(user.Username)

	r.mu.Lock()
	defer r.mu.Unlock()

	old, ok := r.byID[user.ID]
	if !ok {
		return domain.NewDomainError(domain.ErrNotFound, "user not found")
	}

	oldUsernameKey := indexKey(old.Username)
	if oldUsernameKey != newUsernameKey {
		if ownerID, exists := r.usernameID[newUsernameKey]; exists && ownerID != user.ID {
			return domain.NewDomainError(domain.ErrConflict, "username already exists")
		}
		delete(r.usernameID, oldUsernameKey)
		r.usernameID[newUsernameKey] = user.ID
	}

	r.byID[user.ID] = cloneUser(user)
	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id domain.UserID) (*domain.User, error) {
	if err := checkContext(ctx); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	user, ok := r.byID[id]
	if !ok {
		return nil, domain.NewDomainError(domain.ErrNotFound, "user not found")
	}
	return cloneUser(user), nil
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	if err := checkContext(ctx); err != nil {
		return nil, err
	}

	key := indexKey(username)

	r.mu.RLock()
	defer r.mu.RUnlock()

	id, ok := r.usernameID[key]
	if !ok {
		return nil, domain.NewDomainError(domain.ErrNotFound, "user not found")
	}

	user, ok := r.byID[id]
	if !ok {
		return nil, domain.NewDomainError(domain.ErrNotFound, "user not found")
	}

	return cloneUser(user), nil
}

func (r *UserRepository) Delete(ctx context.Context, id domain.UserID) error {
	if err := checkContext(ctx); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	user, ok := r.byID[id]
	if !ok {
		return domain.NewDomainError(domain.ErrNotFound, "user not found")
	}

	delete(r.byID, id)
	delete(r.usernameID, indexKey(user.Username))
	return nil
}

func (r *UserRepository) List(ctx context.Context) ([]*domain.User, error) {
	if err := checkContext(ctx); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	items := make([]*domain.User, 0, len(r.byID))
	for _, user := range r.byID {
		items = append(items, cloneUser(user))
	}

	sortUsers(items)
	return items, nil
}
