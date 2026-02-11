// Package memstore In Memory TestBert Datastore
package memstore

import (
	"sync"

	"testbert/server/datastore"
	"testbert/server/model"

	"github.com/google/uuid"
)

type memStore struct {
	collections   map[uuid.UUID]*model.Collection
	sharingTokens map[string]*model.SharingToken
	lock          sync.Mutex
}

func NewMemStore() datastore.TestBertDatastore {
	return &memStore{
		collections:   map[uuid.UUID]*model.Collection{},
		sharingTokens: map[string]*model.SharingToken{},
		lock:          sync.Mutex{},
	}
}
