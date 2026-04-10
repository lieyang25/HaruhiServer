package repository

import (
	"context"

	"HaruhiServer/internal/domain"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	Update(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id domain.UserID) (*domain.User, error)
	GetByUsername(ctx context.Context, username string) (*domain.User, error)
	Delete(ctx context.Context, id domain.UserID) error
	List(ctx context.Context) ([]*domain.User, error)
}

type ProjectRepository interface {
	Create(ctx context.Context, project *domain.Project) error
	Update(ctx context.Context, project *domain.Project) error
	GetByID(ctx context.Context, id domain.ProjectID) (*domain.Project, error)
	Delete(ctx context.Context, id domain.ProjectID) error
	List(ctx context.Context) ([]*domain.Project, error)
	ListByOwnerID(ctx context.Context, ownerID domain.UserID) ([]*domain.Project, error)
}

type TaskRepository interface {
	Create(ctx context.Context, task *domain.Task) error
	Update(ctx context.Context, task *domain.Task) error
	GetByID(ctx context.Context, id domain.TaskID) (*domain.Task, error)
	Delete(ctx context.Context, id domain.TaskID) error
	List(ctx context.Context) ([]*domain.Task, error)
	ListByProjectID(ctx context.Context, projectID domain.ProjectID) ([]*domain.Task, error)
	ListByAssigneeID(ctx context.Context, assigneeID domain.UserID) ([]*domain.Task, error)
}

type NoteRepository interface {
	Create(ctx context.Context, note *domain.Note) error
	Update(ctx context.Context, note *domain.Note) error
	GetByID(ctx context.Context, id domain.NoteID) (*domain.Note, error)
	Delete(ctx context.Context, id domain.NoteID) error
	List(ctx context.Context) ([]*domain.Note, error)
	ListByProjectID(ctx context.Context, projectID domain.ProjectID) ([]*domain.Note, error)
}

type SessionRepository interface {
	Create(ctx context.Context, session *domain.Session) error
	Update(ctx context.Context, session *domain.Session) error
	GetByID(ctx context.Context, id domain.SessionID) (*domain.Session, error)
	GetByTokenHash(ctx context.Context, tokenHash string) (*domain.Session, error)
	Delete(ctx context.Context, id domain.SessionID) error
	List(ctx context.Context) ([]*domain.Session, error)
	ListByUserID(ctx context.Context, userID domain.UserID) ([]*domain.Session, error)
}

type AuditLogRepository interface {
	Create(ctx context.Context, log *domain.AuditLog) error
	GetByID(ctx context.Context, id domain.AuditLogID) (*domain.AuditLog, error)
	List(ctx context.Context) ([]*domain.AuditLog, error)
	ListByActorID(ctx context.Context, actorID domain.UserID) ([]*domain.AuditLog, error)
	ListByResource(ctx context.Context, resourceType domain.AuditResourceType, resourceID string) ([]*domain.AuditLog, error)
}

type Repositories struct {
	Users     UserRepository
	Projects  ProjectRepository
	Tasks     TaskRepository
	Notes     NoteRepository
	Sessions  SessionRepository
	AuditLogs AuditLogRepository
}
