package repository_test

import (
	"context"
	"testing"
	"time"

	"HaruhiServer/internal/domain"
	"HaruhiServer/internal/repository/memory"
)

func makeUser(id, username string, at time.Time) *domain.User {
	return &domain.User{
		ID:          domain.UserID(id),
		Username:    username,
		DisplayName: username,
		Email:       username + "@example.com",
		Role:        domain.UserRoleMember,
		Status:      domain.UserStatusActive,
		CreatedAt:   at,
		UpdatedAt:   at,
	}
}

func makeProject(id string, owner domain.UserID, at time.Time) *domain.Project {
	return &domain.Project{
		ID:         domain.ProjectID(id),
		OwnerID:    owner,
		Name:       id,
		Visibility: domain.ProjectVisibilityPrivate,
		Status:     domain.ProjectStatusActive,
		CreatedAt:  at,
		UpdatedAt:  at,
	}
}

func makeTask(id string, project domain.ProjectID, creator domain.UserID, assignee *domain.UserID, at time.Time) *domain.Task {
	return &domain.Task{
		ID:         domain.TaskID(id),
		ProjectID:  project,
		CreatorID:  creator,
		AssigneeID: assignee,
		Title:      id,
		Status:     domain.TaskStatusTodo,
		Priority:   domain.TaskPriorityMedium,
		CreatedAt:  at,
		UpdatedAt:  at,
	}
}

func makeNote(id string, project domain.ProjectID, author domain.UserID, at time.Time) *domain.Note {
	return &domain.Note{
		ID:        domain.NoteID(id),
		ProjectID: project,
		AuthorID:  author,
		Title:     id,
		Format:    domain.NoteFormatPlainText,
		Status:    domain.NoteStatusActive,
		CreatedAt: at,
		UpdatedAt: at,
	}
}

func makeSession(id string, user domain.UserID, token string, at time.Time) *domain.Session {
	return &domain.Session{
		ID:        domain.SessionID(id),
		UserID:    user,
		TokenHash: token,
		Status:    domain.SessionStatusActive,
		CreatedAt: at,
		ExpiresAt: at.Add(time.Hour),
	}
}

func makeAuditLog(id string, actor *domain.UserID, resType domain.AuditResourceType, resID string, at time.Time) *domain.AuditLog {
	return &domain.AuditLog{
		ID:           domain.AuditLogID(id),
		ActorID:      actor,
		ResourceType: resType,
		ResourceID:   resID,
		Action:       domain.AuditActionUpdate,
		CreatedAt:    at,
	}
}

func TestMemoryUserRepository_CRUDIndexCloneAndSort(t *testing.T) {
	repo := memory.NewUserRepository()
	ctx := context.Background()
	base := time.Unix(1710000000, 0)

	u2 := makeUser("u2", "user2", base.Add(time.Minute))
	u1 := makeUser("u1", "user1", base)
	if err := repo.Create(ctx, u2); err != nil {
		t.Fatalf("Create(u2) err = %v", err)
	}
	if err := repo.Create(ctx, u1); err != nil {
		t.Fatalf("Create(u1) err = %v", err)
	}
	if err := repo.Create(ctx, makeUser("u3", "user1", base)); err == nil {
		t.Fatal("duplicate username create err = nil")
	}

	got, err := repo.GetByUsername(ctx, "user1")
	if err != nil {
		t.Fatalf("GetByUsername err = %v", err)
	}
	got.Username = "mutated"
	again, err := repo.GetByID(ctx, domain.UserID("u1"))
	if err != nil {
		t.Fatalf("GetByID err = %v", err)
	}
	if again.Username != "user1" {
		t.Fatalf("clone isolation broken, username = %q", again.Username)
	}

	u1Updated := makeUser("u1", "user1-new", base)
	if err := repo.Update(ctx, u1Updated); err != nil {
		t.Fatalf("Update err = %v", err)
	}
	if _, err := repo.GetByUsername(ctx, "user1"); err == nil {
		t.Fatal("old username still searchable after update")
	}
	if _, err := repo.GetByUsername(ctx, "user1-new"); err != nil {
		t.Fatalf("new username not searchable: %v", err)
	}

	list, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("List err = %v", err)
	}
	if len(list) != 2 || string(list[0].ID) != "u1" || string(list[1].ID) != "u2" {
		t.Fatalf("List order = %#v", list)
	}

	if err := repo.Delete(ctx, domain.UserID("u1")); err != nil {
		t.Fatalf("Delete err = %v", err)
	}
	if _, err := repo.GetByUsername(ctx, "user1-new"); err == nil {
		t.Fatal("username index not cleaned after delete")
	}
}

