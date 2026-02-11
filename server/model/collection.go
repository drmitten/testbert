// Package model TestBert Data Models
package model

import "github.com/google/uuid"

type Collection struct {
	ID       uuid.UUID `db:"id"`
	UserID   uuid.UUID `db:"user_id"`
	OrgID    uuid.UUID `db:"org_id"`
	Data     string    `db:"data"`
	OrgView  bool      `db:"org_view"`
	OrgEdit  bool      `db:"org_edit"`
	OrgShare bool      `db:"org_share"`
}
