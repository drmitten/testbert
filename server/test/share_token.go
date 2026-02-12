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

	tests := []struct {
		name       string
		user       *uuid.UUID
		org        *uuid.UUID
		collection string
		wantErr    bool
		errMsg     string
	}{
		{
			name:       "owner can share private collection",
			user:       &userOne,
			org:        &orgOne,
			collection: one.CollectionId,
			wantErr:    false,
		},
		{
			name:       "other user cannot share private collection",
			user:       &userTwo,
			org:        &orgOne,
			collection: one.CollectionId,
			wantErr:    true,
			errMsg:     "only creating user can share private collections",
		},
		{
			name:       "org member can share org-shareable collection",
			user:       &userTwo,
			org:        &orgOne,
			collection: two.CollectionId,
			wantErr:    false,
		},
		{
			name:       "user from different org cannot share org collection",
			user:       &userTwo,
			org:        &orgTwo,
			collection: two.CollectionId,
			wantErr:    true,
			errMsg:     "only same org users can share org collections",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tc.ShareCollection(tt.collection, tt.user, tt.org)
			if tt.wantErr {
				assert.Error(t, err, tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
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

	tests := []struct {
		name     string
		token    string
		wantErr  bool
		wantData string
	}{
		{
			name:     "access shared private collection",
			token:    oneShared.Token,
			wantErr:  false,
			wantData: one.CollectionData,
		},
		{
			name:     "access shared org collection",
			token:    twoShared.Token,
			wantErr:  false,
			wantData: two.CollectionData,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tc.GetSharedCollection(tt.token)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantData, result.CollectionData)
			}
		})
	}
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

	tests := []struct {
		name    string
		user    *uuid.UUID
		org     *uuid.UUID
		token   string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "owner can revoke token on private collection",
			user:    &userOne,
			org:     &orgOne,
			token:   tokenOne.Token,
			wantErr: false,
		},
		{
			name:    "other user cannot revoke token on private collection",
			user:    &userTwo,
			org:     &orgOne,
			token:   tokenOne.Token,
			wantErr: true,
			errMsg:  "only creating user can revoke shared tokens on private collections",
		},
		{
			name:    "user from different org cannot revoke token on org collection",
			user:    &userTwo,
			org:     &orgTwo,
			token:   tokenTwo.Token,
			wantErr: true,
			errMsg:  "only same org users can revoke shared tokens on org collections",
		},
		{
			name:    "org member can revoke token on org collection",
			user:    &userTwo,
			org:     &orgOne,
			token:   tokenTwo.Token,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tc.DeleteShareToken(tt.token, tt.user, tt.org)
			if tt.wantErr {
				assert.Error(t, err, tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func mustCreateShareToken(t *testing.T, tc *TestClient, id string, user, org *uuid.UUID) *collection.ShareToken {
	out, err := tc.ShareCollection(id, user, org)
	assert.NoError(t, err)
	return out
}
