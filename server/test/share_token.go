package test

import (
	"testing"

	"testbert/protobuf/collection"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func testCreateShareToken(t *testing.T, tc *TestClient) {
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
		OrgShare:       true,
	}, &userOne, &orgOne)

	result := mustCreateShareToken(t, tc, one.CollectionId, &userOne, &orgOne)
	assert.Equal(t, one.CollectionId, result.CollectionId)

	_, err := tc.ShareCollection(one.CollectionId, &userTwo, &orgOne)
	assert.Error(t, err, "only creating user can share private collections")

	result = mustCreateShareToken(t, tc, two.CollectionId, &userTwo, &orgOne)
	assert.Equal(t, two.CollectionId, result.CollectionId)

	_, err = tc.ShareCollection(two.CollectionId, &userTwo, &orgTwo)
	assert.Error(t, err, "only same org users can share org collections")
}

func testGetSharedCollection(t *testing.T, tc *TestClient) {
	user := uuid.New()
	org := uuid.New()

	one := mustCreateCollection(t, tc, &collection.Collection{
		CollectionData: "one",
	}, &user, &org)

	two := mustCreateCollection(t, tc, &collection.Collection{
		CollectionData: "two",
		OrgView:        true,
		OrgEdit:        true,
		OrgShare:       true,
	}, &user, &org)

	oneShared := mustCreateShareToken(t, tc, one.CollectionId, &user, &org)
	twoShared := mustCreateShareToken(t, tc, two.CollectionId, &user, &org)

	result, err := tc.GetSharedCollection(oneShared.Token)
	assert.NoError(t, err)
	assert.Equal(t, one.CollectionData, result.CollectionData)

	result, err = tc.GetSharedCollection(twoShared.Token)
	assert.NoError(t, err)
	assert.Equal(t, two.CollectionData, result.CollectionData)
}

func testRateLimitedShareToken(t *testing.T, tc *TestClient) {
	user := uuid.New()
	org := uuid.New()

	one := mustCreateCollection(t, tc, &collection.Collection{
		CollectionData: "one",
	}, &user, &org)

	shared := mustCreateShareToken(t, tc, one.CollectionId, &user, &org)

	for range 250 {
		_, err := tc.GetSharedCollection(shared.Token)
		assert.NoError(t, err)
	}

	_, err := tc.GetSharedCollection(shared.Token)
	assert.Error(t, err)
}

func testDeleteSharedToken(t *testing.T, tc *TestClient) {
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
		OrgShare:       true,
	}, &userOne, &orgOne)

	tokenOne := mustCreateShareToken(t, tc, one.CollectionId, &userOne, &orgOne)
	tokenTwo := mustCreateShareToken(t, tc, two.CollectionId, &userOne, &orgOne)

	err := tc.DeleteShareToken(tokenOne.Token, &userTwo, &orgOne)
	assert.Error(t, err, "only creating user can revoke shared tokens on private collections")

	err = tc.DeleteShareToken(tokenOne.Token, &userOne, &orgOne)
	assert.NoError(t, err)

	err = tc.DeleteShareToken(tokenTwo.Token, &userTwo, &orgTwo)
	assert.Error(t, err, "only same org users can revoke shared tokens on org collections")

	err = tc.DeleteShareToken(tokenTwo.Token, &userTwo, &orgOne)
	assert.NoError(t, err)
}

func testDeleteCollection(t *testing.T, tc *TestClient) {
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
		OrgShare:       true,
	}, &userOne, &orgOne)

	tokenOne := mustCreateShareToken(t, tc, one.CollectionId, &userOne, &orgOne)

	err := tc.DeleteCollection(one.CollectionId, &userTwo, &orgOne)
	assert.Error(t, err, "only creating user can delete private collections")

	err = tc.DeleteCollection(one.CollectionId, &userOne, &orgOne)
	assert.NoError(t, err)

	_, err = tc.GetSharedCollection(tokenOne.Token)
	assert.Error(t, err)

	err = tc.DeleteCollection(two.CollectionId, &userTwo, &orgTwo)
	assert.Error(t, err, "only same org users can delete org collections")

	err = tc.DeleteCollection(two.CollectionId, &userTwo, &orgOne)
	assert.NoError(t, err)
}

func mustCreateShareToken(t *testing.T, tc *TestClient, id string, user, org *uuid.UUID) *collection.ShareToken {
	out, err := tc.ShareCollection(id, user, org)
	assert.NoError(t, err)
	return out
}
