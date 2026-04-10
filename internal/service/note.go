package service

import (
	"context"
	"fmt"

	"HaruhiServer/internal/domain"
)

type noteService struct {
	deps deps
}

var _ NoteService = (*noteService)(nil)

func (s *noteService) Create(ctx context.Context, in CreateNoteInput) (*domain.Note, error) {
	if _, err := s.deps.repos.Projects.GetByID(ctx, in.ProjectID); err != nil {
		return nil, err
	}
	if _, err := s.deps.repos.Users.GetByID(ctx, in.AuthorID); err != nil {
		return nil, err
	}

	now := s.deps.now()

	format := in.Format
	if format == "" {
		format = domain.NoteFormatPlainText
	}

	note := &domain.Note{
		ID:        domain.NoteID(s.deps.idg.NewID()),
		ProjectID: in.ProjectID,
		AuthorID:  in.AuthorID,
		Title:     in.Title,
		Body:      in.Body,
		Format:    format,
		Status:    domain.NoteStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := note.Validate(); err != nil {
		return nil, err
	}
	if err := s.deps.repos.Notes.Create(ctx, note); err != nil {
		return nil, err
	}

	if err := recordAudit(
		ctx,
		s.deps.repos,
		s.deps.idg,
		now,
		in.Meta,
		domain.AuditResourceNote,
		string(note.ID),
		domain.AuditActionCreate,
		fmt.Sprintf("create note %s", note.Title),
	); err != nil {
		rollbackErr := s.deps.repos.Notes.Delete(ctx, note.ID)
		return nil, joinAuditAndRollbackError(err, "rollback note create", rollbackErr)
	}

	return note, nil
}

func (s *noteService) GetByID(ctx context.Context, id domain.NoteID) (*domain.Note, error) {
	return s.deps.repos.Notes.GetByID(ctx, id)
}

func (s *noteService) List(ctx context.Context) ([]*domain.Note, error) {
	return s.deps.repos.Notes.List(ctx)
}

func (s *noteService) ListByProjectID(ctx context.Context, projectID domain.ProjectID) ([]*domain.Note, error) {
	return s.deps.repos.Notes.ListByProjectID(ctx, projectID)
}

func (s *noteService) Archive(ctx context.Context, in NoteActionInput) (*domain.Note, error) {
	note, err := s.deps.repos.Notes.GetByID(ctx, in.NoteID)
	if err != nil {
		return nil, err
	}

	now := s.deps.now()
	if err := note.Archive(now); err != nil {
		return nil, err
	}
	if err := s.deps.repos.Notes.Update(ctx, note); err != nil {
		return nil, err
	}

	if err := recordAudit(
		ctx,
		s.deps.repos,
		s.deps.idg,
		now,
		in.Meta,
		domain.AuditResourceNote,
		string(note.ID),
		domain.AuditActionArchive,
		"archive note",
	); err != nil {
		return nil, err
	}

	return note, nil
}

func (s *noteService) Restore(ctx context.Context, in NoteActionInput) (*domain.Note, error) {
	note, err := s.deps.repos.Notes.GetByID(ctx, in.NoteID)
	if err != nil {
		return nil, err
	}

	now := s.deps.now()
	if err := note.Restore(now); err != nil {
		return nil, err
	}
	if err := s.deps.repos.Notes.Update(ctx, note); err != nil {
		return nil, err
	}

	if err := recordAudit(
		ctx,
		s.deps.repos,
		s.deps.idg,
		now,
		in.Meta,
		domain.AuditResourceNote,
		string(note.ID),
		domain.AuditActionRestore,
		"restore note",
	); err != nil {
		return nil, err
	}

	return note, nil
}
