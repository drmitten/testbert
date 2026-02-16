package main

import (
	"context"
	"embed"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"

	"testbert/protobuf/collection"
	"testbert/server/config"
	"testbert/server/datastore/sqlstore"
	"testbert/server/interceptors/auth"
	"testbert/server/server"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"

	_ "github.com/joho/godotenv/autoload"
)

//go:embed migrations/*.sql
var migrations embed.FS

func main() {
	cfg := config.NewConfig()

	db, err := setupDB(cfg)
	if err != nil {
		log.Fatalf("error setting up DB: %v", err)
	}
	defer db.Close()

	srv := grpc.NewServer(grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.UnaryInterceptor(auth.Interceptor(cfg)))

	collection.RegisterCollectionServiceServer(srv, server.NewCollectionServer(sqlstore.NewSqlStore(db), cfg.AuthSecret))

	err = listenAndServe(cfg, srv)
	if err != nil {
		log.Fatalf("server error: %v", err)
	}
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