func TestMemoryProjectTaskNoteRepositories_FilterAndClone(t *testing.T) {
	ctx := context.Background()
	base := time.Unix(1710000000, 0)
	owner := domain.UserID("u1")
	otherOwner := domain.UserID("u2")

	projectRepo := memory.NewProjectRepository()
	if err := projectRepo.Create(ctx, makeProject("p1", owner, base)); err != nil {
		t.Fatalf("project create p1 err = %v", err)
	}
	if err := projectRepo.Create(ctx, makeProject("p2", otherOwner, base.Add(time.Minute))); err != nil {
		t.Fatalf("project create p2 err = %v", err)
	}
	projects, err := projectRepo.ListByOwnerID(ctx, owner)
	if err != nil {
		t.Fatalf("ListByOwnerID err = %v", err)
	}
	if len(projects) != 1 || string(projects[0].ID) != "p1" {
		t.Fatalf("projects = %#v", projects)
	}
	projects[0].Name = "mutated"
	p1, _ := projectRepo.GetByID(ctx, domain.ProjectID("p1"))
	if p1.Name != "p1" {
		t.Fatalf("project clone isolation broken, got %q", p1.Name)
	}

	taskRepo := memory.NewTaskRepository()
	assignee := domain.UserID("u3")
	if err := taskRepo.Create(ctx, makeTask("t1", domain.ProjectID("p1"), owner, &assignee, base)); err != nil {
		t.Fatalf("task create t1 err = %v", err)
	}
	if err := taskRepo.Create(ctx, makeTask("t2", domain.ProjectID("p2"), owner, nil, base.Add(time.Minute))); err != nil {
		t.Fatalf("task create t2 err = %v", err)
	}
	byProject, err := taskRepo.ListByProjectID(ctx, domain.ProjectID("p1"))
	if err != nil {
		t.Fatalf("ListByProjectID err = %v", err)
	}
	if len(byProject) != 1 || string(byProject[0].ID) != "t1" {
		t.Fatalf("byProject = %#v", byProject)
	}
	byAssignee, err := taskRepo.ListByAssigneeID(ctx, assignee)
	if err != nil {
		t.Fatalf("ListByAssigneeID err = %v", err)
	}
	if len(byAssignee) != 1 || string(byAssignee[0].ID) != "t1" {
		t.Fatalf("byAssignee = %#v", byAssignee)
	}
	byAssignee[0].Title = "mutated"
	t1, _ := taskRepo.GetByID(ctx, domain.TaskID("t1"))
	if t1.Title != "t1" {
		t.Fatalf("task clone isolation broken, got %q", t1.Title)
	}

	noteRepo := memory.NewNoteRepository()
	if err := noteRepo.Create(ctx, makeNote("n1", domain.ProjectID("p1"), owner, base)); err != nil {
		t.Fatalf("note create n1 err = %v", err)
	}
	if err := noteRepo.Create(ctx, makeNote("n2", domain.ProjectID("p2"), owner, base.Add(time.Minute))); err != nil {
		t.Fatalf("note create n2 err = %v", err)
	}
	byNoteProject, err := noteRepo.ListByProjectID(ctx, domain.ProjectID("p1"))
	if err != nil {
		t.Fatalf("note ListByProjectID err = %v", err)
	}
	if len(byNoteProject) != 1 || string(byNoteProject[0].ID) != "n1" {
		t.Fatalf("byNoteProject = %#v", byNoteProject)
	}
}

