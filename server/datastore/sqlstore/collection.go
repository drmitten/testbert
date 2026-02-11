package sqlstore

import (
	"database/sql"
	"log"

	"testbert/server/model"
	"testbert/server/tberrors"

	"github.com/google/uuid"
)

// CreateCollection implements [datastore.TestBertDatastore].
func (s *sqlStore) CreateCollection(c *model.Collection, user *uuid.UUID, org *uuid.UUID) (*model.Collection, error) {
	c.ID = uuid.New()
	c.UserID = *user
	c.OrgID = *org

	query := `
	INSERT INTO collections(id, user_id, org_id, data, org_view, org_edit, org_share)
	VALUES(:id, :user_id, :org_id, :data, :org_view, :org_edit, :org_share);`
	_, err := s.db.NamedExec(query, c)
	if err != nil {
		log.Printf("database error: %v", err)
		return nil, tberrors.ErrInternal
	}

	return c, nil
}

// DeleteCollection implements [datastore.TestBertDatastore].
func (s *sqlStore) DeleteCollection(id *uuid.UUID, user *uuid.UUID, org *uuid.UUID) error {
	query := `
	DELETE FROM collections 
	WHERE id = $1
	  AND (user_id = $2 OR (org_edit AND org_id = $3));`

	result, err := s.db.Exec(query, id, user, org)
	if err != nil {
		log.Printf("database error: %v", err)
		return tberrors.ErrInternal
	}

	if count, _ := result.RowsAffected(); count == 0 {
		return tberrors.ErrCollectionNotFound
	}

	return nil
}

// GetCollection implements [datastore.TestBertDatastore].
func (s *sqlStore) GetCollection(id *uuid.UUID, user *uuid.UUID, org *uuid.UUID) (*model.Collection, error) {
	query := `
	SELECT id, user_id, org_id, data, org_view, org_edit, org_share
	FROM collections
	WHERE id = $1
	  AND (user_id = $2 OR (org_id = $3 AND org_view))`

	out := &model.Collection{}

	err := s.db.Get(out, query, id, user, org)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, tberrors.ErrCollectionNotFound
		} else {
			log.Printf("database error: %v", err)
			return nil, tberrors.ErrInternal
		}
	}

	return out, nil
}

// GetCollectionFromSharingToken implements [datastore.TestBertDatastore].
func (s *sqlStore) GetCollectionFromSharingToken(token string) (*model.Collection, error) {
	query := `
	SELECT c.id, c.user_id, c.org_id, c.data, c.org_view, c.org_edit, c.org_share
	FROM collections c
		JOIN shared_tokens t ON c.id = t.collection_id
	WHERE t.token = $1;`

	out := &model.Collection{}
	err := s.db.Get(out, query, token)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, tberrors.ErrCollectionNotFound
		} else {
			log.Printf("database error: %v", err)
			return nil, tberrors.ErrInternal
		}
	}

	return out, nil
}

// UpdateCollection implements [datastore.TestBertDatastore].
func (s *sqlStore) UpdateCollection(c *model.Collection, user *uuid.UUID, org *uuid.UUID) (*model.Collection, error) {
	existing, err := s.GetCollection(&c.ID, user, org)
	if err != nil {
		return nil, err
	}

	if existing.UserID != *user {
		if !existing.OrgEdit || existing.OrgID != *org {
			return nil, tberrors.ErrUnauthorized
		}
	}

	query := `
	UPDATE collections
	SET data = :data,
		org_view = :org_view,
		org_edit = :org_edit,
		org_share = :org_share
	WHERE id = :id
		AND (user_id = :user_id OR (org_id = :org_id AND org_edit));`

	_, err = s.db.NamedExec(query, c)
	if err != nil {
		log.Printf("database error: %v", err)
		return nil, tberrors.ErrInternal
	}

	return c, nil
}
