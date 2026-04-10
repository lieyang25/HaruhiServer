package memory

import (
	"context"
	"sync"

	"HaruhiServer/internal/domain"
)

type ProjectRepository struct {
	mu   sync.RWMutex
	byID map[domain.ProjectID]*domain.Project
}

func NewProjectRepository() *ProjectRepository {
	return &ProjectRepository{
		byID: make(map[domain.ProjectID]*domain.Project),
	}
}

func (r *ProjectRepository) Create(ctx context.Context, project *domain.Project) error {
	if err := checkContext(ctx); err != nil {
		return err
	}
	if project == nil {
		return domain.NewDomainError(domain.ErrInvalidArgument, "project is required")
	}
	if err := project.Validate(); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.byID[project.ID]; ok {
		return domain.NewDomainError(domain.ErrConflict, "project id already exists")
	}

	r.byID[project.ID] = cloneProject(project)
	return nil
}

func (r *ProjectRepository) Update(ctx context.Context, project *domain.Project) error {
	if err := checkContext(ctx); err != nil {
		return err
	}
	if project == nil {
		return domain.NewDomainError(domain.ErrInvalidArgument, "project is required")
	}
	if err := project.Validate(); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.byID[project.ID]; !ok {
		return domain.NewDomainError(domain.ErrNotFound, "project not found")
	}

	r.byID[project.ID] = cloneProject(project)
	return nil
}

func (r *ProjectRepository) GetByID(ctx context.Context, id domain.ProjectID) (*domain.Project, error) {
	if err := checkContext(ctx); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	project, ok := r.byID[id]
	if !ok {
		return nil, domain.NewDomainError(domain.ErrNotFound, "project not found")
	}

	return cloneProject(project), nil
}

func (r *ProjectRepository) Delete(ctx context.Context, id domain.ProjectID) error {
	if err := checkContext(ctx); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.byID[id]; !ok {
		return domain.NewDomainError(domain.ErrNotFound, "project not found")
	}

	delete(r.byID, id)
	return nil
}

func (r *ProjectRepository) List(ctx context.Context) ([]*domain.Project, error) {
	if err := checkContext(ctx); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	items := make([]*domain.Project, 0, len(r.byID))
	for _, project := range r.byID {
		items = append(items, cloneProject(project))
	}

	sortProjects(items)
	return items, nil
}

func (r *ProjectRepository) ListByOwnerID(ctx context.Context, ownerID domain.UserID) ([]*domain.Project, error) {
	if err := checkContext(ctx); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	items := make([]*domain.Project, 0)
	for _, project := range r.byID {
		if project.OwnerID == ownerID {
			items = append(items, cloneProject(project))
		}
	}

	sortProjects(items)
	return items, nil
}
