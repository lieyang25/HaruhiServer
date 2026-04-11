package service_test

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"HaruhiServer/internal/domain"
	"HaruhiServer/internal/service"
)

func TestNewServicesValidation(t *testing.T) {
	if _, err := service.NewServices(nil, nil, nil); err == nil {
		t.Fatal("NewServices(nil) err = nil")
	}
}

func TestService_UserCreateAndAudit(t *testing.T) {
	repos := newMemoryRepos()
	now := time.Unix(1710000000, 0)
	svcs, err := service.NewServices(repos, newSequenceIDGenerator("id"), fixedNow(now))
	if err != nil {
		t.Fatalf("NewServices err = %v", err)
	}

	u, err := svcs.Users.Create(context.Background(), service.CreateUserInput{
		Username:    "haruhi",
		DisplayName: "Haruhi",
		Email:       "haruhi@example.com",
	})
	if err != nil {
		t.Fatalf("Users.Create err = %v", err)
	}
	if u.Role != domain.UserRoleMember {
		t.Fatalf("default role = %q, want member", u.Role)
	}
	logs, err := repos.AuditLogs.List(context.Background())
	if err != nil {
		t.Fatalf("AuditLogs.List err = %v", err)
	}
	if len(logs) != 1 {
		t.Fatalf("audit log count = %d, want 1", len(logs))
	}
}

func TestService_ProjectTaskNoteSessionBasicFlows(t *testing.T) {
	repos := newMemoryRepos()
	now := time.Unix(1710000000, 0)
	svcs, err := service.NewServices(repos, newSequenceIDGenerator("id"), fixedNow(now))
	if err != nil {
		t.Fatalf("NewServices err = %v", err)
	}
	ctx := context.Background()

	u, err := svcs.Users.Create(ctx, service.CreateUserInput{
		Username:    "u1",
		DisplayName: "U1",
		Email:       "u1@example.com",
	})
	if err != nil {
		t.Fatalf("Users.Create err = %v", err)
	}

	p, err := svcs.Projects.Create(ctx, service.CreateProjectInput{
		OwnerID:    u.ID,
		Name:       "P1",
		Visibility: domain.ProjectVisibilityPrivate,
	})
	if err != nil {
		t.Fatalf("Projects.Create err = %v", err)
	}

	tk, err := svcs.Tasks.Create(ctx, service.CreateTaskInput{
		ProjectID: p.ID,
		CreatorID: u.ID,
		Title:     "T1",
	})
	if err != nil {
		t.Fatalf("Tasks.Create err = %v", err)
	}
	if _, err := svcs.Tasks.Start(ctx, service.TaskActionInput{TaskID: tk.ID}); err != nil {
		t.Fatalf("Tasks.Start err = %v", err)
	}

	n, err := svcs.Notes.Create(ctx, service.CreateNoteInput{
		ProjectID: p.ID,
		AuthorID:  u.ID,
		Title:     "N1",
	})
	if err != nil {
		t.Fatalf("Notes.Create err = %v", err)
	}
	if _, err := svcs.Notes.Archive(ctx, service.NoteActionInput{NoteID: n.ID}); err != nil {
		t.Fatalf("Notes.Archive err = %v", err)
	}

	s, err := svcs.Sessions.Create(ctx, service.CreateSessionInput{
		UserID:    u.ID,
		TokenHash: "token-1",
		TTL:       time.Hour,
	})
	if err != nil {
		t.Fatalf("Sessions.Create err = %v", err)
	}
	if _, err := svcs.Sessions.Touch(ctx, service.SessionTouchInput{SessionID: s.ID}); err != nil {
		t.Fatalf("Sessions.Touch err = %v", err)
	}
}

func TestService_ProjectCreateRequiresOwner(t *testing.T) {
	repos := newMemoryRepos()
	now := time.Unix(1710000000, 0)
	svcs, err := service.NewServices(repos, newSequenceIDGenerator("id"), fixedNow(now))
	if err != nil {
		t.Fatalf("NewServices err = %v", err)
	}

	_, err = svcs.Projects.Create(context.Background(), service.CreateProjectInput{
		OwnerID: domain.UserID("missing"),
		Name:    "P1",
	})
	if err == nil {
		t.Fatal("Projects.Create without owner err = nil")
	}
}

