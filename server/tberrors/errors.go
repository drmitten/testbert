// Package tberrors TestBert Errors
package tberrors

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrCollectionNotFound = status.Error(codes.NotFound, "collection not found")
	ErrTokenNotFound      = status.Error(codes.NotFound, "sharing token not found")
	ErrUnauthorized       = status.Error(codes.Unauthenticated, "unauthorized")
	ErrInternal           = status.Error(codes.Internal, "internal error")
	ErrRateLimited        = status.Error(codes.ResourceExhausted, "rate limited")
)
