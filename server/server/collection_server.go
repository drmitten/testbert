// Package server TestBert Collection Server
package server

import (
	"context"
	"strings"
	"time"

	"testbert/protobuf/collection"
	"testbert/server/datastore"
	"testbert/server/model"
	"testbert/server/presenters"
	"testbert/server/tberrors"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/maypok86/otter/v2"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
)

type collectionServer struct {
	collection.UnimplementedCollectionServiceServer
	store   datastore.TestBertDatastore
	key     []byte
	cache   *otter.Cache[string, int]
	publish chan any
}

// CreateCollection implements [collection.CollectionServiceServer].
func (s *collectionServer) CreateCollection(ctx context.Context, req *collection.CreateCollectionRequest) (*collection.Collection, error) {
	user, org, err := s.getLoggedInUserAndOrg(ctx)
	if err != nil {
		return nil, err
	}

	out, err := s.store.CreateCollection(&model.Collection{
		ID:       uuid.New(),
		Data:     req.CollectionData,
		OrgView:  req.OrgView,
		OrgEdit:  req.OrgEdit,
		OrgShare: req.OrgShare,
	}, user, org)
	if err != nil {
		return nil, err
	}

	s.publish <- &AccessEvent{
		CollectionID: string(out.ID.String()),
		User:         user,
		OrgID:        org,
		Action:       "create",
	}
	return presenters.Collection(out), nil
}

// CreateShareToken implements [collection.CollectionServiceServer].
func (s *collectionServer) CreateShareToken(ctx context.Context, req *collection.CreateShareTokenRequest) (*collection.ShareToken, error) {
	user, org, err := s.getLoggedInUserAndOrg(ctx)
	if err != nil {
		return nil, err
	}

	id, err := uuid.Parse(req.CollectionId)
	if err != nil {
		return nil, tberrors.ErrCollectionNotFound
	}

	out, err := s.store.CreateSharingToken(&id, user, org)
	if err != nil {
		return nil, err
	}

	s.publish <- &AccessEvent{
		SharedToken:  out.Token,
		CollectionID: string(id.String()),
		User:         user,
		OrgID:        org,
		Action:       "share",
	}
	return presenters.SharingToken(out), nil
}

// DeleteCollection implements [collection.CollectionServiceServer].
func (s *collectionServer) DeleteCollection(ctx context.Context, req *collection.DeleteCollectionRequest) (*emptypb.Empty, error) {
	user, org, err := s.getLoggedInUserAndOrg(ctx)
	if err != nil {
		return nil, err
	}

	id, err := uuid.Parse(req.CollectionId)
	if err != nil {
		return nil, tberrors.ErrCollectionNotFound
	}

	err = s.store.DeleteCollection(&id, user, org)
	if err != nil {
		return nil, err
	}

	s.publish <- &AccessEvent{
		CollectionID: string(id.String()),
		User:         user,
		OrgID:        org,
		Action:       "delete",
	}
	return &emptypb.Empty{}, nil
}

// GetCollection implements [collection.CollectionServiceServer].
func (s *collectionServer) GetCollection(ctx context.Context, req *collection.GetCollectionRequest) (*collection.Collection, error) {
	user, org, err := s.getLoggedInUserAndOrg(ctx)
	if err != nil {
		return nil, err
	}

	id, err := uuid.Parse(req.CollectionId)
	if err != nil {
		return nil, tberrors.ErrCollectionNotFound
	}

	out, err := s.store.GetCollection(&id, user, org)
	if err != nil {
		return nil, err
	}

	s.publish <- &AccessEvent{
		CollectionID: string(id.String()),
		User:         user,
		OrgID:        org,
		Action:       "read",
	}
	return presenters.Collection(out), nil
}