func TestService_UserLifecycleAndQueries(t *testing.T) {
	repos := newMemoryRepos()
	now := time.Unix(1710000000, 0)
	svcs, err := service.NewServices(repos, newSequenceIDGenerator("id"), fixedNow(now))
	if err != nil {
		t.Fatalf("NewServices err = %v", err)
	}
	ctx := context.Background()

	u, err := svcs.Users.Create(ctx, service.CreateUserInput{
		Username:    "user-q",
		DisplayName: "User Q",
		Email:       "user-q@example.com",
	})
	if err != nil {
		t.Fatalf("Users.Create err = %v", err)
	}

	if _, err := svcs.Users.GetByID(ctx, u.ID); err != nil {
		t.Fatalf("Users.GetByID err = %v", err)
	}
	if _, err := svcs.Users.GetByUsername(ctx, u.Username); err != nil {
		t.Fatalf("Users.GetByUsername err = %v", err)
	}
	users, err := svcs.Users.List(ctx)
	if err != nil {
		t.Fatalf("Users.List err = %v", err)
	}
	if len(users) != 1 {
		t.Fatalf("Users.List len = %d, want 1", len(users))
	}

	if _, err := svcs.Users.Suspend(ctx, service.UserActionInput{UserID: u.ID}); err != nil {
		t.Fatalf("Users.Suspend err = %v", err)
	}
	got, err := svcs.Users.GetByID(ctx, u.ID)
	if err != nil {
		t.Fatalf("Users.GetByID(after suspend) err = %v", err)
	}
	if got.Status != domain.UserStatusSuspended {
		t.Fatalf("status = %q, want suspended", got.Status)
	}

	if _, err := svcs.Users.Activate(ctx, service.UserActionInput{UserID: u.ID}); err != nil {
		t.Fatalf("Users.Activate err = %v", err)
	}
}

func TestService_ProjectLifecycleAndQueries(t *testing.T) {
	repos := newMemoryRepos()
	now := time.Unix(1710000000, 0)
	svcs, err := service.NewServices(repos, newSequenceIDGenerator("id"), fixedNow(now))
	if err != nil {
		t.Fatalf("NewServices err = %v", err)
	}
	ctx := context.Background()

	owner, err := svcs.Users.Create(ctx, service.CreateUserInput{
		Username:    "owner-q",
		DisplayName: "Owner Q",
		Email:       "owner-q@example.com",
	})
	if err != nil {
		t.Fatalf("Users.Create(owner) err = %v", err)
	}

	p, err := svcs.Projects.Create(ctx, service.CreateProjectInput{
		OwnerID: owner.ID,
		Name:    "Project Q",
	})
	if err != nil {
		t.Fatalf("Projects.Create err = %v", err)
	}

	if _, err := svcs.Projects.GetByID(ctx, p.ID); err != nil {
		t.Fatalf("Projects.GetByID err = %v", err)
	}
	if list, err := svcs.Projects.List(ctx); err != nil || len(list) != 1 {
		t.Fatalf("Projects.List got len=%d err=%v", len(list), err)
	}
	if list, err := svcs.Projects.ListByOwnerID(ctx, owner.ID); err != nil || len(list) != 1 {
		t.Fatalf("Projects.ListByOwnerID got len=%d err=%v", len(list), err)
	}

	if _, err := svcs.Projects.Archive(ctx, service.ProjectActionInput{ProjectID: p.ID}); err != nil {
		t.Fatalf("Projects.Archive err = %v", err)
	}
	archived, _ := svcs.Projects.GetByID(ctx, p.ID)
	if archived.Status != domain.ProjectStatusArchived {
		t.Fatalf("status = %q, want archived", archived.Status)
	}

	if _, err := svcs.Projects.Reopen(ctx, service.ProjectActionInput{ProjectID: p.ID}); err != nil {
		t.Fatalf("Projects.Reopen err = %v", err)
	}

	newName := "Project Q2"
	newVisibility := domain.ProjectVisibilityPublic
	updated, err := svcs.Projects.Update(ctx, service.UpdateProjectInput{
		ProjectID:  p.ID,
		Name:       &newName,
		Visibility: &newVisibility,
	})
	if err != nil {
		t.Fatalf("Projects.Update err = %v", err)
	}
	if updated.Name != newName {
		t.Fatalf("updated name = %q, want %q", updated.Name, newName)
	}

	if err := svcs.Projects.Delete(ctx, service.ProjectActionInput{ProjectID: p.ID}); err != nil {
		t.Fatalf("Projects.Delete err = %v", err)
	}
	if _, err := svcs.Projects.GetByID(ctx, p.ID); err == nil {
		t.Fatal("Projects.GetByID after delete err = nil")
	}
}

