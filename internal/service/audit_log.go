package service

import (
	"context"
	"strings"

	"HaruhiServer/internal/domain"
)

type auditLogService struct {
	deps deps
}

var _ AuditLogService = (*auditLogService)(nil)

func (s *auditLogService) Append(ctx context.Context, in AppendAuditLogInput) (*domain.AuditLog, error) {
	now := s.deps.now()

	log := &domain.AuditLog{
		ID:           domain.AuditLogID(s.deps.idg.NewID()),
		ActorID:      cloneUserIDPtr(in.ActorID),
		ResourceType: in.ResourceType,
		ResourceID:   strings.TrimSpace(in.ResourceID),
		Action:       in.Action,
		RequestID:    normalizeRequestID(in.RequestID),
		Detail:       in.Detail,
		CreatedAt:    now,
	}

	if err := log.Validate(); err != nil {
		return nil, err
	}
	if err := s.deps.repos.AuditLogs.Create(ctx, log); err != nil {
		return nil, err
	}

	return log, nil
}

func (s *auditLogService) GetByID(ctx context.Context, id domain.AuditLogID) (*domain.AuditLog, error) {
	return s.deps.repos.AuditLogs.GetByID(ctx, id)
}

func (s *auditLogService) List(ctx context.Context) ([]*domain.AuditLog, error) {
	return s.deps.repos.AuditLogs.List(ctx)
}

func (s *auditLogService) ListByActorID(ctx context.Context, actorID domain.UserID) ([]*domain.AuditLog, error) {
	return s.deps.repos.AuditLogs.ListByActorID(ctx, actorID)
}

func (s *auditLogService) ListByResource(
	ctx context.Context,
	resourceType domain.AuditResourceType,
	resourceID string,
) ([]*domain.AuditLog, error) {
	return s.deps.repos.AuditLogs.ListByResource(ctx, resourceType, strings.TrimSpace(resourceID))
}
