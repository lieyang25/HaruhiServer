package transporthttp

import (
	"encoding/json"
	"net/http"
	"strings"

	"HaruhiServer/internal/apperr"
	"HaruhiServer/internal/domain"
	"HaruhiServer/internal/response"
	"HaruhiServer/internal/service"
)

type createProjectRequest struct {
	OwnerID     string `json:"owner_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Visibility  string `json:"visibility"`
}

type patchProjectRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Visibility  *string `json:"visibility"`
	Archived    *bool   `json:"archived"`
}

func (h *Handler) projects(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.projectsList(w, r)
	case http.MethodPost:
		h.projectsCreate(w, r)
	default:
		response.Error(w, apperr.New(apperr.CodeMethodNotAllowed, "method not allowed"))
	}
}

func (h *Handler) projectByID(w http.ResponseWriter, r *http.Request) {
	if projectID, ok := parseProjectTaskCollectionPath(r.URL.Path); ok {
		h.projectTasks(w, r, projectID)
		return
	}
	if projectID, taskID, ok := parseProjectTaskItemPath(r.URL.Path); ok {
		h.projectTaskByID(w, r, projectID, taskID)
		return
	}

	id, ok := parseProjectID(r.URL.Path)
	if !ok {
		response.Error(w, apperr.New(apperr.CodeNotFound, "project not found"))
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.projectGet(w, r, id)
	case http.MethodPatch:
		h.projectPatch(w, r, id)
	case http.MethodDelete:
		h.projectDelete(w, r, id)
	default:
		response.Error(w, apperr.New(apperr.CodeMethodNotAllowed, "method not allowed"))
	}
}

func (h *Handler) projectsCreate(w http.ResponseWriter, r *http.Request) {
	var req createProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, apperr.Wrap(apperr.CodeInvalidArgument, "invalid json body", err))
		return
	}

	project, err := h.services.Projects.Create(r.Context(), service.CreateProjectInput{
		OwnerID:     domain.UserID(req.OwnerID),
		Name:        req.Name,
		Description: req.Description,
		Visibility:  domain.ProjectVisibility(req.Visibility),
		Meta: service.ActionMeta{
			RequestID: RequestIDFromContext(r.Context()),
		},
	})
	if err != nil {
		response.Error(w, mapError(err))
		return
	}

	response.OK(w, project)
}

func (h *Handler) projectsList(w http.ResponseWriter, r *http.Request) {
	projects, err := h.services.Projects.List(r.Context())
	if err != nil {
		response.Error(w, mapError(err))
		return
	}

	response.OK(w, map[string]any{"items": projects})
}

func (h *Handler) projectGet(w http.ResponseWriter, r *http.Request, id domain.ProjectID) {
	project, err := h.services.Projects.GetByID(r.Context(), id)
	if err != nil {
		response.Error(w, mapError(err))
		return
	}

	response.OK(w, project)
}

func (h *Handler) projectPatch(w http.ResponseWriter, r *http.Request, id domain.ProjectID) {
	var req patchProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, apperr.Wrap(apperr.CodeInvalidArgument, "invalid json body", err))
		return
	}

	var visibility *domain.ProjectVisibility
	if req.Visibility != nil {
		v := domain.ProjectVisibility(*req.Visibility)
		visibility = &v
	}

	project, err := h.services.Projects.Update(r.Context(), service.UpdateProjectInput{
		ProjectID:   id,
		Name:        req.Name,
		Description: req.Description,
		Visibility:  visibility,
		Archive:     req.Archived,
		Meta: service.ActionMeta{
			RequestID: RequestIDFromContext(r.Context()),
		},
	})
	if err != nil {
		response.Error(w, mapError(err))
		return
	}

	response.OK(w, project)
}

func (h *Handler) projectDelete(w http.ResponseWriter, r *http.Request, id domain.ProjectID) {
	if err := h.services.Projects.Delete(r.Context(), service.ProjectActionInput{
		ProjectID: id,
		Meta: service.ActionMeta{
			RequestID: RequestIDFromContext(r.Context()),
		},
	}); err != nil {
		response.Error(w, mapError(err))
		return
	}

	response.OK(w, map[string]any{"deleted": true})
}

func parseProjectID(path string) (domain.ProjectID, bool) {
	if !strings.HasPrefix(path, "/api/v1/projects/") {
		return "", false
	}
	id := strings.TrimPrefix(path, "/api/v1/projects/")
	id = strings.Trim(id, "/")
	if id == "" || strings.Contains(id, "/") {
		return "", false
	}
	return domain.ProjectID(id), true
}

func mapError(err error) error {
	if de := domain.AsDomainError(err); de != nil {
		switch de.Code {
		case domain.ErrInvalidArgument, domain.ErrInvalidState:
			return apperr.Wrap(apperr.CodeInvalidArgument, de.Message, err)
		case domain.ErrNotFound:
			return apperr.Wrap(apperr.CodeNotFound, de.Message, err)
		case domain.ErrConflict:
			return apperr.Wrap(apperr.CodeConflict, de.Message, err)
		default:
			return apperr.Wrap(apperr.CodeInternal, "internal error", err)
		}
	}
	return apperr.Wrap(apperr.CodeInternal, "internal error", err)
}