func TestService_TaskLifecycleAndQueries(t *testing.T) {
	repos := newMemoryRepos()
	now := time.Unix(1710000000, 0)
	svcs, err := service.NewServices(repos, newSequenceIDGenerator("id"), fixedNow(now))
	if err != nil {
		t.Fatalf("NewServices err = %v", err)
	}
	ctx := context.Background()

	creator, err := svcs.Users.Create(ctx, service.CreateUserInput{
		Username:    "creator-q",
		DisplayName: "Creator Q",
		Email:       "creator-q@example.com",
	})
	if err != nil {
		t.Fatalf("Users.Create(creator) err = %v", err)
	}
	assignee, err := svcs.Users.Create(ctx, service.CreateUserInput{
		Username:    "assignee-q",
		DisplayName: "Assignee Q",
		Email:       "assignee-q@example.com",
	})
	if err != nil {
		t.Fatalf("Users.Create(assignee) err = %v", err)
	}
	p, err := svcs.Projects.Create(ctx, service.CreateProjectInput{
		OwnerID: creator.ID,
		Name:    "Task Project",
	})
	if err != nil {
		t.Fatalf("Projects.Create err = %v", err)
	}

	task, err := svcs.Tasks.Create(ctx, service.CreateTaskInput{
		ProjectID:  p.ID,
		CreatorID:  creator.ID,
		AssigneeID: &assignee.ID,
		Title:      "Task Q",
	})
	if err != nil {
		t.Fatalf("Tasks.Create err = %v", err)
	}

	if _, err := svcs.Tasks.GetByID(ctx, task.ID); err != nil {
		t.Fatalf("Tasks.GetByID err = %v", err)
	}
	if list, err := svcs.Tasks.List(ctx); err != nil || len(list) != 1 {
		t.Fatalf("Tasks.List len=%d err=%v", len(list), err)
	}
	if list, err := svcs.Tasks.ListByProjectID(ctx, p.ID); err != nil || len(list) != 1 {
		t.Fatalf("Tasks.ListByProjectID len=%d err=%v", len(list), err)
	}
	if list, err := svcs.Tasks.ListByAssigneeID(ctx, assignee.ID); err != nil || len(list) != 1 {
		t.Fatalf("Tasks.ListByAssigneeID len=%d err=%v", len(list), err)
	}

	if _, err := svcs.Tasks.Start(ctx, service.TaskActionInput{TaskID: task.ID}); err != nil {
		t.Fatalf("Tasks.Start err = %v", err)
	}
	if _, err := svcs.Tasks.MarkDone(ctx, service.TaskActionInput{TaskID: task.ID}); err != nil {
		t.Fatalf("Tasks.MarkDone err = %v", err)
	}
	if _, err := svcs.Tasks.Reopen(ctx, service.TaskActionInput{TaskID: task.ID}); err != nil {
		t.Fatalf("Tasks.Reopen err = %v", err)
	}
	if _, err := svcs.Tasks.Cancel(ctx, service.TaskActionInput{TaskID: task.ID}); err != nil {
		t.Fatalf("Tasks.Cancel err = %v", err)
	}
}

