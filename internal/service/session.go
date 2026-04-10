package service

import (
	"context"
	"strings"

	"HaruhiServer/internal/domain"
)

type sessionService struct {
	deps deps
}

var _ SessionService = (*sessionService)(nil)

func (s *sessionService) Create(ctx context.Context, in CreateSessionInput) (*domain.Session, error) {
	if in.TTL <= 0 {
		return nil, domain.NewDomainError(domain.ErrInvalidArgument, "session.ttl must be positive")
	}
	if _, err := s.deps.repos.Users.GetByID(ctx, in.UserID); err != nil {
		return nil, err
	}

	now := s.deps.now()

	session := &domain.Session{
		ID:        domain.SessionID(s.deps.idg.NewID()),
		UserID:    in.UserID,
		TokenHash: strings.TrimSpace(in.TokenHash),
		Status:    domain.SessionStatusActive,
		CreatedAt: now,
		ExpiresAt: now.Add(in.TTL),
	}

	if err := session.Validate(); err != nil {
		return nil, err
	}
	if err := s.deps.repos.Sessions.Create(ctx, session); err != nil {
		return nil, err
	}

	if err := recordAudit(
		ctx,
		s.deps.repos,
		s.deps.idg,
		now,
		in.Meta,
		domain.AuditResourceSession,
		string(session.ID),
		domain.AuditActionLogin,
		"create session",
	); err != nil {
		rollbackErr := s.deps.repos.Sessions.Delete(ctx, session.ID)
		return nil, joinAuditAndRollbackError(err, "rollback session create", rollbackErr)
	}

	return session, nil
}

func (s *sessionService) GetByID(ctx context.Context, id domain.SessionID) (*domain.Session, error) {
	return s.deps.repos.Sessions.GetByID(ctx, id)
}

func (s *sessionService) GetByTokenHash(ctx context.Context, tokenHash string) (*domain.Session, error) {
	return s.deps.repos.Sessions.GetByTokenHash(ctx, strings.TrimSpace(tokenHash))
}

func (s *sessionService) List(ctx context.Context) ([]*domain.Session, error) {
	return s.deps.repos.Sessions.List(ctx)
}

func (s *sessionService) ListByUserID(ctx context.Context, userID domain.UserID) ([]*domain.Session, error) {
	return s.deps.repos.Sessions.ListByUserID(ctx, userID)
}

func (s *sessionService) Revoke(ctx context.Context, in SessionActionInput) (*domain.Session, error) {
	session, err := s.deps.repos.Sessions.GetByID(ctx, in.SessionID)
	if err != nil {
		return nil, err
	}

	now := s.deps.now()
	if err := session.Revoke(now); err != nil {
		return nil, err
	}
	if err := s.deps.repos.Sessions.Update(ctx, session); err != nil {
		return nil, err
	}

	if err := recordAudit(
		ctx,
		s.deps.repos,
		s.deps.idg,
		now,
		in.Meta,
		domain.AuditResourceSession,
		string(session.ID),
		domain.AuditActionLogout,
		"revoke session",
	); err != nil {
		return nil, err
	}

	return session, nil
}

func (s *sessionService) Touch(ctx context.Context, in SessionTouchInput) (*domain.Session, error) {
	session, err := s.deps.repos.Sessions.GetByID(ctx, in.SessionID)
	if err != nil {
		return nil, err
	}

	now := s.deps.now()
	if err := session.Touch(now); err != nil {
		return nil, err
	}
	if err := s.deps.repos.Sessions.Update(ctx, session); err != nil {
		if de := domain.AsDomainError(err); de != nil && de.Code == domain.ErrInvalidState {
			latest, latestErr := s.deps.repos.Sessions.GetByID(ctx, in.SessionID)
			if latestErr == nil && latest.Status == domain.SessionStatusRevoked {
				return latest, nil
			}
		}
		return nil, err
	}

	return session, nil
}
