package datastore

import (
	"testbert/server/model"

	"github.com/google/uuid"
)

type CollectionStore interface {
	CreateCollection(c *model.Collection, user, org *uuid.UUID) (*model.Collection, error)
	GetCollection(id, user, org *uuid.UUID) (*model.Collection, error)
	GetCollectionFromSharingToken(token string) (*model.Collection, error)
	UpdateCollection(c *model.Collection, user, org *uuid.UUID) (*model.Collection, error)
	DeleteCollection(id, user, org *uuid.UUID) error
}
