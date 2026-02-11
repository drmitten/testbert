package memstore

import (
	"testbert/server/model"
	"testbert/server/tberrors"

	"github.com/google/uuid"
)

// CreateCollection implements [datastore.CollectionStore].
func (m *memStore) CreateCollection(c *model.Collection, user, org *uuid.UUID) (*model.Collection, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	c.ID = uuid.New()
	c.UserID = *user
	c.OrgID = *org
	m.collections[c.ID] = c
	return c, nil
}

// DeleteCollection implements [datastore.CollectionStore].
func (m *memStore) DeleteCollection(id, user, org *uuid.UUID) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	c, ok := m.collections[*id]
	if !ok {
		return tberrors.ErrCollectionNotFound
	}

	if c.UserID != *user {
		if !c.OrgEdit || c.OrgID != *org {
			return tberrors.ErrUnauthorized
		}
	}

	delete(m.collections, *id)

	// Cascade delete
	for k, t := range m.sharingTokens {
		if t.CollectionID == *id {
			delete(m.sharingTokens, k)
		}
	}
	return nil
}

// GetCollection implements [datastore.CollectionStore].
func (m *memStore) GetCollection(id, user, org *uuid.UUID) (*model.Collection, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	c, ok := m.collections[*id]
	if !ok {
		return nil, tberrors.ErrCollectionNotFound
	}

	if c.UserID != *user {
		if !c.OrgView || c.OrgID != *org {
			return nil, tberrors.ErrUnauthorized
		}
	}

	return c, nil
}

// GetCollectionFromSharingToken implements [datastore.CollectionStore].
func (m *memStore) GetCollectionFromSharingToken(token string) (*model.Collection, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	t, ok := m.sharingTokens[token]
	if !ok {
		return nil, tberrors.ErrTokenNotFound
	}
	c, ok := m.collections[t.CollectionID]
	if !ok {
		return nil, tberrors.ErrCollectionNotFound
	}
	return c, nil
}

// UpdateCollection implements [datastore.CollectionStore].
func (m *memStore) UpdateCollection(c *model.Collection, user, org *uuid.UUID) (*model.Collection, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	existing, ok := m.collections[c.ID]
	if !ok {
		return nil, tberrors.ErrCollectionNotFound
	}

	if existing.UserID != *user {
		if !existing.OrgEdit || existing.OrgID != *org {
			return nil, tberrors.ErrUnauthorized
		}
	}

	c.UserID = existing.UserID
	c.OrgID = existing.OrgID

	m.collections[existing.ID] = c
	return c, nil
}
