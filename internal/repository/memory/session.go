package memory

import (
	"context"
	"sync"

	"HaruhiServer/internal/domain"
)

type SessionRepository struct {
	mu          sync.RWMutex
	byID        map[domain.SessionID]*domain.Session
	tokenHashID map[string]domain.SessionID
}

func NewSessionRepository() *SessionRepository {
	return &SessionRepository{
		byID:        make(map[domain.SessionID]*domain.Session),
		tokenHashID: make(map[string]domain.SessionID),
	}
}

func (r *SessionRepository) Create(ctx context.Context, session *domain.Session) error {
	if err := checkContext(ctx); err != nil {
		return err
	}
	if session == nil {
		return domain.NewDomainError(domain.ErrInvalidArgument, "session is required")
	}
	if err := session.Validate(); err != nil {
		return err
	}

	key := indexKey(session.TokenHash)

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.byID[session.ID]; ok {
		return domain.NewDomainError(domain.ErrConflict, "session id already exists")
	}
	if _, ok := r.tokenHashID[key]; ok {
		return domain.NewDomainError(domain.ErrConflict, "session token hash already exists")
	}

	r.byID[session.ID] = cloneSession(session)
	r.tokenHashID[key] = session.ID
	return nil
}

func (r *SessionRepository) Update(ctx context.Context, session *domain.Session) error {
	if err := checkContext(ctx); err != nil {
		return err
	}
	if session == nil {
		return domain.NewDomainError(domain.ErrInvalidArgument, "session is required")
	}
	if err := session.Validate(); err != nil {
		return err
	}

	newKey := indexKey(session.TokenHash)

	r.mu.Lock()
	defer r.mu.Unlock()

	old, ok := r.byID[session.ID]
	if !ok {
		return domain.NewDomainError(domain.ErrNotFound, "session not found")
	}

	oldKey := indexKey(old.TokenHash)
	if oldKey != newKey {
		if ownerID, exists := r.tokenHashID[newKey]; exists && ownerID != session.ID {
			return domain.NewDomainError(domain.ErrConflict, "session token hash already exists")
		}
		delete(r.tokenHashID, oldKey)
		r.tokenHashID[newKey] = session.ID
	}

	r.byID[session.ID] = cloneSession(session)
	return nil
}

func (r *SessionRepository) GetByID(ctx context.Context, id domain.SessionID) (*domain.Session, error) {
	if err := checkContext(ctx); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	session, ok := r.byID[id]
	if !ok {
		return nil, domain.NewDomainError(domain.ErrNotFound, "session not found")
	}

	return cloneSession(session), nil
}

func (r *SessionRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*domain.Session, error) {
	if err := checkContext(ctx); err != nil {
		return nil, err
	}

	key := indexKey(tokenHash)

	r.mu.RLock()
	defer r.mu.RUnlock()

	id, ok := r.tokenHashID[key]
	if !ok {
		return nil, domain.NewDomainError(domain.ErrNotFound, "session not found")
	}

	session, ok := r.byID[id]
	if !ok {
		return nil, domain.NewDomainError(domain.ErrNotFound, "session not found")
	}

	return cloneSession(session), nil
}

func (r *SessionRepository) Delete(ctx context.Context, id domain.SessionID) error {
	if err := checkContext(ctx); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	session, ok := r.byID[id]
	if !ok {
		return domain.NewDomainError(domain.ErrNotFound, "session not found")
	}

	delete(r.byID, id)
	delete(r.tokenHashID, indexKey(session.TokenHash))
	return nil
}

func (r *SessionRepository) List(ctx context.Context) ([]*domain.Session, error) {
	if err := checkContext(ctx); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	items := make([]*domain.Session, 0, len(r.byID))
	for _, session := range r.byID {
		items = append(items, cloneSession(session))
	}

	sortSessions(items)
	return items, nil
}

func (r *SessionRepository) ListByUserID(ctx context.Context, userID domain.UserID) ([]*domain.Session, error) {
	if err := checkContext(ctx); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	items := make([]*domain.Session, 0)
	for _, session := range r.byID {
		if session.UserID == userID {
			items = append(items, cloneSession(session))
		}
	}

	sortSessions(items)
	return items, nil
}
