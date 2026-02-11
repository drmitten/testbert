// Package sqlstore TestBert Datastore backed by Postgresql
package sqlstore

import (
	"testbert/server/datastore"

	"github.com/jmoiron/sqlx"
)

type sqlStore struct {
	db *sqlx.DB
}

func NewSqlStore(db *sqlx.DB) datastore.TestBertDatastore {
	return &sqlStore{
		db: db,
	}
}
