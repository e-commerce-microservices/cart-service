package main

import (
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/e-commerce-microservices/cart-service/pb"
	"github.com/e-commerce-microservices/cart-service/repository"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"

	// postgres driver
	_ "github.com/lib/pq"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}
}
func main() {
	// init user db connection
	pgDSN := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWD"), os.Getenv("DB_DBNAME"),
	)
	cartDB, err := sql.Open("postgres", pgDSN)
	if err != nil {
		log.Fatal(err)
	}
	defer cartDB.Close()
	if err := cartDB.Ping(); err != nil {
		log.Fatal("can't ping to user db", err)
	}
	grpcServer := grpc.NewServer()

	authConn, err := grpc.Dial("auth-service:8080", grpc.WithInsecure())
	if err != nil {
		log.Fatal("can't dial auth service", err)
	}
	authClient := pb.NewAuthServiceClient(authConn)
	// init queries
	queries := repository.New(cartDB)
	cartService := cartService{
		authClient: authClient,
		queries:    queries,
	}
	pb.RegisterCartServiceServer(grpcServer, cartService)

	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal("cannot create listener: ", err)
	}
	log.Printf("start gRPC server on %s", listener.Addr().String())
	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatal("cannot create grpc server: ", err)
	}

}