func TestMemorySessionRepository_TokenIndexAndClone(t *testing.T) {
	repo := memory.NewSessionRepository()
	ctx := context.Background()
	base := time.Unix(1710000000, 0)
	uid := domain.UserID("u1")

	s1 := makeSession("s1", uid, "token-a", base)
	if err := repo.Create(ctx, s1); err != nil {
		t.Fatalf("Create(s1) err = %v", err)
	}
	if err := repo.Create(ctx, makeSession("s2", uid, "token-a", base)); err == nil {
		t.Fatal("duplicate token hash create err = nil")
	}

	s1upd := makeSession("s1", uid, "token-b", base)
	now := base.Add(5 * time.Minute)
	s1upd.LastSeenAt = &now
	if err := repo.Update(ctx, s1upd); err != nil {
		t.Fatalf("Update(s1) err = %v", err)
	}
	if _, err := repo.GetByTokenHash(ctx, "token-a"); err == nil {
		t.Fatal("old token hash still searchable after update")
	}
	byToken, err := repo.GetByTokenHash(ctx, "token-b")
	if err != nil {
		t.Fatalf("GetByTokenHash(new) err = %v", err)
	}
	if byToken.LastSeenAt == nil {
		t.Fatal("LastSeenAt missing")
	}
	*byToken.LastSeenAt = byToken.LastSeenAt.Add(time.Hour)
	again, _ := repo.GetByID(ctx, domain.SessionID("s1"))
	if again.LastSeenAt != nil && again.LastSeenAt.Equal(*byToken.LastSeenAt) {
		t.Fatal("session clone isolation broken")
	}

	if err := repo.Delete(ctx, domain.SessionID("s1")); err != nil {
		t.Fatalf("Delete err = %v", err)
	}
	if _, err := repo.GetByTokenHash(ctx, "token-b"); err == nil {
		t.Fatal("token index not cleaned after delete")
	}
}

func TestMemoryAuditLogRepository_FilterAndClone(t *testing.T) {
	repo := memory.NewAuditLogRepository()
	ctx := context.Background()
	base := time.Unix(1710000000, 0)
	actor1 := domain.UserID("u1")
	actor2 := domain.UserID("u2")

	if err := repo.Create(ctx, makeAuditLog("a1", &actor1, domain.AuditResourceTask, "t1", base)); err != nil {
		t.Fatalf("Create(a1) err = %v", err)
	}
	if err := repo.Create(ctx, makeAuditLog("a2", &actor2, domain.AuditResourceProject, "p1", base.Add(time.Minute))); err != nil {
		t.Fatalf("Create(a2) err = %v", err)
	}

	byActor, err := repo.ListByActorID(ctx, actor1)
	if err != nil {
		t.Fatalf("ListByActorID err = %v", err)
	}
	if len(byActor) != 1 || string(byActor[0].ID) != "a1" {
		t.Fatalf("byActor = %#v", byActor)
	}
	byResource, err := repo.ListByResource(ctx, domain.AuditResourceTask, "t1")
	if err != nil {
		t.Fatalf("ListByResource err = %v", err)
	}
	if len(byResource) != 1 || string(byResource[0].ID) != "a1" {
		t.Fatalf("byResource = %#v", byResource)
	}

	fetched, _ := repo.GetByID(ctx, domain.AuditLogID("a1"))
	if fetched.ActorID == nil {
		t.Fatal("ActorID should not be nil")
	}
	*fetched.ActorID = domain.UserID("mutated")
	again, _ := repo.GetByID(ctx, domain.AuditLogID("a1"))
	if again.ActorID != nil && *again.ActorID == domain.UserID("mutated") {
		t.Fatal("audit log clone isolation broken")
	}
}

