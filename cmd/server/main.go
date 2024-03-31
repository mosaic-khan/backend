package main

import (
	"google.golang.org/grpc"
	"log"
	utils "main/internal/service"
	"main/internal/service/User"
	"main/pkg/UserAPIService"
	"net"
)

func StartUserAPIServer() {
	lis, err := net.Listen("tcp", ":9090")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
		return
	}

	userS := User.NewServer()
	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(utils.MiddleWareAuth()))
	UserAPIService.RegisterUserAPIServer(grpcServer, userS)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

func StartPostAPIServer() {
	lis, err := net.Listen("tcp", ":9190")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
		return
	}

	//postS := Post.NewServer()
	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(utils.MiddleWareAuth()))
	//UserAPIService.RegisterUserAPIServer(grpcServer, postS)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

func main() {
	println("hello world")

	go StartUserAPIServer()
	go StartPostAPIServer()

}
