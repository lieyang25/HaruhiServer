package domain

import "time"

type UserRole string

const (
	UserRoleMember UserRole = "member"
	UserRoleAdmin  UserRole = "admin"
)

func (r UserRole) IsValid() bool {
	switch r {
	case UserRoleMember, UserRoleAdmin:
		return true
	default:
		return false
	}
}

type UserStatus string

const (
	UserStatusActive    UserStatus = "active"
	UserStatusSuspended UserStatus = "suspended"
)

func (s UserStatus) IsValid() bool {
	switch s {
	case UserStatusActive, UserStatusSuspended:
		return true
	default:
		return false
	}
}

type User struct {
	ID          UserID
	Username    string
	DisplayName string
	Email       string
	Role        UserRole
	Status      UserStatus
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (u *User) Validate() error {
	if err := requireNonEmpty("user.id", string(u.ID)); err != nil {
		return err
	}
	if err := requireNonEmpty("user.username", u.Username); err != nil {
		return err
	}
	if err := requireMaxRunes("user.username", u.Username, 32); err != nil {
		return err
	}
	if err := requireNonEmpty("user.display_name", u.DisplayName); err != nil {
		return err
	}
	if err := requireMaxRunes("user.display_name", u.DisplayName, 64); err != nil {
		return err
	}
	if err := requireNonEmpty("user.email", u.Email); err != nil {
		return err
	}
	if !u.Role.IsValid() {
		return NewDomainError(ErrInvalidArgument, "user.role is invalid")
	}
	if !u.Status.IsValid() {
		return NewDomainError(ErrInvalidArgument, "user.status is invalid")
	}
	if u.CreatedAt.IsZero() {
		return NewDomainError(ErrInvalidArgument, "user.created_at is required")
	}
	if u.UpdatedAt.IsZero() {
		return NewDomainError(ErrInvalidArgument, "user.updated_at is required")
	}
	return nil
}

func (u *User) Suspend(now time.Time) error {
	if u.Status == UserStatusSuspended {
		return NewDomainError(ErrInvalidState, "user is already suspended")
	}
	u.Status = UserStatusSuspended
	u.UpdatedAt = now
	return nil
}

func (u *User) Activate(now time.Time) error {
	if u.Status == UserStatusActive {
		return NewDomainError(ErrInvalidState, "user is already active")
	}
	u.Status = UserStatusActive
	u.UpdatedAt = now
	return nil
}