func TestMemoryRepositories_UpdateDeleteAndListByUser(t *testing.T) {
	ctx := context.Background()
	base := time.Unix(1710000000, 0)
	userID := domain.UserID("u1")

	projectRepo := memory.NewProjectRepository()
	p := makeProject("p1", userID, base)
	if err := projectRepo.Create(ctx, p); err != nil {
		t.Fatalf("project create err = %v", err)
	}
	p.Name = "p1-updated"
	if err := projectRepo.Update(ctx, p); err != nil {
		t.Fatalf("project update err = %v", err)
	}
	gotP, _ := projectRepo.GetByID(ctx, p.ID)
	if gotP.Name != "p1-updated" {
		t.Fatalf("project name = %q, want p1-updated", gotP.Name)
	}
	if err := projectRepo.Delete(ctx, p.ID); err != nil {
		t.Fatalf("project delete err = %v", err)
	}
	if err := projectRepo.Delete(ctx, p.ID); err == nil {
		t.Fatal("project second delete err = nil")
	}

	taskRepo := memory.NewTaskRepository()
	tk := makeTask("t1", domain.ProjectID("p1"), userID, nil, base)
	if err := taskRepo.Create(ctx, tk); err != nil {
		t.Fatalf("task create err = %v", err)
	}
	tk.Title = "t1-updated"
	if err := taskRepo.Update(ctx, tk); err != nil {
		t.Fatalf("task update err = %v", err)
	}
	gotT, _ := taskRepo.GetByID(ctx, tk.ID)
	if gotT.Title != "t1-updated" {
		t.Fatalf("task title = %q, want t1-updated", gotT.Title)
	}
	if err := taskRepo.Delete(ctx, tk.ID); err != nil {
		t.Fatalf("task delete err = %v", err)
	}
	if err := taskRepo.Delete(ctx, tk.ID); err == nil {
		t.Fatal("task second delete err = nil")
	}

	noteRepo := memory.NewNoteRepository()
	n := makeNote("n1", domain.ProjectID("p1"), userID, base)
	if err := noteRepo.Create(ctx, n); err != nil {
		t.Fatalf("note create err = %v", err)
	}
	n.Title = "n1-updated"
	if err := noteRepo.Update(ctx, n); err != nil {
		t.Fatalf("note update err = %v", err)
	}
	gotN, _ := noteRepo.GetByID(ctx, n.ID)
	if gotN.Title != "n1-updated" {
		t.Fatalf("note title = %q, want n1-updated", gotN.Title)
	}
	if err := noteRepo.Delete(ctx, n.ID); err != nil {
		t.Fatalf("note delete err = %v", err)
	}
	if err := noteRepo.Delete(ctx, n.ID); err == nil {
		t.Fatal("note second delete err = nil")
	}

	sessionRepo := memory.NewSessionRepository()
	s1 := makeSession("s1", userID, "token-1", base)
	s2 := makeSession("s2", domain.UserID("u2"), "token-2", base.Add(time.Minute))
	if err := sessionRepo.Create(ctx, s1); err != nil {
		t.Fatalf("session create s1 err = %v", err)
	}
	if err := sessionRepo.Create(ctx, s2); err != nil {
		t.Fatalf("session create s2 err = %v", err)
	}
	byUser, err := sessionRepo.ListByUserID(ctx, userID)
	if err != nil {
		t.Fatalf("session ListByUserID err = %v", err)
	}
	if len(byUser) != 1 || string(byUser[0].ID) != "s1" {
		t.Fatalf("session by user = %#v", byUser)
	}

	auditRepo := memory.NewAuditLogRepository()
	actor := userID
	if err := auditRepo.Create(ctx, makeAuditLog("a1", &actor, domain.AuditResourceUser, "u1", base)); err != nil {
		t.Fatalf("audit create err = %v", err)
	}
	if err := auditRepo.Create(ctx, makeAuditLog("a1", &actor, domain.AuditResourceUser, "u1", base)); err == nil {
		t.Fatal("audit duplicate id create err = nil")
	}
}

func TestMemoryRepositories_ContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := memory.NewUserRepository().List(ctx); err == nil {
		t.Fatal("user repo List(canceled) err = nil")
	}
	if _, err := memory.NewProjectRepository().List(ctx); err == nil {
		t.Fatal("project repo List(canceled) err = nil")
	}
	if _, err := memory.NewTaskRepository().List(ctx); err == nil {
		t.Fatal("task repo List(canceled) err = nil")
	}
	if _, err := memory.NewNoteRepository().List(ctx); err == nil {
		t.Fatal("note repo List(canceled) err = nil")
	}
	if _, err := memory.NewSessionRepository().List(ctx); err == nil {
		t.Fatal("session repo List(canceled) err = nil")
	}
	if _, err := memory.NewAuditLogRepository().List(ctx); err == nil {
		t.Fatal("audit repo List(canceled) err = nil")
	}
}
