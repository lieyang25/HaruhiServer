package memory

import (
	"context"
	"sort"
	"strings"
	"time"

	"HaruhiServer/internal/domain"
)

func checkContext(ctx context.Context) error {
	if ctx == nil {
		return nil
	}
	return ctx.Err()
}

func indexKey(s string) string {
	return strings.TrimSpace(s)
}

func cloneTimePtr(v *time.Time) *time.Time {
	if v == nil {
		return nil
	}
	cp := *v
	return &cp
}

func cloneUserIDPtr(v *domain.UserID) *domain.UserID {
	if v == nil {
		return nil
	}
	cp := *v
	return &cp
}

func cloneUser(v *domain.User) *domain.User {
	if v == nil {
		return nil
	}
	cp := *v
	return &cp
}

func cloneProject(v *domain.Project) *domain.Project {
	if v == nil {
		return nil
	}
	cp := *v
	return &cp
}

func cloneTask(v *domain.Task) *domain.Task {
	if v == nil {
		return nil
	}
	cp := *v
	cp.AssigneeID = cloneUserIDPtr(v.AssigneeID)
	cp.DueAt = cloneTimePtr(v.DueAt)
	cp.CompletedAt = cloneTimePtr(v.CompletedAt)
	return &cp
}

func cloneNote(v *domain.Note) *domain.Note {
	if v == nil {
		return nil
	}
	cp := *v
	return &cp
}

func cloneSession(v *domain.Session) *domain.Session {
	if v == nil {
		return nil
	}
	cp := *v
	cp.LastSeenAt = cloneTimePtr(v.LastSeenAt)
	cp.RevokedAt = cloneTimePtr(v.RevokedAt)
	return &cp
}

func cloneAuditLog(v *domain.AuditLog) *domain.AuditLog {
	if v == nil {
		return nil
	}
	cp := *v
	cp.ActorID = cloneUserIDPtr(v.ActorID)
	return &cp
}

func sortUsers(items []*domain.User) {
	sort.Slice(items, func(i, j int) bool {
		if items[i].CreatedAt.Equal(items[j].CreatedAt) {
			return string(items[i].ID) < string(items[j].ID)
		}
		return items[i].CreatedAt.Before(items[j].CreatedAt)
	})
}

func sortProjects(items []*domain.Project) {
	sort.Slice(items, func(i, j int) bool {
		if items[i].CreatedAt.Equal(items[j].CreatedAt) {
			return string(items[i].ID) < string(items[j].ID)
		}
		return items[i].CreatedAt.Before(items[j].CreatedAt)
	})
}

func sortTasks(items []*domain.Task) {
	sort.Slice(items, func(i, j int) bool {
		if items[i].CreatedAt.Equal(items[j].CreatedAt) {
			return string(items[i].ID) < string(items[j].ID)
		}
		return items[i].CreatedAt.Before(items[j].CreatedAt)
	})
}

func sortNotes(items []*domain.Note) {
	sort.Slice(items, func(i, j int) bool {
		if items[i].CreatedAt.Equal(items[j].CreatedAt) {
			return string(items[i].ID) < string(items[j].ID)
		}
		return items[i].CreatedAt.Before(items[j].CreatedAt)
	})
}

func sortSessions(items []*domain.Session) {
	sort.Slice(items, func(i, j int) bool {
		if items[i].CreatedAt.Equal(items[j].CreatedAt) {
			return string(items[i].ID) < string(items[j].ID)
		}
		return items[i].CreatedAt.Before(items[j].CreatedAt)
	})
}

func sortAuditLogs(items []*domain.AuditLog) {
	sort.Slice(items, func(i, j int) bool {
		if items[i].CreatedAt.Equal(items[j].CreatedAt) {
			return string(items[i].ID) < string(items[j].ID)
		}
		return items[i].CreatedAt.Before(items[j].CreatedAt)
	})
}
