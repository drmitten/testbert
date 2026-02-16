package main

import (
	"context"
	"embed"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"time"

	"testbert/protobuf/collection"
	"testbert/server/config"
	"testbert/server/datastore/sqlstore"
	"testbert/server/interceptors/auth"
	"testbert/server/interceptors/metrics"
	"testbert/server/server"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"google.golang.org/grpc"

	_ "github.com/joho/godotenv/autoload"
)

//go:embed migrations/*.sql
var migrations embed.FS

func main() {
	cfg := config.NewConfig()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tracerProvider, err := newTracerProvider(ctx, cfg)
	if err != nil {
		log.Fatalf("failed to create tracer provider: %v", err)
	}
	otel.SetTracerProvider(tracerProvider)
	defer func() {
		if err := tracerProvider.Shutdown(context.Background()); err != nil {
			log.Printf("error shutting down tracer provider: %v", err)
		}
	}()

	db, err := setupDB(cfg)
	if err != nil {
		log.Fatalf("error setting up DB: %v", err)
	}
	defer db.Close()

	srv := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(
			metrics.UnaryServerInterceptor(),
			auth.Interceptor(cfg),
		))

	collection.RegisterCollectionServiceServer(srv, server.NewCollectionServer(sqlstore.NewSQLStore(db), cfg.AuthSecret))

	err = listenAndServe(cfg, srv)
	if err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func newTracerProvider(ctx context.Context, cfg *config.Configuration) (*sdktrace.TracerProvider, error) {
	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(fmt.Sprintf("%s:%s", cfg.OtlpEndpoint, cfg.OtlpPort)),
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithTimeout(5*time.Second),
	)
	if err != nil {
		return nil, fmt.Errorf("creating OTLP exporter: %w", err)
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(cfg.OtelServiceName),
			attribute.String("environment", cfg.OtelEnvironment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("creating resource: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)
	return tp, nil
}

func setupDB(cfg *config.Configuration) (*sqlx.DB, error) {
	dbConn := fmt.Sprintf("host=%v port=%v dbname=%v user=%v password=%v sslmode=disable",
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
		cfg.DBUser,
		cfg.DBPassword,
	)

	db, err := sqlx.Open("postgres", dbConn)
	if err != nil {
		return nil, err
	}

	goose.SetBaseFS(migrations)

	_ = goose.SetDialect(string(goose.DialectPostgres))

	err = goose.Up(db.DB, "migrations")
	if err != nil {
		return nil, err
	}

	return db, nil
}

func listenAndServe(cfg *config.Configuration, srv *grpc.Server) error {
	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", cfg.ServerPort))
	if err != nil {
		return err
	}

	log.Println("listening for connections ...")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	srvError := make(chan error, 1)
	go func() {
		srvError <- srv.Serve(listener)
	}()

	select {
	case err = <-srvError:
		return err
	case <-ctx.Done():
		stop()
		log.Println("stopping server ...")
	}

	srv.Stop()

	return nil
}
