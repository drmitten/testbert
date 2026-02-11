package memstore

import (
	"testbert/server/model"
	"testbert/server/tberrors"

	"github.com/google/uuid"
)

// CreateSharingToken implements [datastore.SharingTokenStore].
func (m *memStore) CreateSharingToken(collectionID, user, org *uuid.UUID) (*model.SharingToken, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	c, ok := m.collections[*collectionID]
	if !ok {
		return nil, tberrors.ErrCollectionNotFound
	}

	if c.UserID != *user {
		if !c.OrgShare || c.OrgID != *org {
			return nil, tberrors.ErrUnauthorized
		}
	}

	t := &model.SharingToken{
		Token:        uuid.NewString(),
		CollectionID: *collectionID,
		UserID:       *user,
		OrgID:        *org,
	}

	m.sharingTokens[t.Token] = t
	return t, nil
}

// DeleteSharingToken implements [datastore.SharingTokenStore].
func (m *memStore) DeleteSharingToken(token string, user, org *uuid.UUID) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	t, ok := m.sharingTokens[token]
	if !ok {
		return tberrors.ErrTokenNotFound
	}

	c, ok := m.collections[t.CollectionID]
	if !ok {
		delete(m.sharingTokens, t.Token)
		return nil
	}

	if t.UserID != *user {
		if !c.OrgShare || t.OrgID != *org {
			return tberrors.ErrUnauthorized
		}
	}

	delete(m.sharingTokens, token)
	return nil
}
