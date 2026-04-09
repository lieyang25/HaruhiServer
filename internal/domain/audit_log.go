package domain

import "time"

type AuditResourceType string

const (
	AuditResourceUser    AuditResourceType = "user"
	AuditResourceProject AuditResourceType = "project"
	AuditResourceTask    AuditResourceType = "task"
	AuditResourceNote    AuditResourceType = "note"
	AuditResourceSession AuditResourceType = "session"
)

func (t AuditResourceType) IsValid() bool {
	switch t {
	case AuditResourceUser, AuditResourceProject, AuditResourceTask, AuditResourceNote, AuditResourceSession:
		return true
	default:
		return false
	}
}

type AuditAction string

const (
	AuditActionCreate  AuditAction = "create"
	AuditActionUpdate  AuditAction = "update"
	AuditActionDelete  AuditAction = "delete"
	AuditActionArchive AuditAction = "archive"
	AuditActionRestore AuditAction = "restore"
	AuditActionLogin   AuditAction = "login"
	AuditActionLogout  AuditAction = "logout"
)

func (a AuditAction) IsValid() bool {
	switch a {
	case AuditActionCreate, AuditActionUpdate, AuditActionDelete, AuditActionArchive, AuditActionRestore, AuditActionLogin, AuditActionLogout:
		return true
	default:
		return false
	}
}

type AuditLog struct {
	ID           AuditLogID
	ActorID      *UserID
	ResourceType AuditResourceType
	ResourceID   string
	Action       AuditAction
	RequestID    string
	Detail       string
	CreatedAt    time.Time
}

func (a *AuditLog) Validate() error {
	if err := requireNonEmpty("audit_log.id", string(a.ID)); err != nil {
		return err
	}
	if !a.ResourceType.IsValid() {
		return NewDomainError(ErrInvalidArgument, "audit_log.resource_type is invalid")
	}
	if err := requireNonEmpty("audit_log.resource_id", a.ResourceID); err != nil {
		return err
	}
	if !a.Action.IsValid() {
		return NewDomainError(ErrInvalidArgument, "audit_log.action is invalid")
	}
	if a.CreatedAt.IsZero() {
		return NewDomainError(ErrInvalidArgument, "audit_log.created_at is required")
	}
	if err := requireMaxRunes("audit_log.detail", a.Detail, 2048); err != nil {
		return err
	}
	return nil
}