func TestService_NoteLifecycleAndQueries(t *testing.T) {
	repos := newMemoryRepos()
	now := time.Unix(1710000000, 0)
	svcs, err := service.NewServices(repos, newSequenceIDGenerator("id"), fixedNow(now))
	if err != nil {
		t.Fatalf("NewServices err = %v", err)
	}
	ctx := context.Background()

	author, err := svcs.Users.Create(ctx, service.CreateUserInput{
		Username:    "author-q",
		DisplayName: "Author Q",
		Email:       "author-q@example.com",
	})
	if err != nil {
		t.Fatalf("Users.Create(author) err = %v", err)
	}
	p, err := svcs.Projects.Create(ctx, service.CreateProjectInput{
		OwnerID: author.ID,
		Name:    "Note Project",
	})
	if err != nil {
		t.Fatalf("Projects.Create err = %v", err)
	}

	note, err := svcs.Notes.Create(ctx, service.CreateNoteInput{
		ProjectID: p.ID,
		AuthorID:  author.ID,
		Title:     "Note Q",
	})
	if err != nil {
		t.Fatalf("Notes.Create err = %v", err)
	}
	if _, err := svcs.Notes.GetByID(ctx, note.ID); err != nil {
		t.Fatalf("Notes.GetByID err = %v", err)
	}
	if list, err := svcs.Notes.List(ctx); err != nil || len(list) != 1 {
		t.Fatalf("Notes.List len=%d err=%v", len(list), err)
	}
	if list, err := svcs.Notes.ListByProjectID(ctx, p.ID); err != nil || len(list) != 1 {
		t.Fatalf("Notes.ListByProjectID len=%d err=%v", len(list), err)
	}
	if _, err := svcs.Notes.Archive(ctx, service.NoteActionInput{NoteID: note.ID}); err != nil {
		t.Fatalf("Notes.Archive err = %v", err)
	}
	if _, err := svcs.Notes.Restore(ctx, service.NoteActionInput{NoteID: note.ID}); err != nil {
		t.Fatalf("Notes.Restore err = %v", err)
	}
}

func TestService_SessionLifecycleAndQueries(t *testing.T) {
	repos := newMemoryRepos()
	now := time.Unix(1710000000, 0)
	svcs, err := service.NewServices(repos, newSequenceIDGenerator("id"), fixedNow(now))
	if err != nil {
		t.Fatalf("NewServices err = %v", err)
	}
	ctx := context.Background()

	u, err := svcs.Users.Create(ctx, service.CreateUserInput{
		Username:    "session-user",
		DisplayName: "Session User",
		Email:       "session-user@example.com",
	})
	if err != nil {
		t.Fatalf("Users.Create err = %v", err)
	}
	s, err := svcs.Sessions.Create(ctx, service.CreateSessionInput{
		UserID:    u.ID,
		TokenHash: "token-q",
		TTL:       time.Hour,
	})
	if err != nil {
		t.Fatalf("Sessions.Create err = %v", err)
	}
	if _, err := svcs.Sessions.GetByID(ctx, s.ID); err != nil {
		t.Fatalf("Sessions.GetByID err = %v", err)
	}
	if _, err := svcs.Sessions.GetByTokenHash(ctx, " token-q "); err != nil {
		t.Fatalf("Sessions.GetByTokenHash(trimmed) err = %v", err)
	}
	if list, err := svcs.Sessions.List(ctx); err != nil || len(list) != 1 {
		t.Fatalf("Sessions.List len=%d err=%v", len(list), err)
	}
	if list, err := svcs.Sessions.ListByUserID(ctx, u.ID); err != nil || len(list) != 1 {
		t.Fatalf("Sessions.ListByUserID len=%d err=%v", len(list), err)
	}
	if _, err := svcs.Sessions.Revoke(ctx, service.SessionActionInput{SessionID: s.ID}); err != nil {
		t.Fatalf("Sessions.Revoke err = %v", err)
	}
}

