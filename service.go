package main

import (
	"context"
	"log"
	"strconv"

	"github.com/e-commerce-microservices/cart-service/pb"
	"github.com/e-commerce-microservices/cart-service/repository"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc/metadata"
)

type cartService struct {
	authClient pb.AuthServiceClient
	queries    *repository.Queries
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
	})
	if err != nil {
		log.Println("error when create cart: ", err)
		return nil, err
	}

	return &pb.CreateCartResponse{
		Message: "Thêm sản phẩm thành công vào giỏ hàng",
	}, nil
}

func (srv cartService) DeleteCart(ctx context.Context, req *pb.DeleteCartRequest) (*pb.DeleteCartResponse, error) {
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

	err = srv.queries.DeleteCart(ctx, repository.DeleteCartParams{
		ID:         req.GetCartId(),
		CustomerID: id,
	})
	if err != nil {
		log.Println("error when delete cart: ", err)
		return nil, err
	}

	return &pb.DeleteCartResponse{
		Message: "Xóa sản phẩm thành công khỏi giỏ hàng",
	}, nil
}

func (srv cartService) Ping(context.Context, *empty.Empty) (*pb.Pong, error) {
	return &pb.Pong{
		Message: "pong",
	}, nil
}
