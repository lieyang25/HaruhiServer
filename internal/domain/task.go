package domain

import "time"

type TaskStatus string

const (
	TaskStatusTodo       TaskStatus = "todo"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusDone       TaskStatus = "done"
	TaskStatusCanceled   TaskStatus = "canceled"
)

func (s TaskStatus) IsValid() bool {
	switch s {
	case TaskStatusTodo, TaskStatusInProgress, TaskStatusDone, TaskStatusCanceled:
		return true
	default:
		return false
	}
}

type TaskPriority string

const (
	TaskPriorityLow    TaskPriority = "low"
	TaskPriorityMedium TaskPriority = "medium"
	TaskPriorityHigh   TaskPriority = "high"
	TaskPriorityUrgent TaskPriority = "urgent"
)

func (p TaskPriority) IsValid() bool {
	switch p {
	case TaskPriorityLow, TaskPriorityMedium, TaskPriorityHigh, TaskPriorityUrgent:
		return true
	default:
		return false
	}
}

type Task struct {
	ID          TaskID
	ProjectID   ProjectID
	CreatorID   UserID
	AssigneeID  *UserID
	Title       string
	Description string
	Status      TaskStatus
	Priority    TaskPriority
	DueAt       *time.Time
	CompletedAt *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (t *Task) Validate() error {
	if err := requireNonEmpty("task.id", string(t.ID)); err != nil {
		return err
	}
	if err := requireNonEmpty("task.project_id", string(t.ProjectID)); err != nil {
		return err
	}
	if err := requireNonEmpty("task.creator_id", string(t.CreatorID)); err != nil {
		return err
	}
	if err := requireNonEmpty("task.title", t.Title); err != nil {
		return err
	}
	if err := requireMaxRunes("task.title", t.Title, 256); err != nil {
		return err
	}
	if err := requireMaxRunes("task.description", t.Description, 4096); err != nil {
		return err
	}
	if !t.Status.IsValid() {
		return NewDomainError(ErrInvalidArgument, "task.status is invalid")
	}
	if !t.Priority.IsValid() {
		return NewDomainError(ErrInvalidArgument, "task.priority is invalid")
	}
	if t.CreatedAt.IsZero() {
		return NewDomainError(ErrInvalidArgument, "task.created_at is required")
	}
	if t.UpdatedAt.IsZero() {
		return NewDomainError(ErrInvalidArgument, "task.updated_at is required")
	}
	if t.DueAt != nil && t.DueAt.Before(t.CreatedAt) {
		return NewDomainError(ErrInvalidArgument, "task.due_at cannot be before created_at")
	}
	if t.Status == TaskStatusDone && t.CompletedAt == nil {
		return NewDomainError(ErrInvalidState, "task.completed_at is required when status is done")
	}
	if t.Status != TaskStatusDone && t.CompletedAt != nil {
		return NewDomainError(ErrInvalidState, "task.completed_at must be nil when status is not done")
	}
	return nil
}

func (t *Task) Start(now time.Time) error {
	switch t.Status {
	case TaskStatusTodo:
		t.Status = TaskStatusInProgress
		t.UpdatedAt = now
		return nil
	case TaskStatusInProgress:
		return NewDomainError(ErrInvalidState, "task is already in progress")
	default:
		return NewDomainError(ErrInvalidState, "task cannot be started from current status")
	}
}

func (t *Task) MarkDone(now time.Time) error {
	if t.Status == TaskStatusCanceled {
		return NewDomainError(ErrInvalidState, "canceled task cannot be completed")
	}
	if t.Status == TaskStatusDone {
		return NewDomainError(ErrInvalidState, "task is already done")
	}

	t.Status = TaskStatusDone
	t.CompletedAt = &now
	t.UpdatedAt = now
	return nil
}

func (t *Task) Reopen(now time.Time) error {
	if t.Status != TaskStatusDone && t.Status != TaskStatusCanceled {
		return NewDomainError(ErrInvalidState, "only done or canceled task can be reopened")
	}

	t.Status = TaskStatusTodo
	t.CompletedAt = nil
	t.UpdatedAt = now
	return nil
}

func (t *Task) Cancel(now time.Time) error {
	if t.Status == TaskStatusDone {
		return NewDomainError(ErrInvalidState, "done task cannot be canceled directly")
	}
	if t.Status == TaskStatusCanceled {
		return NewDomainError(ErrInvalidState, "task is already canceled")
	}

	t.Status = TaskStatusCanceled
	t.CompletedAt = nil
	t.UpdatedAt = now
	return nil
}
