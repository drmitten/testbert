package datastore

import (
	"testbert/server/model"

	"github.com/google/uuid"
)

type SharingTokenStore interface {
	CreateSharingToken(collectionID, user, org *uuid.UUID) (*model.SharingToken, error)
	DeleteSharingToken(t string, user, org *uuid.UUID) error
}
