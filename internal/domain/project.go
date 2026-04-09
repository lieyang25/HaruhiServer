package domain

import "time"

type ProjectVisibility string

const (
	ProjectVisibilityPrivate ProjectVisibility = "private"
	ProjectVisibilityTeam    ProjectVisibility = "team"
	ProjectVisibilityPublic  ProjectVisibility = "public"
)

func (v ProjectVisibility) IsValid() bool {
	switch v {
	case ProjectVisibilityPrivate, ProjectVisibilityTeam, ProjectVisibilityPublic:
		return true
	default:
		return false
	}
}

type ProjectStatus string

const (
	ProjectStatusActive   ProjectStatus = "active"
	ProjectStatusArchived ProjectStatus = "archived"
)

func (s ProjectStatus) IsValid() bool {
	switch s {
	case ProjectStatusActive, ProjectStatusArchived:
		return true
	default:
		return false
	}
}

type Project struct {
	ID          ProjectID
	OwnerID     UserID
	Name        string
	Description string
	Visibility  ProjectVisibility
	Status      ProjectStatus
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (p *Project) Validate() error {
	if err := requireNonEmpty("project.id", string(p.ID)); err != nil {
		return err
	}
	if err := requireNonEmpty("project.owner_id", string(p.OwnerID)); err != nil {
		return err
	}
	if err := requireNonEmpty("project.name", p.Name); err != nil {
		return err
	}
	if err := requireMaxRunes("project.name", p.Name, 128); err != nil {
		return err
	}
	if err := requireMaxRunes("project.description", p.Description, 1024); err != nil {
		return err
	}
	if !p.Visibility.IsValid() {
		return NewDomainError(ErrInvalidArgument, "project.visibility is invalid")
	}
	if !p.Status.IsValid() {
		return NewDomainError(ErrInvalidArgument, "project.status is invalid")
	}
	if p.CreatedAt.IsZero() {
		return NewDomainError(ErrInvalidArgument, "project.created_at is required")
	}
	if p.UpdatedAt.IsZero() {
		return NewDomainError(ErrInvalidArgument, "project.updated_at is required")
	}
	return nil
}

func (p *Project) Archive(now time.Time) error {
	if p.Status == ProjectStatusArchived {
		return NewDomainError(ErrInvalidState, "project is already archived")
	}
	p.Status = ProjectStatusArchived
	p.UpdatedAt = now
	return nil
}

func (p *Project) Reopen(now time.Time) error {
	if p.Status == ProjectStatusActive {
		return NewDomainError(ErrInvalidState, "project is already active")
	}
	p.Status = ProjectStatusActive
	p.UpdatedAt = now
	return nil
}
