package service_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"HaruhiServer/internal/domain"
	"HaruhiServer/internal/repository"
	"HaruhiServer/internal/repository/memory"
	"HaruhiServer/internal/service"
)

type sequenceIDGenerator struct {
	mu    sync.Mutex
	next  int
	label string
}

func newSequenceIDGenerator(label string) *sequenceIDGenerator {
	if label == "" {
		label = "id"
	}
	return &sequenceIDGenerator{label: label}
}

func (g *sequenceIDGenerator) NewID() string {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.next++
	return fmt.Sprintf("%s-%03d", g.label, g.next)
}

func fixedNow(t time.Time) service.NowFunc {
	return func() time.Time { return t }
}

func newMemoryRepos() *repository.Repositories {
	return memory.NewRepositories()
}

type failingAuditLogRepository struct {
	createErr error
}

func (r *failingAuditLogRepository) Create(context.Context, *domain.AuditLog) error {
	if r.createErr == nil {
		return errors.New("forced audit log create error")
	}
	return r.createErr
}

func (r *failingAuditLogRepository) GetByID(context.Context, domain.AuditLogID) (*domain.AuditLog, error) {
	return nil, domain.NewDomainError(domain.ErrNotFound, "audit log not found")
}

func (r *failingAuditLogRepository) List(context.Context) ([]*domain.AuditLog, error) {
	return nil, nil
}

func (r *failingAuditLogRepository) ListByActorID(context.Context, domain.UserID) ([]*domain.AuditLog, error) {
	return nil, nil
}

func (r *failingAuditLogRepository) ListByResource(context.Context, domain.AuditResourceType, string) ([]*domain.AuditLog, error) {
	return nil, nil
}

type scriptedSessionRepository struct {
	inner       repository.SessionRepository
	onGetByID   func()
	startUpdate <-chan struct{}
	revokeDone  chan struct{}
}

func (r *scriptedSessionRepository) Create(ctx context.Context, session *domain.Session) error {
	return r.inner.Create(ctx, session)
}

func (r *scriptedSessionRepository) Update(ctx context.Context, session *domain.Session) error {
	if r.startUpdate != nil {
		<-r.startUpdate
	}

	if session.Status == domain.SessionStatusActive && session.LastSeenAt != nil && r.revokeDone != nil {
		<-r.revokeDone
		return r.inner.Update(ctx, session)
	}

	err := r.inner.Update(ctx, session)
	if err == nil && session.Status == domain.SessionStatusRevoked && r.revokeDone != nil {
		select {
		case <-r.revokeDone:
		default:
			close(r.revokeDone)
		}
	}
	return err
}

func (r *scriptedSessionRepository) GetByID(ctx context.Context, id domain.SessionID) (*domain.Session, error) {
	session, err := r.inner.GetByID(ctx, id)
	if err == nil && r.onGetByID != nil {
		r.onGetByID()
	}
	return session, err
}

func (r *scriptedSessionRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*domain.Session, error) {
	return r.inner.GetByTokenHash(ctx, tokenHash)
}

func (r *scriptedSessionRepository) Delete(ctx context.Context, id domain.SessionID) error {
	return r.inner.Delete(ctx, id)
}

func (r *scriptedSessionRepository) List(ctx context.Context) ([]*domain.Session, error) {
	return r.inner.List(ctx)
}

func (r *scriptedSessionRepository) ListByUserID(ctx context.Context, userID domain.UserID) ([]*domain.Session, error) {
	return r.inner.ListByUserID(ctx, userID)
}

type scriptedTaskRepository struct {
	inner       repository.TaskRepository
	onGetByID   func()
	startUpdate <-chan struct{}
	doneDone    chan struct{}
}

func (r *scriptedTaskRepository) Create(ctx context.Context, task *domain.Task) error {
	return r.inner.Create(ctx, task)
}

func (r *scriptedTaskRepository) Update(ctx context.Context, task *domain.Task) error {
	if r.startUpdate != nil {
		<-r.startUpdate
	}

	if task.Status == domain.TaskStatusCanceled && r.doneDone != nil {
		<-r.doneDone
		return r.inner.Update(ctx, task)
	}

	err := r.inner.Update(ctx, task)
	if err == nil && task.Status == domain.TaskStatusDone && r.doneDone != nil {
		select {
		case <-r.doneDone:
		default:
			close(r.doneDone)
		}
	}
	return err
}

func (r *scriptedTaskRepository) GetByID(ctx context.Context, id domain.TaskID) (*domain.Task, error) {
	task, err := r.inner.GetByID(ctx, id)
	if err == nil && r.onGetByID != nil {
		r.onGetByID()
	}
	return task, err
}

func (r *scriptedTaskRepository) Delete(ctx context.Context, id domain.TaskID) error {
	return r.inner.Delete(ctx, id)
}

func (r *scriptedTaskRepository) List(ctx context.Context) ([]*domain.Task, error) {
	return r.inner.List(ctx)
}

func (r *scriptedTaskRepository) ListByProjectID(ctx context.Context, projectID domain.ProjectID) ([]*domain.Task, error) {
	return r.inner.ListByProjectID(ctx, projectID)
}

func (r *scriptedTaskRepository) ListByAssigneeID(ctx context.Context, assigneeID domain.UserID) ([]*domain.Task, error) {
	return r.inner.ListByAssigneeID(ctx, assigneeID)
}
