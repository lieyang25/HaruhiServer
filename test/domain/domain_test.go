package domain_test

import (
	"errors"
	"testing"
	"time"

	"HaruhiServer/internal/domain"
)

func validUser() *domain.User {
	now := time.Unix(1710000000, 0)
	return &domain.User{
		ID:          domain.UserID("u-1"),
		Username:    "haruhi",
		DisplayName: "Haruhi Suzumiya",
		Email:       "haruhi@example.com",
		Role:        domain.UserRoleMember,
		Status:      domain.UserStatusActive,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func validProject() *domain.Project {
	now := time.Unix(1710000000, 0)
	return &domain.Project{
		ID:          domain.ProjectID("p-1"),
		OwnerID:     domain.UserID("u-1"),
		Name:        "Alpha",
		Description: "Project Alpha",
		Visibility:  domain.ProjectVisibilityPrivate,
		Status:      domain.ProjectStatusActive,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func validTask() *domain.Task {
	now := time.Unix(1710000000, 0)
	due := now.Add(24 * time.Hour)
	return &domain.Task{
		ID:          domain.TaskID("t-1"),
		ProjectID:   domain.ProjectID("p-1"),
		CreatorID:   domain.UserID("u-1"),
		Title:       "Task 1",
		Description: "Do something",
		Status:      domain.TaskStatusTodo,
		Priority:    domain.TaskPriorityMedium,
		DueAt:       &due,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func validNote() *domain.Note {
	now := time.Unix(1710000000, 0)
	return &domain.Note{
		ID:        domain.NoteID("n-1"),
		ProjectID: domain.ProjectID("p-1"),
		AuthorID:  domain.UserID("u-1"),
		Title:     "Note",
		Body:      "content",
		Format:    domain.NoteFormatMarkdown,
		Status:    domain.NoteStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func validSession() *domain.Session {
	now := time.Unix(1710000000, 0)
	return &domain.Session{
		ID:        domain.SessionID("s-1"),
		UserID:    domain.UserID("u-1"),
		TokenHash: "hash-1",
		Status:    domain.SessionStatusActive,
		CreatedAt: now,
		ExpiresAt: now.Add(time.Hour),
	}
}

func validAuditLog() *domain.AuditLog {
	now := time.Unix(1710000000, 0)
	actor := domain.UserID("u-1")
	return &domain.AuditLog{
		ID:           domain.AuditLogID("a-1"),
		ActorID:      &actor,
		ResourceType: domain.AuditResourceTask,
		ResourceID:   "t-1",
		Action:       domain.AuditActionUpdate,
		RequestID:    "req-1",
		Detail:       "update task",
		CreatedAt:    now,
	}
}

func TestDomain_User(t *testing.T) {
	u := validUser()
	if err := u.Validate(); err != nil {
		t.Fatalf("user Validate() err = %v", err)
	}
	u2 := validUser()
	u2.Role = domain.UserRole("owner")
	if err := u2.Validate(); err == nil {
		t.Fatal("user Validate() err = nil, want invalid role")
	}

	t1 := u.UpdatedAt.Add(time.Minute)
	if err := u.Suspend(t1); err != nil {
		t.Fatalf("Suspend() err = %v", err)
	}
	if err := u.Suspend(t1.Add(time.Minute)); err == nil {
		t.Fatal("second Suspend() err = nil")
	}
	if err := u.Activate(t1.Add(2 * time.Minute)); err != nil {
		t.Fatalf("Activate() err = %v", err)
	}
}

func TestDomain_Project(t *testing.T) {
	p := validProject()
	if err := p.Validate(); err != nil {
		t.Fatalf("project Validate() err = %v", err)
	}
	p2 := validProject()
	p2.Visibility = domain.ProjectVisibility("org")
	if err := p2.Validate(); err == nil {
		t.Fatal("project Validate() err = nil, want invalid visibility")
	}

	t1 := p.UpdatedAt.Add(time.Minute)
	if err := p.Archive(t1); err != nil {
		t.Fatalf("Archive() err = %v", err)
	}
	if err := p.Archive(t1.Add(time.Minute)); err == nil {
		t.Fatal("second Archive() err = nil")
	}
	if err := p.Reopen(t1.Add(2 * time.Minute)); err != nil {
		t.Fatalf("Reopen() err = %v", err)
	}
}

func TestDomain_Task(t *testing.T) {
	tk := validTask()
	if err := tk.Validate(); err != nil {
		t.Fatalf("task Validate() err = %v", err)
	}
	badDue := validTask()
	early := badDue.CreatedAt.Add(-time.Second)
	badDue.DueAt = &early
	if err := badDue.Validate(); err == nil {
		t.Fatal("task Validate() err = nil, want due validation")
	}

	now := tk.UpdatedAt.Add(time.Minute)
	if err := tk.Start(now); err != nil {
		t.Fatalf("Start() err = %v", err)
	}
	if err := tk.MarkDone(now.Add(time.Minute)); err != nil {
		t.Fatalf("MarkDone() err = %v", err)
	}
	if err := tk.Cancel(now.Add(2 * time.Minute)); err == nil {
		t.Fatal("Cancel() on done err = nil")
	}
	if err := tk.Reopen(now.Add(3 * time.Minute)); err != nil {
		t.Fatalf("Reopen() err = %v", err)
	}
}

func TestDomain_Note(t *testing.T) {
	n := validNote()
	if err := n.Validate(); err != nil {
		t.Fatalf("note Validate() err = %v", err)
	}
	n2 := validNote()
	n2.Format = domain.NoteFormat("html")
	if err := n2.Validate(); err == nil {
		t.Fatal("note Validate() err = nil, want invalid format")
	}

	t1 := n.UpdatedAt.Add(time.Minute)
	if err := n.Archive(t1); err != nil {
		t.Fatalf("Archive() err = %v", err)
	}
	if err := n.Restore(t1.Add(time.Minute)); err != nil {
		t.Fatalf("Restore() err = %v", err)
	}
}

func TestDomain_Session(t *testing.T) {
	s := validSession()
	if err := s.Validate(); err != nil {
		t.Fatalf("session Validate() err = %v", err)
	}
	if !s.IsExpired(s.ExpiresAt) {
		t.Fatal("IsExpired(at expires_at) = false")
	}
	if err := s.Revoke(s.CreatedAt.Add(time.Minute)); err != nil {
		t.Fatalf("Revoke() err = %v", err)
	}
	if err := s.Touch(s.CreatedAt.Add(2 * time.Minute)); err == nil {
		t.Fatal("Touch() on revoked err = nil")
	}
}

func TestDomain_AuditLogAndErrors(t *testing.T) {
	log := validAuditLog()
	if err := log.Validate(); err != nil {
		t.Fatalf("audit log Validate() err = %v", err)
	}
	bad := validAuditLog()
	bad.Action = domain.AuditAction("noop")
	if err := bad.Validate(); err == nil {
		t.Fatal("audit log Validate() err = nil, want invalid action")
	}

	base := errors.New("boom")
	err := domain.WrapDomainError(domain.ErrConflict, "conflict", base)
	de := domain.AsDomainError(err)
	if de == nil || de.Code != domain.ErrConflict {
		t.Fatalf("AsDomainError(err) = %#v", de)
	}
	if !errors.Is(err, base) {
		t.Fatal("errors.Is(err, base) = false")
	}
}

func TestDomain_DomainErrorString(t *testing.T) {
	var nilErr *domain.DomainError
	if got := nilErr.Error(); got != "<nil>" {
		t.Fatalf("nil DomainError.Error() = %q, want <nil>", got)
	}

	e1 := &domain.DomainError{Code: domain.ErrInvalidArgument, Message: "bad input"}
	if got := e1.Error(); got != "INVALID_ARGUMENT: bad input" {
		t.Fatalf("Error() with message = %q", got)
	}

	e2 := &domain.DomainError{Code: domain.ErrNotFound}
	if got := e2.Error(); got != "NOT_FOUND" {
		t.Fatalf("Error() without message = %q", got)
	}
}
