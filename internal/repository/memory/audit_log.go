package memory

import (
	"context"
	"sync"

	"HaruhiServer/internal/domain"
)

type AuditLogRepository struct {
	mu   sync.RWMutex
	byID map[domain.AuditLogID]*domain.AuditLog
}

func NewAuditLogRepository() *AuditLogRepository {
	return &AuditLogRepository{
		byID: make(map[domain.AuditLogID]*domain.AuditLog),
	}
}

func (r *AuditLogRepository) Create(ctx context.Context, log *domain.AuditLog) error {
	if err := checkContext(ctx); err != nil {
		return err
	}
	if log == nil {
		return domain.NewDomainError(domain.ErrInvalidArgument, "audit log is required")
	}
	if err := log.Validate(); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.byID[log.ID]; ok {
		return domain.NewDomainError(domain.ErrConflict, "audit log id already exists")
	}

	r.byID[log.ID] = cloneAuditLog(log)
	return nil
}

func (r *AuditLogRepository) GetByID(ctx context.Context, id domain.AuditLogID) (*domain.AuditLog, error) {
	if err := checkContext(ctx); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	log, ok := r.byID[id]
	if !ok {
		return nil, domain.NewDomainError(domain.ErrNotFound, "audit log not found")
	}

	return cloneAuditLog(log), nil
}

func (r *AuditLogRepository) List(ctx context.Context) ([]*domain.AuditLog, error) {
	if err := checkContext(ctx); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	items := make([]*domain.AuditLog, 0, len(r.byID))
	for _, log := range r.byID {
		items = append(items, cloneAuditLog(log))
	}

	sortAuditLogs(items)
	return items, nil
}

func (r *AuditLogRepository) ListByActorID(ctx context.Context, actorID domain.UserID) ([]*domain.AuditLog, error) {
	if err := checkContext(ctx); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	items := make([]*domain.AuditLog, 0)
	for _, log := range r.byID {
		if log.ActorID != nil && *log.ActorID == actorID {
			items = append(items, cloneAuditLog(log))
		}
	}

	sortAuditLogs(items)
	return items, nil
}

func (r *AuditLogRepository) ListByResource(ctx context.Context, resourceType domain.AuditResourceType, resourceID string) ([]*domain.AuditLog, error) {
	if err := checkContext(ctx); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	items := make([]*domain.AuditLog, 0)
	for _, log := range r.byID {
		if log.ResourceType == resourceType && log.ResourceID == resourceID {
			items = append(items, cloneAuditLog(log))
		}
	}

	sortAuditLogs(items)
	return items, nil
}