func TestService_SessionCreateRequiresPositiveTTL(t *testing.T) {
	repos := newMemoryRepos()
	now := time.Unix(1710000000, 0)
	svcs, err := service.NewServices(repos, newSequenceIDGenerator("id"), fixedNow(now))
	if err != nil {
		t.Fatalf("NewServices err = %v", err)
	}
	ctx := context.Background()
	u, err := svcs.Users.Create(ctx, service.CreateUserInput{
		Username:    "ttl-user",
		DisplayName: "TTL User",
		Email:       "ttl-user@example.com",
	})
	if err != nil {
		t.Fatalf("Users.Create err = %v", err)
	}

	if _, err := svcs.Sessions.Create(ctx, service.CreateSessionInput{
		UserID:    u.ID,
		TokenHash: "token-ttl",
		TTL:       0,
	}); err == nil {
		t.Fatal("Sessions.Create with non-positive TTL err = nil")
	}
}

func TestService_AuditLogServiceAppendAndQueries(t *testing.T) {
	repos := newMemoryRepos()
	now := time.Unix(1710000000, 0)
	svcs, err := service.NewServices(repos, newSequenceIDGenerator("id"), fixedNow(now))
	if err != nil {
		t.Fatalf("NewServices err = %v", err)
	}
	ctx := context.Background()
	actor := domain.UserID("actor-1")

	log, err := svcs.AuditLogs.Append(ctx, service.AppendAuditLogInput{
		ActorID:      &actor,
		ResourceType: domain.AuditResourceTask,
		ResourceID:   "  t-1  ",
		Action:       domain.AuditActionUpdate,
		RequestID:    " req-1 ",
		Detail:       "update",
	})
	if err != nil {
		t.Fatalf("AuditLogs.Append err = %v", err)
	}
	if log.ResourceID != "t-1" {
		t.Fatalf("ResourceID = %q, want t-1", log.ResourceID)
	}
	if log.RequestID != "req-1" {
		t.Fatalf("RequestID = %q, want req-1", log.RequestID)
	}
	if _, err := svcs.AuditLogs.GetByID(ctx, log.ID); err != nil {
		t.Fatalf("AuditLogs.GetByID err = %v", err)
	}
	if list, err := svcs.AuditLogs.List(ctx); err != nil || len(list) != 1 {
		t.Fatalf("AuditLogs.List len=%d err=%v", len(list), err)
	}
	if list, err := svcs.AuditLogs.ListByActorID(ctx, actor); err != nil || len(list) != 1 {
		t.Fatalf("AuditLogs.ListByActorID len=%d err=%v", len(list), err)
	}
	if list, err := svcs.AuditLogs.ListByResource(ctx, domain.AuditResourceTask, " t-1 "); err != nil || len(list) != 1 {
		t.Fatalf("AuditLogs.ListByResource len=%d err=%v", len(list), err)
	}
}

func TestService_RandomIDGenerator(t *testing.T) {
	g := service.RandomIDGenerator{}
	id1 := g.NewID()
	id2 := g.NewID()
	if id1 == "" || id2 == "" {
		t.Fatalf("RandomIDGenerator returned empty id: %q %q", id1, id2)
	}
	if id1 == id2 {
		t.Fatalf("RandomIDGenerator generated duplicate ids: %q", id1)
	}
}

func TestRisk_WriteErrorShouldNotPersistUserWhenAuditFails(t *testing.T) {
	repos := newMemoryRepos()
	repos.AuditLogs = &failingAuditLogRepository{createErr: errors.New("audit down")}
	now := time.Unix(1710000000, 0)
	svcs, err := service.NewServices(repos, newSequenceIDGenerator("id"), fixedNow(now))
	if err != nil {
		t.Fatalf("NewServices err = %v", err)
	}

	_, err = svcs.Users.Create(context.Background(), service.CreateUserInput{
		Username:    "risk-user",
		DisplayName: "Risk User",
		Email:       "risk@example.com",
	})
	if err == nil {
		t.Fatal("Users.Create expected error when audit fails")
	}
	if _, err := repos.Users.GetByUsername(context.Background(), "risk-user"); err == nil {
		t.Fatalf("semantic risk: Users.Create returned error but user is persisted")
	}
}

