package test

import (
	"testing"

	"testbert/protobuf/collection"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func testCreateCollection(t *testing.T, tc *TestClient) {
	user := uuid.New()
	org := uuid.New()

	// Authenticated users should be able to create a collection
	mustCreateCollection(t, tc, &collection.Collection{
		CollectionData: "foo",
	}, &user, &org)

	_, err := tc.client.CreateCollection(t.Context(), &collection.CreateCollectionRequest{
		CollectionData: "foo",
	})
	assert.Error(t, err, "Unauthenticated users should not be able to create a collection")
}

func testGetCollection(t *testing.T, tc *TestClient) {
	userOne := uuid.New()
	userTwo := uuid.New()
	orgOne := uuid.New()
	orgTwo := uuid.New()

	one := mustCreateCollection(t, tc, &collection.Collection{
		CollectionData: "one",
	}, &userOne, &orgOne)

	two := mustCreateCollection(t, tc, &collection.Collection{
		CollectionData: "two",
		OrgView:        true,
	}, &userOne, &orgOne)

	result := mustGetCollection(t, tc, one.CollectionId, &userOne, &orgOne)
	assert.Equal(t, one.CollectionData, result.CollectionData)

	_, err := tc.GetCollection(&userTwo, &orgOne, one.CollectionId)
	assert.Error(t, err, "only creating user can access private collections")

	// Other users in the same org should be able to see org viewable collections
	mustGetCollection(t, tc, two.CollectionId, &userTwo, &orgOne)

	_, err = tc.GetCollection(&userTwo, &orgTwo, one.CollectionId)
	assert.Error(t, err, "only users in the same org can see org collections")
}

func testUpdateCollection(t *testing.T, tc *TestClient) {
	userOne := uuid.New()
	userTwo := uuid.New()
	orgOne := uuid.New()
	orgTwo := uuid.New()

	one := mustCreateCollection(t, tc, &collection.Collection{
		CollectionData: "one",
	}, &userOne, &orgOne)

	two := mustCreateCollection(t, tc, &collection.Collection{
		CollectionData: "two",
		OrgView:        true,
		OrgEdit:        true,
	}, &userOne, &orgOne)

	oneUpdate := &collection.Collection{
		CollectionId:   one.CollectionId,
		CollectionData: "one more",
	}

	twoUpdate := &collection.Collection{
		CollectionId:   two.CollectionId,
		CollectionData: "two more",
	}

	_, err := tc.UpdateCollection(oneUpdate, &userTwo, &orgOne)
	assert.Error(t, err, "only creating user can update private collections")

	result := mustUpdateCollection(t, tc, oneUpdate, &userOne, &orgOne)
	assert.Equal(t, oneUpdate.CollectionData, result.CollectionData)

	_, err = tc.UpdateCollection(twoUpdate, &userTwo, &orgTwo)
	assert.Error(t, err, "only same org can update org collections")

	result = mustUpdateCollection(t, tc, twoUpdate, &userTwo, &orgOne)
	assert.Equal(t, twoUpdate.CollectionData, result.CollectionData)
}

func mustCreateCollection(t *testing.T, tc *TestClient, in *collection.Collection, user, org *uuid.UUID) *collection.Collection {
	out, err := tc.CreateCollection(in, user, org)
	assert.NoError(t, err)
	return out
}

func mustGetCollection(t *testing.T, tc *TestClient, id string, user, org *uuid.UUID) *collection.Collection {
	out, err := tc.GetCollection(user, org, id)
	assert.NoError(t, err)
	return out
}

func mustUpdateCollection(t *testing.T, tc *TestClient, in *collection.Collection, user, org *uuid.UUID) *collection.Collection {
	out, err := tc.UpdateCollection(in, user, org)
	assert.NoError(t, err)
	return out
}
