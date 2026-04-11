package transporthttp

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"HaruhiServer/internal/apperr"
	"HaruhiServer/internal/domain"
	"HaruhiServer/internal/response"
	"HaruhiServer/internal/service"
)

type createTaskRequest struct {
	CreatorID   string  `json:"creator_id"`
	AssigneeID  *string `json:"assignee_id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Priority    string  `json:"priority"`
	DueAt       *string `json:"due_at"`
}

type patchTaskRequest struct {
	Title         *string `json:"title"`
	Description   *string `json:"description"`
	AssigneeID    *string `json:"assignee_id"`
	ClearAssignee bool    `json:"clear_assignee"`
	Priority      *string `json:"priority"`
	DueAt         *string `json:"due_at"`
	ClearDueAt    bool    `json:"clear_due_at"`
}

func (h *Handler) projectTasks(w http.ResponseWriter, r *http.Request, projectID domain.ProjectID) {
	switch r.Method {
	case http.MethodGet:
		h.projectTaskList(w, r, projectID)
	case http.MethodPost:
		h.projectTaskCreate(w, r, projectID)
	default:
		response.Error(w, apperr.New(apperr.CodeMethodNotAllowed, "method not allowed"))
	}
}

func (h *Handler) projectTaskByID(w http.ResponseWriter, r *http.Request, projectID domain.ProjectID, taskID domain.TaskID) {
	switch r.Method {
	case http.MethodGet:
		h.projectTaskGet(w, r, projectID, taskID)
	case http.MethodPatch:
		h.projectTaskPatch(w, r, projectID, taskID)
	case http.MethodDelete:
		h.projectTaskDelete(w, r, projectID, taskID)
	default:
		response.Error(w, apperr.New(apperr.CodeMethodNotAllowed, "method not allowed"))
	}
}

func (h *Handler) projectTaskList(w http.ResponseWriter, r *http.Request, projectID domain.ProjectID) {
	if _, err := h.services.Projects.GetByID(r.Context(), projectID); err != nil {
		response.Error(w, mapError(err))
		return
	}

	tasks, err := h.services.Tasks.ListByProjectID(r.Context(), projectID)
	if err != nil {
		response.Error(w, mapError(err))
		return
	}

	response.OK(w, map[string]any{"items": tasks})
}

func (h *Handler) projectTaskCreate(w http.ResponseWriter, r *http.Request, projectID domain.ProjectID) {
	var req createTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, apperr.Wrap(apperr.CodeInvalidArgument, "invalid json body", err))
		return
	}

	assigneeID := taskAssigneeID(req.AssigneeID)
	dueAt, err := parseTaskDueAt(req.DueAt)
	if err != nil {
		response.Error(w, err)
		return
	}

	task, err := h.services.Tasks.Create(r.Context(), service.CreateTaskInput{
		ProjectID:   projectID,
		CreatorID:   domain.UserID(req.CreatorID),
		AssigneeID:  assigneeID,
		Title:       req.Title,
		Description: req.Description,
		Priority:    domain.TaskPriority(req.Priority),
		DueAt:       dueAt,
		Meta: service.ActionMeta{
			RequestID: RequestIDFromContext(r.Context()),
		},
	})
	if err != nil {
		response.Error(w, mapError(err))
		return
	}

	response.OK(w, task)
}

func (h *Handler) projectTaskGet(w http.ResponseWriter, r *http.Request, projectID domain.ProjectID, taskID domain.TaskID) {
	task, err := h.getTaskInProject(r.Context(), projectID, taskID)
	if err != nil {
		response.Error(w, mapError(err))
		return
	}

	response.OK(w, task)
}

func (h *Handler) projectTaskPatch(w http.ResponseWriter, r *http.Request, projectID domain.ProjectID, taskID domain.TaskID) {
	if _, err := h.getTaskInProject(r.Context(), projectID, taskID); err != nil {
		response.Error(w, mapError(err))
		return
	}

	var req patchTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, apperr.Wrap(apperr.CodeInvalidArgument, "invalid json body", err))
		return
	}

	var priority *domain.TaskPriority
	if req.Priority != nil {
		p := domain.TaskPriority(*req.Priority)
		priority = &p
	}

	assigneeID := taskAssigneeID(req.AssigneeID)
	dueAt, err := parseTaskDueAt(req.DueAt)
	if err != nil {
		response.Error(w, err)
		return
	}

	task, err := h.services.Tasks.Update(r.Context(), service.UpdateTaskInput{
		TaskID:        taskID,
		Title:         req.Title,
		Description:   req.Description,
		AssigneeID:    assigneeID,
		ClearAssignee: req.ClearAssignee,
		Priority:      priority,
		DueAt:         dueAt,
		ClearDueAt:    req.ClearDueAt,
		Meta: service.ActionMeta{
			RequestID: RequestIDFromContext(r.Context()),
		},
	})
	if err != nil {
		response.Error(w, mapError(err))
		return
	}

	response.OK(w, task)
}

func (h *Handler) projectTaskDelete(w http.ResponseWriter, r *http.Request, projectID domain.ProjectID, taskID domain.TaskID) {
	if _, err := h.getTaskInProject(r.Context(), projectID, taskID); err != nil {
		response.Error(w, mapError(err))
		return
	}

	if err := h.services.Tasks.Delete(r.Context(), service.TaskActionInput{
		TaskID: taskID,
		Meta: service.ActionMeta{
			RequestID: RequestIDFromContext(r.Context()),
		},
	}); err != nil {
		response.Error(w, mapError(err))
		return
	}

	response.OK(w, map[string]any{"deleted": true})
}

func (h *Handler) getTaskInProject(ctx context.Context, projectID domain.ProjectID, taskID domain.TaskID) (*domain.Task, error) {
	task, err := h.services.Tasks.GetByID(ctx, taskID)
	if err != nil {
		return nil, err
	}
	if task.ProjectID != projectID {
		return nil, domain.NewDomainError(domain.ErrNotFound, "task not found")
	}
	return task, nil
}

func parseProjectTaskCollectionPath(path string) (domain.ProjectID, bool) {
	parts, ok := splitProjectSubpath(path)
	if !ok || len(parts) != 2 || parts[1] != "tasks" {
		return "", false
	}
	return domain.ProjectID(parts[0]), true
}

func parseProjectTaskItemPath(path string) (domain.ProjectID, domain.TaskID, bool) {
	parts, ok := splitProjectSubpath(path)
	if !ok || len(parts) != 3 || parts[1] != "tasks" {
		return "", "", false
	}
	return domain.ProjectID(parts[0]), domain.TaskID(parts[2]), true
}

func splitProjectSubpath(path string) ([]string, bool) {
	if !strings.HasPrefix(path, "/api/v1/projects/") {
		return nil, false
	}
	rest := strings.Trim(strings.TrimPrefix(path, "/api/v1/projects/"), "/")
	if rest == "" {
		return nil, false
	}

	parts := strings.Split(rest, "/")
	for _, part := range parts {
		if part == "" {
			return nil, false
		}
	}
	return parts, true
}

func taskAssigneeID(v *string) *domain.UserID {
	if v == nil {
		return nil
	}
	id := domain.UserID(strings.TrimSpace(*v))
	return &id
}

func parseTaskDueAt(v *string) (*time.Time, error) {
	if v == nil {
		return nil, nil
	}
	ts, err := time.Parse(time.RFC3339, strings.TrimSpace(*v))
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeInvalidArgument, "due_at must be RFC3339", err)
	}
	return &ts, nil
}
