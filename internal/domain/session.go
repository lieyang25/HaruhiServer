package domain

import "time"

type SessionStatus string

const (
	SessionStatusActive  SessionStatus = "active"
	SessionStatusRevoked SessionStatus = "revoked"
)

func (s SessionStatus) IsValid() bool {
	switch s {
	case SessionStatusActive, SessionStatusRevoked:
		return true
	default:
		return false
	}
}

type Session struct {
	ID         SessionID
	UserID     UserID
	TokenHash  string
	Status     SessionStatus
	CreatedAt  time.Time
	ExpiresAt  time.Time
	LastSeenAt *time.Time
	RevokedAt  *time.Time
}

func (s *Session) Validate() error {
	if err := requireNonEmpty("session.id", string(s.ID)); err != nil {
		return err
	}
	if err := requireNonEmpty("session.user_id", string(s.UserID)); err != nil {
		return err
	}
	if err := requireNonEmpty("session.token_hash", s.TokenHash); err != nil {
		return err
	}
	if !s.Status.IsValid() {
		return NewDomainError(ErrInvalidArgument, "session.status is invalid")
	}
	if s.CreatedAt.IsZero() {
		return NewDomainError(ErrInvalidArgument, "session.created_at is required")
	}
	if s.ExpiresAt.IsZero() {
		return NewDomainError(ErrInvalidArgument, "session.expires_at is required")
	}
	if !s.ExpiresAt.After(s.CreatedAt) {
		return NewDomainError(ErrInvalidArgument, "session.expires_at must be after created_at")
	}
	if s.Status == SessionStatusRevoked && s.RevokedAt == nil {
		return NewDomainError(ErrInvalidState, "session.revoked_at is required when status is revoked")
	}
	if s.Status != SessionStatusRevoked && s.RevokedAt != nil {
		return NewDomainError(ErrInvalidState, "session.revoked_at must be nil when status is not revoked")
	}
	return nil
}

func (s *Session) IsExpired(now time.Time) bool {
	return !now.Before(s.ExpiresAt)
}

func (s *Session) Revoke(now time.Time) error {
	if s.Status == SessionStatusRevoked {
		return NewDomainError(ErrInvalidState, "session is already revoked")
	}
	s.Status = SessionStatusRevoked
	s.RevokedAt = &now
	return nil
}

func (s *Session) Touch(now time.Time) error {
	if s.Status == SessionStatusRevoked {
		return NewDomainError(ErrInvalidState, "revoked session cannot be touched")
	}
	if s.IsExpired(now) {
		return NewDomainError(ErrForbidden, "expired session cannot be touched")
	}
	s.LastSeenAt = &now
	return nil
}
