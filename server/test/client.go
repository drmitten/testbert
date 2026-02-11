// Package test TestBert Tests ;)
package test

import (
	"context"
	"time"

	"testbert/protobuf/collection"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"google.golang.org/grpc"
)

type TestClient struct {
	client collection.CollectionServiceClient
	key    []byte
}

func NewClient(csc collection.CollectionServiceClient, secret string) *TestClient {
	return &TestClient{
		client: csc,
		key:    []byte(secret),
	}
}

func (tc *TestClient) CreateCollection(in *collection.Collection, user, org *uuid.UUID) (*collection.Collection, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	out, err := tc.client.CreateCollection(ctx, &collection.CreateCollectionRequest{
		CollectionData: in.CollectionData,
		OrgView:        in.OrgView,
		OrgEdit:        in.OrgEdit,
		OrgShare:       in.OrgShare,
	}, tc.withCredentials(user, org))
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (tc *TestClient) GetCollection(user, org *uuid.UUID, id string) (*collection.Collection, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	out, err := tc.client.GetCollection(ctx, &collection.GetCollectionRequest{
		CollectionId: id,
	}, tc.withCredentials(user, org))
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (tc *TestClient) UpdateCollection(in *collection.Collection, user, org *uuid.UUID) (*collection.Collection, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	out, err := tc.client.UpdateCollection(ctx, &collection.UpdateCollectionRequest{
		CollectionId:   in.CollectionId,
		CollectionData: in.CollectionData,
		OrgView:        in.OrgView,
		OrgEdit:        in.OrgEdit,
		OrgShare:       in.OrgShare,
	}, tc.withCredentials(user, org))
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (tc *TestClient) ShareCollection(id string, user, org *uuid.UUID) (*collection.ShareToken, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	out, err := tc.client.CreateShareToken(ctx, &collection.CreateShareTokenRequest{
		CollectionId: id,
	}, tc.withCredentials(user, org))
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (tc *TestClient) GetSharedCollection(id string) (*collection.Collection, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	out, err := tc.client.GetSharedCollection(ctx, &collection.GetSharedCollectionRequest{
		Token: id,
	})
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (tc *TestClient) DeleteShareToken(id string, user, org *uuid.UUID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := tc.client.RevokeShareToken(ctx, &collection.RevokeShareTokenRequest{
		Token: id,
	}, tc.withCredentials(user, org))

	return err
}

func (tc *TestClient) DeleteCollection(id string, user, org *uuid.UUID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := tc.client.DeleteCollection(ctx, &collection.DeleteCollectionRequest{
		CollectionId: id,
	}, tc.withCredentials(user, org))

	return err
}

func (tc *TestClient) withCredentials(user, org *uuid.UUID) grpc.CallOption {
	return grpc.PerRPCCredentials(&credentials{
		key:  tc.key,
		user: user,
		org:  org,
	})
}

type credentials struct {
	key  []byte
	user *uuid.UUID
	org  *uuid.UUID
}

func (creds *credentials) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iss":  "testbert-authenticator",
		"user": creds.user.String(),
		"org":  creds.org.String(),
	}).SignedString(creds.key)
	if err != nil {
		return nil, err
	}

	return map[string]string{
		"authorization": "Bearer " + token,
	}, nil
}

func (creds *credentials) RequireTransportSecurity() bool {
	return false
}
