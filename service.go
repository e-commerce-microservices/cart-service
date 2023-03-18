package main

import (
	"context"
	"log"
	"strconv"

	"github.com/e-commerce-microservices/cart-service/pb"
	"github.com/e-commerce-microservices/cart-service/repository"
	"github.com/golang/protobuf/ptypes/empty"
	"go.opentelemetry.io/otel"
	"google.golang.org/grpc/metadata"
)

type cartService struct {
	authClient    pb.AuthServiceClient
	productClient pb.ProductServiceClient
	queries       *repository.Queries
	pb.UnimplementedCartServiceServer
}

var _empty = &empty.Empty{}

func (srv cartService) CreateCart(ctx context.Context, req *pb.CreateCartRequest) (*pb.CreateCartResponse, error) {
	// auth
	var err error
	md, _ := metadata.FromIncomingContext(ctx)
	ctx = metadata.NewOutgoingContext(ctx, md)

	// auth
	claims, err := srv.authClient.GetUserClaims(ctx, _empty)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	id, _ := strconv.ParseInt(claims.GetId(), 10, 64)
	err = srv.queries.CreateCart(ctx, repository.CreateCartParams{
		CustomerID: id,
		ProductID:  req.GetProductId(),
		Quantity:   req.GetQuantity(),
	})
	if err != nil {
		log.Println("error when create cart: ", err)
		return nil, err
	}

	return &pb.CreateCartResponse{
		Message: "Thêm sản phẩm thành công vào giỏ hàng",
	}, nil
}

var tracer = otel.Tracer("cart-service")

func (srv cartService) DeleteCart(ctx context.Context, req *pb.DeleteCartRequest) (*pb.DeleteCartResponse, error) {
	ctx, span := tracer.Start(ctx, "deleteCart")
	defer span.End()
	// auth
	var err error
	md, _ := metadata.FromIncomingContext(ctx)
	ctx = metadata.NewOutgoingContext(ctx, md)

	// auth
	claims, err := srv.authClient.GetUserClaims(ctx, _empty)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	id, _ := strconv.ParseInt(claims.GetId(), 10, 64)

	_, span = tracer.Start(ctx, "CartService.Database.Delete")
	err = srv.queries.DeleteCart(ctx, repository.DeleteCartParams{
		ID:         req.GetCartId(),
		CustomerID: id,
	})
	span.End()

	if err != nil {
		log.Println("error when delete cart: ", err)
		return nil, err
	}

	return &pb.DeleteCartResponse{
		Message: "Xóa sản phẩm thành công khỏi giỏ hàng",
	}, nil
}

func (srv cartService) GetCartByCustomer(ctx context.Context, req *pb.GetCartByCustomerRequest) (*pb.GetCartByCustomerResponse, error) {
	// auth
	var err error
	md, _ := metadata.FromIncomingContext(ctx)
	ctx = metadata.NewOutgoingContext(ctx, md)

	// auth
	claims, err := srv.authClient.GetUserClaims(ctx, _empty)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	id, _ := strconv.ParseInt(claims.GetId(), 10, 64)
	listCart, err := srv.queries.GetCart(ctx, id)
	if err != nil {
		return nil, err
	}

	result := make([]*pb.GetCartByCustomerResponse_Cart, 0, len(listCart))

	// get product by id

	for _, cart := range listCart {
		prod, err := srv.productClient.GetProduct(ctx, &pb.GetProductRequest{
			ProductId: cart.ProductID,
		})
		if err != nil {
			continue
		}
		result = append(result, &pb.GetCartByCustomerResponse_Cart{
			Id: cart.ID,
			Product: &pb.Product{
				SupplierId: prod.GetSupplierId(),
				Name:       prod.GetName(),
				Price:      prod.GetPrice(),
				Thumbnail:  prod.GetThumbnail(),
				ProductId:  prod.GetProductId(),
			},
			Quantity: cart.Quantity,
		})
	}

	return &pb.GetCartByCustomerResponse{
		ListCart: result,
	}, nil
}

func (srv cartService) Ping(context.Context, *empty.Empty) (*pb.Pong, error) {
	return &pb.Pong{
		Message: "pong",
	}, nil
}
