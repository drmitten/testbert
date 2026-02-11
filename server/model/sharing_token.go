package model

import (
	"github.com/google/uuid"
)

type SharingToken struct {
	UserID       uuid.UUID `db:"user_id"`
	OrgID        uuid.UUID `db:"org_id"`
	CollectionID uuid.UUID `db:"collection_id"`
	Token        string    `db:"token"`
}
