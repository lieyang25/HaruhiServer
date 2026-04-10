package service

import (
	"context"
	"fmt"

	"HaruhiServer/internal/domain"
)

type userService struct {
	deps deps
}

var _ UserService = (*userService)(nil)

func (s *userService) Create(ctx context.Context, in CreateUserInput) (*domain.User, error) {
	now := s.deps.now()

	role := in.Role
	if role == "" {
		role = domain.UserRoleMember
	}

	user := &domain.User{
		ID:          domain.UserID(s.deps.idg.NewID()),
		Username:    in.Username,
		DisplayName: in.DisplayName,
		Email:       in.Email,
		Role:        role,
		Status:      domain.UserStatusActive,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := user.Validate(); err != nil {
		return nil, err
	}
	if err := s.deps.repos.Users.Create(ctx, user); err != nil {
		return nil, err
	}

	if err := recordAudit(
		ctx,
		s.deps.repos,
		s.deps.idg,
		now,
		in.Meta,
		domain.AuditResourceUser,
		string(user.ID),
		domain.AuditActionCreate,
		fmt.Sprintf("create user %s", user.Username),
	); err != nil {
		rollbackErr := s.deps.repos.Users.Delete(ctx, user.ID)
		return nil, joinAuditAndRollbackError(err, "rollback user create", rollbackErr)
	}

	return user, nil
}

func (s *userService) GetByID(ctx context.Context, id domain.UserID) (*domain.User, error) {
	return s.deps.repos.Users.GetByID(ctx, id)
}

func (s *userService) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	return s.deps.repos.Users.GetByUsername(ctx, username)
}

func (s *userService) List(ctx context.Context) ([]*domain.User, error) {
	return s.deps.repos.Users.List(ctx)
}

func (s *userService) Suspend(ctx context.Context, in UserActionInput) (*domain.User, error) {
	user, err := s.deps.repos.Users.GetByID(ctx, in.UserID)
	if err != nil {
		return nil, err
	}

	now := s.deps.now()
	if err := user.Suspend(now); err != nil {
		return nil, err
	}
	if err := s.deps.repos.Users.Update(ctx, user); err != nil {
		return nil, err
	}

	if err := recordAudit(
		ctx,
		s.deps.repos,
		s.deps.idg,
		now,
		in.Meta,
		domain.AuditResourceUser,
		string(user.ID),
		domain.AuditActionUpdate,
		"suspend user",
	); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *userService) Activate(ctx context.Context, in UserActionInput) (*domain.User, error) {
	user, err := s.deps.repos.Users.GetByID(ctx, in.UserID)
	if err != nil {
		return nil, err
	}

	now := s.deps.now()
	if err := user.Activate(now); err != nil {
		return nil, err
	}
	if err := s.deps.repos.Users.Update(ctx, user); err != nil {
		return nil, err
	}

	if err := recordAudit(
		ctx,
		s.deps.repos,
		s.deps.idg,
		now,
		in.Meta,
		domain.AuditResourceUser,
		string(user.ID),
		domain.AuditActionUpdate,
		"activate user",
	); err != nil {
		return nil, err
	}

	return user, nil
}
