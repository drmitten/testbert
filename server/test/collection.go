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

	tests := []struct {
		name    string
		user    *uuid.UUID
		org     *uuid.UUID
		input   *collection.Collection
		wantErr bool
		errMsg  string
	}{
		{
			name:    "authenticated user can create collection",
			user:    &user,
			org:     &org,
			input:   &collection.Collection{CollectionData: "foo"},
			wantErr: false,
		},
		{
			name:    "unauthenticated user cannot create collection",
			user:    nil,
			org:     nil,
			input:   &collection.Collection{CollectionData: "foo"},
			wantErr: true,
			errMsg:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				_, err := tc.client.CreateCollection(t.Context(), &collection.CreateCollectionRequest{
					CollectionData: tt.input.CollectionData,
				})
				assert.Error(t, err, "Unauthenticated users should not be able to create a collection")
			} else {
				mustCreateCollection(t, tc, tt.input, tt.user, tt.org)
			}
		})
	}
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

	tests := []struct {
		name       string
		user       *uuid.UUID
		org        *uuid.UUID
		collection string
		wantErr    bool
		errMsg     string
	}{
		{
			name:       "owner can access private collection",
			user:       &userOne,
			org:        &orgOne,
			collection: one.CollectionId,
			wantErr:    false,
		},
		{
			name:       "other user cannot access private collection",
			user:       &userTwo,
			org:        &orgOne,
			collection: one.CollectionId,
			wantErr:    true,
			errMsg:     "only creating user can access private collections",
		},
		{
			name:       "org member can access org-viewable collection",
			user:       &userTwo,
			org:        &orgOne,
			collection: two.CollectionId,
			wantErr:    false,
		},
		{
			name:       "user from different org cannot access org collection",
			user:       &userTwo,
			org:        &orgTwo,
			collection: one.CollectionId,
			wantErr:    true,
			errMsg:     "only users in the same org can see org collections",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tc.GetCollection(tt.user, tt.org, tt.collection)
			if tt.wantErr {
				assert.Error(t, err, tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
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

	tests := []struct {
		name    string
		user    *uuid.UUID
		org     *uuid.UUID
		input   *collection.Collection
		wantErr bool
		errMsg  string
	}{
		{
			name:    "owner can update private collection",
			user:    &userOne,
			org:     &orgOne,
			input:   &collection.Collection{CollectionId: one.CollectionId, CollectionData: "one more"},
			wantErr: false,
		},
		{
			name:    "other user cannot update private collection",
			user:    &userTwo,
			org:     &orgOne,
			input:   &collection.Collection{CollectionId: one.CollectionId, CollectionData: "one more"},
			wantErr: true,
			errMsg:  "only creating user can update private collections",
		},
		{
			name:    "org member can update org-editable collection",
			user:    &userTwo,
			org:     &orgOne,
			input:   &collection.Collection{CollectionId: two.CollectionId, CollectionData: "two more"},
			wantErr: false,
		},
		{
			name:    "user from different org cannot update org collection",
			user:    &userTwo,
			org:     &orgTwo,
			input:   &collection.Collection{CollectionId: two.CollectionId, CollectionData: "two more"},
			wantErr: true,
			errMsg:  "only same org can update org collections",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tc.UpdateCollection(tt.input, tt.user, tt.org)
			if tt.wantErr {
				assert.Error(t, err, tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
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

	tests := []struct {
		name         string
		user         *uuid.UUID
		org          *uuid.UUID
		collectionID string
		wantErr      bool
		errMsg       string
	}{
		{
			name:         "owner can delete private collection",
			user:         &userOne,
			org:          &orgOne,
			collectionID: one.CollectionId,
			wantErr:      false,
		},
		{
			name:         "other user cannot delete private collection",
			user:         &userTwo,
			org:          &orgOne,
			collectionID: one.CollectionId,
			wantErr:      true,
			errMsg:       "only creating user can delete private collections",
		},
		{
			name:         "user from different org cannot delete org collection",
			user:         &userTwo,
			org:          &orgTwo,
			collectionID: two.CollectionId,
			wantErr:      true,
			errMsg:       "only same org users can delete org collections",
		},
		{
			name:         "org member can delete org-editable collection",
			user:         &userTwo,
			org:          &orgOne,
			collectionID: two.CollectionId,
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tc.DeleteCollection(tt.collectionID, tt.user, tt.org)
			if tt.wantErr {
				assert.Error(t, err, tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}

	_, err := tc.GetSharedCollection(tokenOne.Token)
	assert.Error(t, err)
}

func mustCreateCollection(t *testing.T, tc *TestClient, in *collection.Collection, user, org *uuid.UUID) *collection.Collection {
	out, err := tc.CreateCollection(in, user, org)
	assert.NoError(t, err)
	return out
}
