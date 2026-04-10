package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"HaruhiServer/internal/domain"
	"HaruhiServer/internal/repository"
)

type NowFunc func() time.Time

type IDGenerator interface {
	NewID() string
}

type RandomIDGenerator struct{}

func (RandomIDGenerator) NewID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err == nil {
		return hex.EncodeToString(b[:])
	}
	return fmt.Sprintf("id-%d", time.Now().UnixNano())
}

type ActionMeta struct {
	ActorID   *domain.UserID
	RequestID string
}

type deps struct {
	repos *repository.Repositories
	idg   IDGenerator
	now   NowFunc
}

func newDeps(repos *repository.Repositories, idg IDGenerator, now NowFunc) deps {
	if idg == nil {
		idg = RandomIDGenerator{}
	}
	if now == nil {
		now = time.Now
	}
	return deps{
		repos: repos,
		idg:   idg,
		now:   now,
	}
}

func cloneUserIDPtr(v *domain.UserID) *domain.UserID {
	if v == nil {
		return nil
	}
	cp := *v
	return &cp
}

func cloneTimePtr(v *time.Time) *time.Time {
	if v == nil {
		return nil
	}
	cp := *v
	return &cp
}

func normalizeRequestID(s string) string {
	return strings.TrimSpace(s)
}

func recordAudit(
	ctx context.Context,
	repos *repository.Repositories,
	idg IDGenerator,
	at time.Time,
	meta ActionMeta,
	resourceType domain.AuditResourceType,
	resourceID string,
	action domain.AuditAction,
	detail string,
) error {
	if repos == nil || repos.AuditLogs == nil {
		return nil
	}

	log := &domain.AuditLog{
		ID:           domain.AuditLogID(idg.NewID()),
		ActorID:      cloneUserIDPtr(meta.ActorID),
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Action:       action,
		RequestID:    normalizeRequestID(meta.RequestID),
		Detail:       detail,
		CreatedAt:    at,
	}

	return repos.AuditLogs.Create(ctx, log)
}

func joinAuditAndRollbackError(auditErr error, rollbackStep string, rollbackErr error) error {
	if rollbackErr == nil {
		return auditErr
	}

	return errors.Join(
		auditErr,
		fmt.Errorf("%s: %w", rollbackStep, rollbackErr),
	)
}
