package domain

import "time"

type NoteFormat string

const (
	NoteFormatPlainText NoteFormat = "plain_text"
	NoteFormatMarkdown  NoteFormat = "markdown"
)

func (f NoteFormat) IsValid() bool {
	switch f {
	case NoteFormatPlainText, NoteFormatMarkdown:
		return true
	default:
		return false
	}
}

type NoteStatus string

const (
	NoteStatusActive   NoteStatus = "active"
	NoteStatusArchived NoteStatus = "archived"
)

func (s NoteStatus) IsValid() bool {
	switch s {
	case NoteStatusActive, NoteStatusArchived:
		return true
	default:
		return false
	}
}

type Note struct {
	ID        NoteID
	ProjectID ProjectID
	AuthorID  UserID
	Title     string
	Body      string
	Format    NoteFormat
	Status    NoteStatus
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (n *Note) Validate() error {
	if err := requireNonEmpty("note.id", string(n.ID)); err != nil {
		return err
	}
	if err := requireNonEmpty("note.project_id", string(n.ProjectID)); err != nil {
		return err
	}
	if err := requireNonEmpty("note.author_id", string(n.AuthorID)); err != nil {
		return err
	}
	if err := requireNonEmpty("note.title", n.Title); err != nil {
		return err
	}
	if err := requireMaxRunes("note.title", n.Title, 256); err != nil {
		return err
	}
	if !n.Format.IsValid() {
		return NewDomainError(ErrInvalidArgument, "note.format is invalid")
	}
	if !n.Status.IsValid() {
		return NewDomainError(ErrInvalidArgument, "note.status is invalid")
	}
	if n.CreatedAt.IsZero() {
		return NewDomainError(ErrInvalidArgument, "note.created_at is required")
	}
	if n.UpdatedAt.IsZero() {
		return NewDomainError(ErrInvalidArgument, "note.updated_at is required")
	}
	return nil
}

func (n *Note) Archive(now time.Time) error {
	if n.Status == NoteStatusArchived {
		return NewDomainError(ErrInvalidState, "note is already archived")
	}
	n.Status = NoteStatusArchived
	n.UpdatedAt = now
	return nil
}

func (n *Note) Restore(now time.Time) error {
	if n.Status == NoteStatusActive {
		return NewDomainError(ErrInvalidState, "note is already active")
	}
	n.Status = NoteStatusActive
	n.UpdatedAt = now
	return nil
}