// GetSharedCollection implements [collection.CollectionServiceServer].
func (s *collectionServer) GetSharedCollection(ctx context.Context, req *collection.GetSharedCollectionRequest) (*collection.Collection, error) {
	count, _ := s.cache.Compute(req.Token, func(oldValue int, _ bool) (int, otter.ComputeOp) {
		if oldValue <= 250 {
			return oldValue + 1, otter.WriteOp
		} else {
			return oldValue, otter.CancelOp
		}
	})

	if count > 250 {
		return nil, tberrors.ErrRateLimited
	}

	out, err := s.store.GetCollectionFromSharingToken(req.Token)
	if err != nil {
		return nil, err
	}

	s.publish <- &AccessEvent{
		SharedToken:  req.Token,
		CollectionID: out.ID.String(),
		User:         nil,
		OrgID:        nil,
		Action:       "readShared",
	}
	return presenters.Collection(out), nil
}

// RevokeShareToken implements [collection.CollectionServiceServer].
func (s *collectionServer) RevokeShareToken(ctx context.Context, req *collection.RevokeShareTokenRequest) (*emptypb.Empty, error) {
	user, org, err := s.getLoggedInUserAndOrg(ctx)
	if err != nil {
		return nil, err
	}

	err = s.store.DeleteSharingToken(req.Token, user, org)
	if err != nil {
		return nil, err
	}

	s.publish <- &AccessEvent{
		SharedToken:  req.Token,
		CollectionID: "",
		User:         user,
		OrgID:        org,
		Action:       "revoke",
	}
	return &emptypb.Empty{}, nil
}

func (s *collectionServer) UpdateCollection(ctx context.Context, req *collection.UpdateCollectionRequest) (*collection.Collection, error) {
	user, org, err := s.getLoggedInUserAndOrg(ctx)
	if err != nil {
		return nil, err
	}

	id, err := uuid.Parse(req.CollectionId)
	if err != nil {
		return nil, tberrors.ErrCollectionNotFound
	}

	out, err := s.store.UpdateCollection(&model.Collection{
		ID:       id,
		Data:     req.CollectionData,
		OrgView:  req.OrgView,
		OrgEdit:  req.OrgEdit,
		OrgShare: req.OrgShare,
	}, user, org)
	if err != nil {
		return nil, err
	}

	s.publish <- &AccessEvent{
		CollectionID: string(id.String()),
		User:         user,
		OrgID:        org,
		Action:       "update",
	}
	return presenters.Collection(out), nil
}

func NewCollectionServer(store datastore.TestBertDatastore, key string) collection.CollectionServiceServer {
	publish := make(chan any, 1000)
	defer close(publish)

	go func() {
		for msg := range publish {
			// publish the message to a queue for counts, dashboards, and analysis
			_ = msg
		}
	}()

	return &collectionServer{
		store: store,
		key:   []byte(key),
		cache: otter.Must(&otter.Options[string, int]{
			ExpiryCalculator: otter.ExpiryCreating[string, int](15 * time.Second),
		}),
		publish: publish,
	}
}

func (s *collectionServer) getLoggedInUserAndOrg(ctx context.Context) (user, org *uuid.UUID, err error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, nil, tberrors.ErrUnauthorized
	}

	auth := md["authorization"]
	if len(auth) < 1 {
		return nil, nil, tberrors.ErrUnauthorized
	}

	tokenString := strings.TrimPrefix(auth[0], "Bearer ")

	token, err := jwt.Parse(tokenString, s.keyFunc)
	if err != nil {
		return nil, nil, tberrors.ErrUnauthorized
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, nil, tberrors.ErrUnauthorized
	}

	userID := uuid.NullUUID{}
	_ = userID.Scan(claims["user"])
	if !userID.Valid {
		return nil, nil, tberrors.ErrUnauthorized
	}

	orgID := uuid.NullUUID{}
	_ = orgID.Scan(claims["org"])
	if !orgID.Valid {
		return nil, nil, tberrors.ErrUnauthorized
	}

	return &userID.UUID, &orgID.UUID, nil
}

func (s *collectionServer) keyFunc(_ *jwt.Token) (any, error) {
	return s.key, nil
}

type AccessEvent struct {
	SharedToken  string
	CollectionID string
	User         *uuid.UUID
	OrgID        *uuid.UUID
	Action       string
}
