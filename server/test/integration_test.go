package test

import (
	"fmt"
	"log"
	"strings"
	"testing"

	"testbert/protobuf/collection"
	"testbert/server/config"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestIntegration(t *testing.T) {
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Fatal(err)
	}
	cfg := config.NewConfig()

	db := setupDB(cfg)

	// Comment to test an inmemory server
	//
	conn, err := grpc.NewClient(fmt.Sprintf("localhost:%s", cfg.ServerPort), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		_ = conn.Close()
	}()
	tc := NewClient(collection.NewCollectionServiceClient(conn), cfg.AuthSecret)

	// Uncomment to run against inmemory server (for local dev)
	//
	// tc, closer := newServer(db)
	// defer closer()

	defer func() {
		_, _ = db.Exec("DELETE FROM collections;")
	}()

	t.Run("Collection Suite", func(t *testing.T) {
		t.Run("Create Collection", func(t *testing.T) {
			testCreateCollection(t, tc)
		})
		t.Run("Get Collection", func(t *testing.T) {
			testGetCollection(t, tc)
		})
		t.Run("Update Collection", func(t *testing.T) {
			testUpdateCollection(t, tc)
		})
		t.Run("DeleteCollection", func(t *testing.T) {
			testDeleteCollection(t, tc)
		})
	})
	t.Run("ShareToken Suite", func(t *testing.T) {
		t.Run("Create ShareToken", func(t *testing.T) {
			testCreateShareToken(t, tc)
		})
		t.Run("Get Collection From Token", func(t *testing.T) {
			testGetSharedCollection(t, tc)
		})
		t.Run("Rate Limiting Share Token", func(t *testing.T) {
			testRateLimitedShareToken(t, tc)
		})
		t.Run("Delete ShareToken", func(t *testing.T) {
			testDeleteSharedToken(t, tc)
		})
	})
}

func setupDB(cfg *config.Configuration) *sqlx.DB {
	dbConn := fmt.Sprintf("host=%v port=%v user=%v password=%v sslmode=disable",
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBUser,
		cfg.DBPassword,
	)

	db, err := sqlx.Open("postgres", dbConn)
	if err != nil {
		log.Fatalf("unable to connect to to database: %v", err)
	}

	_, err = db.Exec("CREATE DATABASE testbert_tests;")
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		log.Fatalf("error creating test database: %v", err)
	}

	_ = db.Close()

	dbConn = fmt.Sprintf("host=%v port=%v dbname=%v user=%v password=%v sslmode=disable",
		cfg.DBHost,
		cfg.DBPort,
		"testbert_tests",
		cfg.DBUser,
		cfg.DBPassword,
	)

	db, err = sqlx.Open("postgres", dbConn)
	if err != nil {
		log.Fatalf("unable to connect to to database: %v", err)
	}

	_ = goose.SetDialect("postgres")
	_ = goose.Up(db.DB, "../migrations")

	return db
}
