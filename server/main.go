package main

import (
	"embed"
	"fmt"
	"log"
	"net"

	"testbert/protobuf/collection"
	"testbert/server/config"
	"testbert/server/datastore/sqlstore"
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

	dbConn := fmt.Sprintf("host=%v port=%v dbname=%v user=%v password=%v sslmode=disable",
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
		cfg.DBUser,
		cfg.DBPassword,
	)

	db, err := sqlx.Open("postgres", dbConn)
	if err != nil {
		log.Fatalf("unable to connect to to database: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	goose.SetBaseFS(migrations)

	err = goose.SetDialect("postgres")
	if err != nil {
		log.Fatalf("unable to migrate database: %v", err)
	}

	err = goose.Up(db.DB, "migrations")
	if err != nil {
		log.Fatalf("unable to migrate database: %v", err)
	}

	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", cfg.ServerPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	srv := grpc.NewServer(grpc.StatsHandler(otelgrpc.NewServerHandler()))

	collection.RegisterCollectionServiceServer(srv, server.NewCollectionServer(sqlstore.NewSqlStore(db), cfg.AuthSecret))

	log.Println("listening for connections...")

	_ = srv.Serve(listener)
}
