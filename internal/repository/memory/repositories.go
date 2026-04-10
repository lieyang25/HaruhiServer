package memory

import "HaruhiServer/internal/repository"

func NewRepositories() *repository.Repositories {
	return &repository.Repositories{
		Users:     NewUserRepository(),
		Projects:  NewProjectRepository(),
		Tasks:     NewTaskRepository(),
		Notes:     NewNoteRepository(),
		Sessions:  NewSessionRepository(),
		AuditLogs: NewAuditLogRepository(),
	}
}
