package sqlstore

import (
	"log"

	"testbert/server/model"
	"testbert/server/tberrors"

	"github.com/google/uuid"
)

// CreateSharingToken implements [datastore.TestBertDatastore].
func (s *sqlStore) CreateSharingToken(collectionID *uuid.UUID, user *uuid.UUID, org *uuid.UUID) (*model.SharingToken, error) {
	existing, err := s.GetCollection(collectionID, user, org)
	if err != nil {
		return nil, err
	}

	if existing.UserID != *user {
		if !existing.OrgShare || existing.OrgID != *org {
			return nil, tberrors.ErrUnauthorized
		}
	}

	out := &model.SharingToken{
		Token:        uuid.NewString(),
		CollectionID: *collectionID,
		UserID:       *user,
		OrgID:        *org,
	}

	query := `
	INSERT INTO shared_tokens(token, collection_id, user_id, org_id)
	VALUES(:token, :collection_id, :user_id, :org_id);`

	_, err = s.db.NamedExec(query, out)
	if err != nil {
		log.Printf("database error: %v", err)
		return nil, tberrors.ErrInternal
	}

	return out, nil
}

// DeleteSharingToken implements [datastore.TestBertDatastore].
func (s *sqlStore) DeleteSharingToken(t string, user *uuid.UUID, org *uuid.UUID) error {
	query := `
	DELETE
	FROM shared_tokens t
	USING collections c
	WHERE t.token = $1
		AND t.collection_id = c.id
	  AND (c.user_id = $2 OR (c.org_id = $3 AND c.org_share));`

	result, err := s.db.Exec(query, t, user, org)
	if err != nil {
		log.Printf("database error: %v", err)
		return tberrors.ErrInternal
	}

	if count, _ := result.RowsAffected(); count == 0 {
		return tberrors.ErrTokenNotFound
	}

	return nil
}
