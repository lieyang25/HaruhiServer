package service

import (
	"context"
	"fmt"

	"HaruhiServer/internal/domain"
)

type projectService struct {
	deps deps
}

var _ ProjectService = (*projectService)(nil)

func (s *projectService) Create(ctx context.Context, in CreateProjectInput) (*domain.Project, error) {
	if _, err := s.deps.repos.Users.GetByID(ctx, in.OwnerID); err != nil {
		return nil, err
	}

	now := s.deps.now()

	visibility := in.Visibility
	if visibility == "" {
		visibility = domain.ProjectVisibilityPrivate
	}

	project := &domain.Project{
		ID:          domain.ProjectID(s.deps.idg.NewID()),
		OwnerID:     in.OwnerID,
		Name:        in.Name,
		Description: in.Description,
		Visibility:  visibility,
		Status:      domain.ProjectStatusActive,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := project.Validate(); err != nil {
		return nil, err
	}
	if err := s.deps.repos.Projects.Create(ctx, project); err != nil {
		return nil, err
	}

	if err := recordAudit(
		ctx,
		s.deps.repos,
		s.deps.idg,
		now,
		in.Meta,
		domain.AuditResourceProject,
		string(project.ID),
		domain.AuditActionCreate,
		fmt.Sprintf("create project %s", project.Name),
	); err != nil {
		return nil, err
	}

	return project, nil
}

func (s *projectService) GetByID(ctx context.Context, id domain.ProjectID) (*domain.Project, error) {
	return s.deps.repos.Projects.GetByID(ctx, id)
}

func (s *projectService) List(ctx context.Context) ([]*domain.Project, error) {
	return s.deps.repos.Projects.List(ctx)
}

func (s *projectService) ListByOwnerID(ctx context.Context, ownerID domain.UserID) ([]*domain.Project, error) {
	return s.deps.repos.Projects.ListByOwnerID(ctx, ownerID)
}

func (s *projectService) Archive(ctx context.Context, in ProjectActionInput) (*domain.Project, error) {
	project, err := s.deps.repos.Projects.GetByID(ctx, in.ProjectID)
	if err != nil {
		return nil, err
	}

	now := s.deps.now()
	if err := project.Archive(now); err != nil {
		return nil, err
	}
	if err := s.deps.repos.Projects.Update(ctx, project); err != nil {
		return nil, err
	}

	if err := recordAudit(
		ctx,
		s.deps.repos,
		s.deps.idg,
		now,
		in.Meta,
		domain.AuditResourceProject,
		string(project.ID),
		domain.AuditActionArchive,
		"archive project",
	); err != nil {
		return nil, err
	}

	return project, nil
}

func (s *projectService) Reopen(ctx context.Context, in ProjectActionInput) (*domain.Project, error) {
	project, err := s.deps.repos.Projects.GetByID(ctx, in.ProjectID)
	if err != nil {
		return nil, err
	}

	now := s.deps.now()
	if err := project.Reopen(now); err != nil {
		return nil, err
	}
	if err := s.deps.repos.Projects.Update(ctx, project); err != nil {
		return nil, err
	}

	if err := recordAudit(
		ctx,
		s.deps.repos,
		s.deps.idg,
		now,
		in.Meta,
		domain.AuditResourceProject,
		string(project.ID),
		domain.AuditActionRestore,
		"reopen project",
	); err != nil {
		return nil, err
	}

	return project, nil
}
