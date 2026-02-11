// Package datastore Interface for datastore interaction
package datastore

// TestBertDatastore Implementation is responsible for enforcing auth logic
type TestBertDatastore interface {
	CollectionStore
	SharingTokenStore
}
