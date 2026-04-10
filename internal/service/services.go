package service

import (
	"fmt"

	"HaruhiServer/internal/repository"
)

type Services struct {
	Users     UserService
	Projects  ProjectService
	Tasks     TaskService
	Notes     NoteService
	Sessions  SessionService
	AuditLogs AuditLogService
}

func NewServices(repos *repository.Repositories, idg IDGenerator, now NowFunc) (*Services, error) {
	if repos == nil {
		return nil, fmt.Errorf("repositories is required")
	}
	if repos.Users == nil {
		return nil, fmt.Errorf("user repository is required")
	}
	if repos.Projects == nil {
		return nil, fmt.Errorf("project repository is required")
	}
	if repos.Tasks == nil {
		return nil, fmt.Errorf("task repository is required")
	}
	if repos.Notes == nil {
		return nil, fmt.Errorf("note repository is required")
	}
	if repos.Sessions == nil {
		return nil, fmt.Errorf("session repository is required")
	}
	if repos.AuditLogs == nil {
		return nil, fmt.Errorf("audit log repository is required")
	}

	d := newDeps(repos, idg, now)

	return &Services{
		Users:     &userService{deps: d},
		Projects:  &projectService{deps: d},
		Tasks:     &taskService{deps: d},
		Notes:     &noteService{deps: d},
		Sessions:  &sessionService{deps: d},
		AuditLogs: &auditLogService{deps: d},
	}, nil
}
