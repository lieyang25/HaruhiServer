package service

import (
	"context"
	"fmt"

	"HaruhiServer/internal/domain"
)

type taskService struct {
	deps deps
}

var _ TaskService = (*taskService)(nil)

func (s *taskService) Create(ctx context.Context, in CreateTaskInput) (*domain.Task, error) {
	if _, err := s.deps.repos.Projects.GetByID(ctx, in.ProjectID); err != nil {
		return nil, err
	}
	if _, err := s.deps.repos.Users.GetByID(ctx, in.CreatorID); err != nil {
		return nil, err
	}
	if in.AssigneeID != nil {
		if _, err := s.deps.repos.Users.GetByID(ctx, *in.AssigneeID); err != nil {
			return nil, err
		}
	}

	now := s.deps.now()

	priority := in.Priority
	if priority == "" {
		priority = domain.TaskPriorityMedium
	}

	task := &domain.Task{
		ID:          domain.TaskID(s.deps.idg.NewID()),
		ProjectID:   in.ProjectID,
		CreatorID:   in.CreatorID,
		AssigneeID:  cloneUserIDPtr(in.AssigneeID),
		Title:       in.Title,
		Description: in.Description,
		Status:      domain.TaskStatusTodo,
		Priority:    priority,
		DueAt:       cloneTimePtr(in.DueAt),
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := task.Validate(); err != nil {
		return nil, err
	}
	if err := s.deps.repos.Tasks.Create(ctx, task); err != nil {
		return nil, err
	}

	if err := recordAudit(
		ctx,
		s.deps.repos,
		s.deps.idg,
		now,
		in.Meta,
		domain.AuditResourceTask,
		string(task.ID),
		domain.AuditActionCreate,
		fmt.Sprintf("create task %s", task.Title),
	); err != nil {
		rollbackErr := s.deps.repos.Tasks.Delete(ctx, task.ID)
		return nil, joinAuditAndRollbackError(err, "rollback task create", rollbackErr)
	}

	return task, nil
}

func (s *taskService) GetByID(ctx context.Context, id domain.TaskID) (*domain.Task, error) {
	return s.deps.repos.Tasks.GetByID(ctx, id)
}

func (s *taskService) List(ctx context.Context) ([]*domain.Task, error) {
	return s.deps.repos.Tasks.List(ctx)
}

func (s *taskService) ListByProjectID(ctx context.Context, projectID domain.ProjectID) ([]*domain.Task, error) {
	return s.deps.repos.Tasks.ListByProjectID(ctx, projectID)
}

func (s *taskService) ListByAssigneeID(ctx context.Context, assigneeID domain.UserID) ([]*domain.Task, error) {
	return s.deps.repos.Tasks.ListByAssigneeID(ctx, assigneeID)
}

func (s *taskService) Start(ctx context.Context, in TaskActionInput) (*domain.Task, error) {
	task, err := s.deps.repos.Tasks.GetByID(ctx, in.TaskID)
	if err != nil {
		return nil, err
	}

	now := s.deps.now()
	if err := task.Start(now); err != nil {
		return nil, err
	}
	if err := s.deps.repos.Tasks.Update(ctx, task); err != nil {
		return nil, err
	}

	if err := recordAudit(
		ctx,
		s.deps.repos,
		s.deps.idg,
		now,
		in.Meta,
		domain.AuditResourceTask,
		string(task.ID),
		domain.AuditActionUpdate,
		"start task",
	); err != nil {
		return nil, err
	}

	return task, nil
}

func (s *taskService) MarkDone(ctx context.Context, in TaskActionInput) (*domain.Task, error) {
	task, err := s.deps.repos.Tasks.GetByID(ctx, in.TaskID)
	if err != nil {
		return nil, err
	}

	now := s.deps.now()
	if err := task.MarkDone(now); err != nil {
		return nil, err
	}
	if err := s.deps.repos.Tasks.Update(ctx, task); err != nil {
		return nil, err
	}

	if err := recordAudit(
		ctx,
		s.deps.repos,
		s.deps.idg,
		now,
		in.Meta,
		domain.AuditResourceTask,
		string(task.ID),
		domain.AuditActionUpdate,
		"mark task done",
	); err != nil {
		return nil, err
	}

	return task, nil
}

func (s *taskService) Cancel(ctx context.Context, in TaskActionInput) (*domain.Task, error) {
	task, err := s.deps.repos.Tasks.GetByID(ctx, in.TaskID)
	if err != nil {
		return nil, err
	}

	now := s.deps.now()
	if err := task.Cancel(now); err != nil {
		return nil, err
	}
	if err := s.deps.repos.Tasks.Update(ctx, task); err != nil {
		if de := domain.AsDomainError(err); de != nil && de.Code == domain.ErrInvalidState {
			latest, latestErr := s.deps.repos.Tasks.GetByID(ctx, in.TaskID)
			if latestErr == nil && latest.Status == domain.TaskStatusDone {
				return latest, nil
			}
		}
		return nil, err
	}

	if err := recordAudit(
		ctx,
		s.deps.repos,
		s.deps.idg,
		now,
		in.Meta,
		domain.AuditResourceTask,
		string(task.ID),
		domain.AuditActionUpdate,
		"cancel task",
	); err != nil {
		return nil, err
	}

	return task, nil
}

func (s *taskService) Reopen(ctx context.Context, in TaskActionInput) (*domain.Task, error) {
	task, err := s.deps.repos.Tasks.GetByID(ctx, in.TaskID)
	if err != nil {
		return nil, err
	}

	now := s.deps.now()
	if err := task.Reopen(now); err != nil {
		return nil, err
	}
	if err := s.deps.repos.Tasks.Update(ctx, task); err != nil {
		return nil, err
	}

	if err := recordAudit(
		ctx,
		s.deps.repos,
		s.deps.idg,
		now,
		in.Meta,
		domain.AuditResourceTask,
		string(task.ID),
		domain.AuditActionRestore,
		"reopen task",
	); err != nil {
		return nil, err
	}

	return task, nil
}