func TestRisk_WriteErrorShouldNotPersistProjectTaskNoteSessionWhenAuditFails(t *testing.T) {
	repos := newMemoryRepos()
	repos.AuditLogs = &failingAuditLogRepository{createErr: errors.New("audit down")}
	now := time.Unix(1710000000, 0)
	svcs, err := service.NewServices(repos, newSequenceIDGenerator("id"), fixedNow(now))
	if err != nil {
		t.Fatalf("NewServices err = %v", err)
	}
	ctx := context.Background()

	owner := &domain.User{
		ID:          domain.UserID("owner-1"),
		Username:    "owner",
		DisplayName: "Owner",
		Email:       "owner@example.com",
		Role:        domain.UserRoleMember,
		Status:      domain.UserStatusActive,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := repos.Users.Create(ctx, owner); err != nil {
		t.Fatalf("seed owner err = %v", err)
	}

	project, err := svcs.Projects.Create(ctx, service.CreateProjectInput{OwnerID: owner.ID, Name: "Risk Project"})
	if err == nil {
		t.Fatal("Projects.Create expected error when audit fails")
	}
	if list, _ := repos.Projects.List(ctx); len(list) != 0 {
		t.Fatalf("semantic risk: project persisted despite returned error")
	}

	seedProject := &domain.Project{
		ID:         domain.ProjectID("p-seed"),
		OwnerID:    owner.ID,
		Name:       "Seed",
		Visibility: domain.ProjectVisibilityPrivate,
		Status:     domain.ProjectStatusActive,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := repos.Projects.Create(ctx, seedProject); err != nil {
		t.Fatalf("seed project err = %v", err)
	}

	_, err = svcs.Tasks.Create(ctx, service.CreateTaskInput{ProjectID: seedProject.ID, CreatorID: owner.ID, Title: "Risk Task"})
	if err == nil {
		t.Fatal("Tasks.Create expected error when audit fails")
	}
	if list, _ := repos.Tasks.List(ctx); len(list) != 0 {
		t.Fatalf("semantic risk: task persisted despite returned error")
	}

	_, err = svcs.Notes.Create(ctx, service.CreateNoteInput{ProjectID: seedProject.ID, AuthorID: owner.ID, Title: "Risk Note"})
	if err == nil {
		t.Fatal("Notes.Create expected error when audit fails")
	}
	if list, _ := repos.Notes.List(ctx); len(list) != 0 {
		t.Fatalf("semantic risk: note persisted despite returned error")
	}

	_, err = svcs.Sessions.Create(ctx, service.CreateSessionInput{UserID: owner.ID, TokenHash: "risk-token", TTL: time.Hour})
	if err == nil {
		t.Fatal("Sessions.Create expected error when audit fails")
	}
	if list, _ := repos.Sessions.List(ctx); len(list) != 0 {
		t.Fatalf("semantic risk: session persisted despite returned error")
	}

	_ = project
}

func TestRisk_ConcurrentSessionRevokeAndTouchMustNotResurrect(t *testing.T) {
	repos := newMemoryRepos()
	now := time.Unix(1710000000, 0)

	owner := &domain.User{
		ID:          domain.UserID("u-1"),
		Username:    "u1",
		DisplayName: "U1",
		Email:       "u1@example.com",
		Role:        domain.UserRoleMember,
		Status:      domain.UserStatusActive,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := repos.Users.Create(context.Background(), owner); err != nil {
		t.Fatalf("seed user err = %v", err)
	}
	sessionID := domain.SessionID("s-1")
	if err := repos.Sessions.Create(context.Background(), &domain.Session{
		ID:        sessionID,
		UserID:    owner.ID,
		TokenHash: "token-1",
		Status:    domain.SessionStatusActive,
		CreatedAt: now,
		ExpiresAt: now.Add(time.Hour),
	}); err != nil {
		t.Fatalf("seed session err = %v", err)
	}

	startUpdate := make(chan struct{})
	revokeDone := make(chan struct{})
	var getCount atomic.Int32
	repos.Sessions = &scriptedSessionRepository{
		inner: repos.Sessions,
		onGetByID: func() {
			if getCount.Add(1) == 2 {
				close(startUpdate)
			}
		},
		startUpdate: startUpdate,
		revokeDone:  revokeDone,
	}

	svcs, err := service.NewServices(repos, newSequenceIDGenerator("id"), fixedNow(now.Add(2*time.Minute)))
	if err != nil {
		t.Fatalf("NewServices err = %v", err)
	}

	ctx := context.Background()
	var wg sync.WaitGroup
	wg.Add(2)
	var revokeErr error
	var touchErr error
	go func() {
		defer wg.Done()
		_, revokeErr = svcs.Sessions.Revoke(ctx, service.SessionActionInput{SessionID: sessionID})
	}()
	go func() {
		defer wg.Done()
		_, touchErr = svcs.Sessions.Touch(ctx, service.SessionTouchInput{SessionID: sessionID})
	}()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("concurrency test timeout")
	}

	if revokeErr != nil || touchErr != nil {
		t.Fatalf("unexpected errors: revoke=%v touch=%v", revokeErr, touchErr)
	}
	final, err := repos.Sessions.GetByID(ctx, sessionID)
	if err != nil {
		t.Fatalf("GetByID(final) err = %v", err)
	}
	if final.Status != domain.SessionStatusRevoked {
		t.Fatalf("semantic risk: final session status = %q, want revoked", final.Status)
	}
}

func TestRisk_ConcurrentTaskMarkDoneAndCancelMustRespectStateMachine(t *testing.T) {
	repos := newMemoryRepos()
	now := time.Unix(1710000000, 0)

	owner := &domain.User{
		ID:          domain.UserID("u-1"),
		Username:    "u1",
		DisplayName: "U1",
		Email:       "u1@example.com",
		Role:        domain.UserRoleMember,
		Status:      domain.UserStatusActive,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := repos.Users.Create(context.Background(), owner); err != nil {
		t.Fatalf("seed user err = %v", err)
	}
	if err := repos.Projects.Create(context.Background(), &domain.Project{
		ID:         domain.ProjectID("p-1"),
		OwnerID:    owner.ID,
		Name:       "P1",
		Visibility: domain.ProjectVisibilityPrivate,
		Status:     domain.ProjectStatusActive,
		CreatedAt:  now,
		UpdatedAt:  now,
	}); err != nil {
		t.Fatalf("seed project err = %v", err)
	}
	taskID := domain.TaskID("t-1")
	if err := repos.Tasks.Create(context.Background(), &domain.Task{
		ID:        taskID,
		ProjectID: domain.ProjectID("p-1"),
		CreatorID: owner.ID,
		Title:     "T1",
		Status:    domain.TaskStatusTodo,
		Priority:  domain.TaskPriorityMedium,
		CreatedAt: now,
		UpdatedAt: now,
	}); err != nil {
		t.Fatalf("seed task err = %v", err)
	}

	startUpdate := make(chan struct{})
	doneDone := make(chan struct{})
	var getCount atomic.Int32
	repos.Tasks = &scriptedTaskRepository{
		inner: repos.Tasks,
		onGetByID: func() {
			if getCount.Add(1) == 2 {
				close(startUpdate)
			}
		},
		startUpdate: startUpdate,
		doneDone:    doneDone,
	}

	svcs, err := service.NewServices(repos, newSequenceIDGenerator("id"), fixedNow(now.Add(2*time.Minute)))
	if err != nil {
		t.Fatalf("NewServices err = %v", err)
	}

	ctx := context.Background()
	var wg sync.WaitGroup
	wg.Add(2)
	var markDoneErr error
	var cancelErr error
	go func() {
		defer wg.Done()
		_, markDoneErr = svcs.Tasks.MarkDone(ctx, service.TaskActionInput{TaskID: taskID})
	}()
	go func() {
		defer wg.Done()
		_, cancelErr = svcs.Tasks.Cancel(ctx, service.TaskActionInput{TaskID: taskID})
	}()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("concurrency test timeout")
	}

	if markDoneErr != nil || cancelErr != nil {
		t.Fatalf("unexpected errors: markDone=%v cancel=%v", markDoneErr, cancelErr)
	}
	final, err := repos.Tasks.GetByID(ctx, taskID)
	if err != nil {
		t.Fatalf("GetByID(final) err = %v", err)
	}
	if final.Status != domain.TaskStatusDone {
		t.Fatalf("semantic risk: final task status = %q, want done", final.Status)
	}
}
