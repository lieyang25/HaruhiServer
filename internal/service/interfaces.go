package service

import (
	"context"
	"time"

	"HaruhiServer/internal/domain"
)

type CreateUserInput struct {
	Username    string
	DisplayName string
	Email       string
	Role        domain.UserRole
	Meta        ActionMeta
}

type UserActionInput struct {
	UserID domain.UserID
	Meta   ActionMeta
}

type UserService interface {
	Create(ctx context.Context, in CreateUserInput) (*domain.User, error)
	GetByID(ctx context.Context, id domain.UserID) (*domain.User, error)
	GetByUsername(ctx context.Context, username string) (*domain.User, error)
	List(ctx context.Context) ([]*domain.User, error)
	Suspend(ctx context.Context, in UserActionInput) (*domain.User, error)
	Activate(ctx context.Context, in UserActionInput) (*domain.User, error)
}

type CreateProjectInput struct {
	OwnerID     domain.UserID
	Name        string
	Description string
	Visibility  domain.ProjectVisibility
	Meta        ActionMeta
}

type ProjectActionInput struct {
	ProjectID domain.ProjectID
	Meta      ActionMeta
}

type ProjectService interface {
	Create(ctx context.Context, in CreateProjectInput) (*domain.Project, error)
	GetByID(ctx context.Context, id domain.ProjectID) (*domain.Project, error)
	List(ctx context.Context) ([]*domain.Project, error)
	ListByOwnerID(ctx context.Context, ownerID domain.UserID) ([]*domain.Project, error)
	Update(ctx context.Context, in UpdateProjectInput) (*domain.Project, error)
	Delete(ctx context.Context, in ProjectActionInput) error
	Archive(ctx context.Context, in ProjectActionInput) (*domain.Project, error)
	Reopen(ctx context.Context, in ProjectActionInput) (*domain.Project, error)
}

type UpdateProjectInput struct {
	ProjectID   domain.ProjectID
	Name        *string
	Description *string
	Visibility  *domain.ProjectVisibility
	Archive     *bool
	Meta        ActionMeta
}

type CreateTaskInput struct {
	ProjectID   domain.ProjectID
	CreatorID   domain.UserID
	AssigneeID  *domain.UserID
	Title       string
	Description string
	Priority    domain.TaskPriority
	DueAt       *time.Time
	Meta        ActionMeta
}

type TaskActionInput struct {
	TaskID domain.TaskID
	Meta   ActionMeta
}

type TaskService interface {
	Create(ctx context.Context, in CreateTaskInput) (*domain.Task, error)
	GetByID(ctx context.Context, id domain.TaskID) (*domain.Task, error)
	List(ctx context.Context) ([]*domain.Task, error)
	ListByProjectID(ctx context.Context, projectID domain.ProjectID) ([]*domain.Task, error)
	ListByAssigneeID(ctx context.Context, assigneeID domain.UserID) ([]*domain.Task, error)
	Update(ctx context.Context, in UpdateTaskInput) (*domain.Task, error)
	Delete(ctx context.Context, in TaskActionInput) error
	Start(ctx context.Context, in TaskActionInput) (*domain.Task, error)
	MarkDone(ctx context.Context, in TaskActionInput) (*domain.Task, error)
	Cancel(ctx context.Context, in TaskActionInput) (*domain.Task, error)
	Reopen(ctx context.Context, in TaskActionInput) (*domain.Task, error)
}

type UpdateTaskInput struct {
	TaskID        domain.TaskID
	Title         *string
	Description   *string
	AssigneeID    *domain.UserID
	ClearAssignee bool
	Priority      *domain.TaskPriority
	DueAt         *time.Time
	ClearDueAt    bool
	Meta          ActionMeta
}

type CreateNoteInput struct {
	ProjectID domain.ProjectID
	AuthorID  domain.UserID
	Title     string
	Body      string
	Format    domain.NoteFormat
	Meta      ActionMeta
}

type NoteActionInput struct {
	NoteID domain.NoteID
	Meta   ActionMeta
}

type NoteService interface {
	Create(ctx context.Context, in CreateNoteInput) (*domain.Note, error)
	GetByID(ctx context.Context, id domain.NoteID) (*domain.Note, error)
	List(ctx context.Context) ([]*domain.Note, error)
	ListByProjectID(ctx context.Context, projectID domain.ProjectID) ([]*domain.Note, error)
	Archive(ctx context.Context, in NoteActionInput) (*domain.Note, error)
	Restore(ctx context.Context, in NoteActionInput) (*domain.Note, error)
}

type CreateSessionInput struct {
	UserID    domain.UserID
	TokenHash string
	TTL       time.Duration
	Meta      ActionMeta
}

type SessionActionInput struct {
	SessionID domain.SessionID
	Meta      ActionMeta
}

type SessionTouchInput struct {
	SessionID domain.SessionID
	Meta      ActionMeta
}

type SessionService interface {
	Create(ctx context.Context, in CreateSessionInput) (*domain.Session, error)
	GetByID(ctx context.Context, id domain.SessionID) (*domain.Session, error)
	GetByTokenHash(ctx context.Context, tokenHash string) (*domain.Session, error)
	List(ctx context.Context) ([]*domain.Session, error)
	ListByUserID(ctx context.Context, userID domain.UserID) ([]*domain.Session, error)
	Revoke(ctx context.Context, in SessionActionInput) (*domain.Session, error)
	Touch(ctx context.Context, in SessionTouchInput) (*domain.Session, error)
}

type AppendAuditLogInput struct {
	ActorID      *domain.UserID
	ResourceType domain.AuditResourceType
	ResourceID   string
	Action       domain.AuditAction
	RequestID    string
	Detail       string
}

type AuditLogService interface {
	Append(ctx context.Context, in AppendAuditLogInput) (*domain.AuditLog, error)
	GetByID(ctx context.Context, id domain.AuditLogID) (*domain.AuditLog, error)
	List(ctx context.Context) ([]*domain.AuditLog, error)
	ListByActorID(ctx context.Context, actorID domain.UserID) ([]*domain.AuditLog, error)
	ListByResource(ctx context.Context, resourceType domain.AuditResourceType, resourceID string) ([]*domain.AuditLog, error)
}
