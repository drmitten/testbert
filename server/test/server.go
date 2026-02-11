package test

import (
	"context"
	"log"
	"net"

	"testbert/protobuf/collection"
	"testbert/server/datastore/sqlstore"
	"testbert/server/server"

	"github.com/jmoiron/sqlx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

func newServer(db *sqlx.DB) (*TestClient, func()) {
	lis := bufconn.Listen(101024 * 1024)

	grpcSrv := grpc.NewServer()

	collection.RegisterCollectionServiceServer(grpcSrv, server.NewCollectionServer(sqlstore.NewSqlStore(db), "testkey"))
	go func() {
		if err := grpcSrv.Serve(lis); err != nil {
			log.Fatal(err)
		}
	}()

	conn, err := grpc.NewClient("passthrough://bufnet", grpc.WithContextDialer(func(_ context.Context, _ string) (net.Conn, error) {
		return lis.Dial()
	}), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}

	closer := func() {
		err := lis.Close()
		if err != nil {
			log.Fatal(err)
		}
		grpcSrv.Stop()
	}

	client := NewClient(collection.NewCollectionServiceClient(conn), "testkey")

	return client, closer
}
