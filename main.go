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
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"google.golang.org/grpc"

	// postgres driver
	_ "github.com/lib/pq"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}
	// tp, tpErr := jaegerTraceProvider()
	// if tpErr != nil {
	// 	log.Fatal(tpErr)
	// }
	// otel.SetTracerProvider(tp)
	// otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
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
	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()))

	authConn, err := grpc.Dial("auth-service:8080", grpc.WithInsecure())
	if err != nil {
		log.Fatal("can't dial auth service", err)
	}
	authClient := pb.NewAuthServiceClient(authConn)

	productConn, err := grpc.Dial("product-service:8080", grpc.WithInsecure())
	if err != nil {
		log.Fatal("can't dial auth service", err)
	}
	productClient := pb.NewProductServiceClient(productConn)

	// init queries
	queries := repository.New(cartDB)
	cartService := cartService{
		authClient:    authClient,
		productClient: productClient,
		queries:       queries,
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
func jaegerTraceProvider() (*sdktrace.TracerProvider, error) {

	// exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint("http://jaeger-all-in-one:14268/api/traces")))
	// exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint("http://jaeger-all-in-one:14268/api/traces")))
	exp, err := jaeger.New(jaeger.WithAgentEndpoint(jaeger.WithAgentHost("10.3.68.12")))

	if err != nil {
		log.Println("err: ", err)
		return nil, err
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("cart service"),
			attribute.String("environment", "development"),
		)),
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(1.0)),
	)

	return tp, nil
}
