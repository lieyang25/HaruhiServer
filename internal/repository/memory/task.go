package memory

import (
	"context"
	"sync"

	"HaruhiServer/internal/domain"
)

type TaskRepository struct {
	mu   sync.RWMutex
	byID map[domain.TaskID]*domain.Task
}

func NewTaskRepository() *TaskRepository {
	return &TaskRepository{
		byID: make(map[domain.TaskID]*domain.Task),
	}
}

func (r *TaskRepository) Create(ctx context.Context, task *domain.Task) error {
	if err := checkContext(ctx); err != nil {
		return err
	}
	if task == nil {
		return domain.NewDomainError(domain.ErrInvalidArgument, "task is required")
	}
	if err := task.Validate(); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.byID[task.ID]; ok {
		return domain.NewDomainError(domain.ErrConflict, "task id already exists")
	}

	r.byID[task.ID] = cloneTask(task)
	return nil
}

func (r *TaskRepository) Update(ctx context.Context, task *domain.Task) error {
	if err := checkContext(ctx); err != nil {
		return err
	}
	if task == nil {
		return domain.NewDomainError(domain.ErrInvalidArgument, "task is required")
	}
	if err := task.Validate(); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.byID[task.ID]; !ok {
		return domain.NewDomainError(domain.ErrNotFound, "task not found")
	}

	r.byID[task.ID] = cloneTask(task)
	return nil
}

func (r *TaskRepository) GetByID(ctx context.Context, id domain.TaskID) (*domain.Task, error) {
	if err := checkContext(ctx); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	task, ok := r.byID[id]
	if !ok {
		return nil, domain.NewDomainError(domain.ErrNotFound, "task not found")
	}

	return cloneTask(task), nil
}

func (r *TaskRepository) Delete(ctx context.Context, id domain.TaskID) error {
	if err := checkContext(ctx); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.byID[id]; !ok {
		return domain.NewDomainError(domain.ErrNotFound, "task not found")
	}

	delete(r.byID, id)
	return nil
}

func (r *TaskRepository) List(ctx context.Context) ([]*domain.Task, error) {
	if err := checkContext(ctx); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	items := make([]*domain.Task, 0, len(r.byID))
	for _, task := range r.byID {
		items = append(items, cloneTask(task))
	}

	sortTasks(items)
	return items, nil
}

func (r *TaskRepository) ListByProjectID(ctx context.Context, projectID domain.ProjectID) ([]*domain.Task, error) {
	if err := checkContext(ctx); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	items := make([]*domain.Task, 0)
	for _, task := range r.byID {
		if task.ProjectID == projectID {
			items = append(items, cloneTask(task))
		}
	}

	sortTasks(items)
	return items, nil
}

func (r *TaskRepository) ListByAssigneeID(ctx context.Context, assigneeID domain.UserID) ([]*domain.Task, error) {
	if err := checkContext(ctx); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	items := make([]*domain.Task, 0)
	for _, task := range r.byID {
		if task.AssigneeID != nil && *task.AssigneeID == assigneeID {
			items = append(items, cloneTask(task))
		}
	}

	sortTasks(items)
	return items, nil
}
