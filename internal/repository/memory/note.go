package memory

import (
	"context"
	"sync"

	"HaruhiServer/internal/domain"
)

type NoteRepository struct {
	mu   sync.RWMutex
	byID map[domain.NoteID]*domain.Note
}

func NewNoteRepository() *NoteRepository {
	return &NoteRepository{
		byID: make(map[domain.NoteID]*domain.Note),
	}
}

func (r *NoteRepository) Create(ctx context.Context, note *domain.Note) error {
	if err := checkContext(ctx); err != nil {
		return err
	}
	if note == nil {
		return domain.NewDomainError(domain.ErrInvalidArgument, "note is required")
	}
	if err := note.Validate(); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.byID[note.ID]; ok {
		return domain.NewDomainError(domain.ErrConflict, "note id already exists")
	}

	r.byID[note.ID] = cloneNote(note)
	return nil
}

func (r *NoteRepository) Update(ctx context.Context, note *domain.Note) error {
	if err := checkContext(ctx); err != nil {
		return err
	}
	if note == nil {
		return domain.NewDomainError(domain.ErrInvalidArgument, "note is required")
	}
	if err := note.Validate(); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.byID[note.ID]; !ok {
		return domain.NewDomainError(domain.ErrNotFound, "note not found")
	}

	r.byID[note.ID] = cloneNote(note)
	return nil
}

func (r *NoteRepository) GetByID(ctx context.Context, id domain.NoteID) (*domain.Note, error) {
	if err := checkContext(ctx); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	note, ok := r.byID[id]
	if !ok {
		return nil, domain.NewDomainError(domain.ErrNotFound, "note not found")
	}

	return cloneNote(note), nil
}

func (r *NoteRepository) Delete(ctx context.Context, id domain.NoteID) error {
	if err := checkContext(ctx); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.byID[id]; !ok {
		return domain.NewDomainError(domain.ErrNotFound, "note not found")
	}

	delete(r.byID, id)
	return nil
}

func (r *NoteRepository) List(ctx context.Context) ([]*domain.Note, error) {
	if err := checkContext(ctx); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	items := make([]*domain.Note, 0, len(r.byID))
	for _, note := range r.byID {
		items = append(items, cloneNote(note))
	}

	sortNotes(items)
	return items, nil
}

func (r *NoteRepository) ListByProjectID(ctx context.Context, projectID domain.ProjectID) ([]*domain.Note, error) {
	if err := checkContext(ctx); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	items := make([]*domain.Note, 0)
	for _, note := range r.byID {
		if note.ProjectID == projectID {
			items = append(items, cloneNote(note))
		}
	}

	sortNotes(items)
	return items, nil
}
