// Package metrics Interceptor for simple metrics
package metrics

import (
	"context"
	"slices"
	"time"

	"testbert/server/tberrors"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	meter            = otel.Meter("testbert")
	requestCount     metric.Int64Counter
	requestLatency   metric.Float64Histogram
	errorCount       metric.Int64Counter
	RateLimitedCount metric.Int64Counter
	expectedErrors   = []error{
		tberrors.ErrUnauthorized,
		tberrors.ErrCollectionNotFound,
		tberrors.ErrTokenNotFound,
		tberrors.ErrRateLimited,
	}
)

func init() {
	var err error
	if requestCount, err = meter.Int64Counter(
		"grpc.server.requests_total",
		metric.WithDescription("Total number of gRPC requests"),
	); err != nil {
		panic(err)
	}

	if requestLatency, err = meter.Float64Histogram(
		"grpc.server.request_duration_seconds",
		metric.WithDescription("gRPC request duration in seconds"),
		metric.WithUnit("s"),
	); err != nil {
		panic(err)
	}

	if errorCount, err = meter.Int64Counter(
		"grpc.server.errors_total",
		metric.WithDescription("Total number of gRPC errors"),
	); err != nil {
		panic(err)
	}

	if RateLimitedCount, err = meter.Int64Counter(
		"grpc.server.rate_limited_total",
		metric.WithDescription("Total number of rate limited requests"),
	); err != nil {
		panic(err)
	}
}

func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		startTime := time.Now()

		attrs := []attribute.KeyValue{
			attribute.String("grpc.method", info.FullMethod),
		}

		resp, err := handler(ctx, req)

		duration := time.Since(startTime).Seconds()
		statusCode := codes.OK
		if err != nil {
			statusCode = status.Code(err)
		}

		attrs = append(attrs, attribute.String("grpc.status_code", statusCode.String()))

		requestCount.Add(ctx, 1, metric.WithAttributes(attrs...))
		requestLatency.Record(ctx, duration, metric.WithAttributes(attrs...))

		if err != nil && !slices.Contains(expectedErrors, err) {
			errorCount.Add(ctx, 1, metric.WithAttributes(attrs...))
		}

		return resp, err
	}
}
