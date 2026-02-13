// Package auth interceptor for ensuring user authenticated
package auth

import (
	"context"
	"strings"

	"testbert/server/config"
	"testbert/server/tberrors"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func Interceptor(cfg *config.Configuration) grpc.UnaryServerInterceptor {
	key := []byte(cfg.AuthSecret)
	keyFunc := func(_ *jwt.Token) (any, error) {
		return key, nil
	}

	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		switch info.FullMethod {
		case "/collection.CollectionService/GetSharedCollection":
			// No auth required
			return handler(ctx, req)
		default:
			ctx, ok := loadUserAndOrgOnContext(ctx, keyFunc)
			if !ok {
				return nil, tberrors.ErrUnauthorized
			}

			return handler(ctx, req)
		}
	}
}

func loadUserAndOrgOnContext(ctx context.Context, keyFunc func(*jwt.Token) (any, error)) (context.Context, bool) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, false
	}

	auth := md["authorization"]
	if len(auth) < 1 {
		return nil, false
	}

	tokenString := strings.TrimPrefix(auth[0], "Bearer ")

	token, err := jwt.Parse(tokenString, keyFunc)
	if err != nil {
		return nil, false
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, false
	}

	userID := uuid.NullUUID{}
	_ = userID.Scan(claims["user"])
	if !userID.Valid {
		return nil, false
	}

	orgID := uuid.NullUUID{}
	_ = orgID.Scan(claims["org"])
	if !orgID.Valid {
		return nil, false
	}

	ctx = context.WithValue(ctx, config.KeyUserID, &userID.UUID)
	ctx = context.WithValue(ctx, config.KeyOrgID, &orgID.UUID)

	return ctx, true
}
