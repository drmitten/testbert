// Package server TestBert Collection Server
package server

import (
	"context"
	"time"

	"testbert/protobuf/collection"
	"testbert/server/config"
	"testbert/server/datastore"
	"testbert/server/interceptors/metrics"
	"testbert/server/model"
	"testbert/server/presenters"
	"testbert/server/tberrors"

	"github.com/google/uuid"
	"github.com/maypok86/otter/v2"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/types/known/emptypb"
)

var tracer = otel.Tracer("collection-server")

type collectionServer struct {
	collection.UnimplementedCollectionServiceServer
	store   datastore.TestBertDatastore
	cache   *otter.Cache[string, int]
	publish chan any
}

// CreateCollection implements [collection.CollectionServiceServer].
func (s *collectionServer) CreateCollection(ctx context.Context, req *collection.CreateCollectionRequest) (*collection.Collection, error) {
	ctx, span := tracer.Start(ctx, "CreateCollection",
		trace.WithAttributes(
			attribute.Bool("collection.create", true),
		))
	defer span.End()

	user, org, err := getLoggedInUserAndOrg(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "unauthorized")
		return nil, err
	}

	span.SetAttributes(
		attribute.String("user.id", user.String()),
		attribute.String("org.id", org.String()),
	)

	out, err := s.store.CreateCollection(&model.Collection{
		ID:       uuid.New(),
		Data:     req.CollectionData,
		OrgView:  req.OrgView,
		OrgEdit:  req.OrgEdit,
		OrgShare: req.OrgShare,
	}, user, org)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "create collection failed")
		return nil, err
	}

	span.SetAttributes(attribute.String("collection.id", out.ID.String()))

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
	ctx, span := tracer.Start(ctx, "CreateShareToken",
		trace.WithAttributes(
			attribute.String("collection.id", req.CollectionId),
		))
	defer span.End()

	user, org, err := getLoggedInUserAndOrg(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "unauthorized")
		return nil, err
	}

	span.SetAttributes(
		attribute.String("user.id", user.String()),
		attribute.String("org.id", org.String()),
	)

	id, err := uuid.Parse(req.CollectionId)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "invalid collection id")
		return nil, tberrors.ErrCollectionNotFound
	}

	out, err := s.store.CreateSharingToken(&id, user, org)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "create share token failed")
		return nil, err
	}

	span.SetAttributes(attribute.String("share.token", out.Token))

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
	ctx, span := tracer.Start(ctx, "DeleteCollection",
		trace.WithAttributes(
			attribute.String("collection.id", req.CollectionId),
		))
	defer span.End()

	user, org, err := getLoggedInUserAndOrg(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "unauthorized")
		return nil, err
	}

	span.SetAttributes(
		attribute.String("user.id", user.String()),
		attribute.String("org.id", org.String()),
	)

	id, err := uuid.Parse(req.CollectionId)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "invalid collection id")
		return nil, tberrors.ErrCollectionNotFound
	}

	err = s.store.DeleteCollection(&id, user, org)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "delete collection failed")
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
	ctx, span := tracer.Start(ctx, "GetCollection",
		trace.WithAttributes(
			attribute.String("collection.id", req.CollectionId),
		))
	defer span.End()

	user, org, err := getLoggedInUserAndOrg(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "unauthorized")
		return nil, err
	}

	span.SetAttributes(
		attribute.String("user.id", user.String()),
		attribute.String("org.id", org.String()),
	)

	id, err := uuid.Parse(req.CollectionId)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "invalid collection id")
		return nil, tberrors.ErrCollectionNotFound
	}

	out, err := s.store.GetCollection(&id, user, org)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "get collection failed")
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
	ctx, span := tracer.Start(ctx, "GetSharedCollection",
		trace.WithAttributes(
			attribute.String("share.token", req.Token),
		))
	defer span.End()

	count, _ := s.cache.Compute(req.Token, func(oldValue int, _ bool) (int, otter.ComputeOp) {
		if oldValue <= 250 {
			return oldValue + 1, otter.WriteOp
		} else {
			return oldValue, otter.CancelOp
		}
	})

	if count > 250 {
		span.SetAttributes(attribute.Bool("rate.limited", true))
		span.SetStatus(codes.Error, "rate limited")
		metrics.RateLimitedCount.Add(ctx, 1)
		return nil, tberrors.ErrRateLimited
	}

	out, err := s.store.GetCollectionFromSharingToken(req.Token)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "get shared collection failed")
		return nil, err
	}

	span.SetAttributes(attribute.String("collection.id", out.ID.String()))

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
	ctx, span := tracer.Start(ctx, "RevokeShareToken",
		trace.WithAttributes(
			attribute.String("share.token", req.Token),
		))
	defer span.End()

	user, org, err := getLoggedInUserAndOrg(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "unauthorized")
		return nil, err
	}

	span.SetAttributes(
		attribute.String("user.id", user.String()),
		attribute.String("org.id", org.String()),
	)

	err = s.store.DeleteSharingToken(req.Token, user, org)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "revoke share token failed")
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
	ctx, span := tracer.Start(ctx, "UpdateCollection",
		trace.WithAttributes(
			attribute.String("collection.id", req.CollectionId),
		))
	defer span.End()

	user, org, err := getLoggedInUserAndOrg(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "unauthorized")
		return nil, err
	}

	span.SetAttributes(
		attribute.String("user.id", user.String()),
		attribute.String("org.id", org.String()),
	)

	id, err := uuid.Parse(req.CollectionId)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "invalid collection id")
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
		span.RecordError(err)
		span.SetStatus(codes.Error, "update collection failed")
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

	go func() {
		for msg := range publish {
			// publish the message to a queue for counts, dashboards, and analysis
			_ = msg
		}
	}()

	return &collectionServer{
		store: store,
		cache: otter.Must(&otter.Options[string, int]{
			ExpiryCalculator: otter.ExpiryCreating[string, int](15 * time.Second),
		}),
		publish: publish,
	}
}

func getLoggedInUserAndOrg(ctx context.Context) (*uuid.UUID, *uuid.UUID, error) {
	user, ok := ctx.Value(config.KeyUserID).(*uuid.UUID)
	if !ok {
		return nil, nil, tberrors.ErrUnauthorized
	}
	org, ok := ctx.Value(config.KeyOrgID).(*uuid.UUID)
	if !ok {
		return nil, nil, tberrors.ErrUnauthorized
	}
	return user, org, nil
}

type AccessEvent struct {
	SharedToken  string
	CollectionID string
	User         *uuid.UUID
	OrgID        *uuid.UUID
	Action       string
}
