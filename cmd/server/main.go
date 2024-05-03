package main

import (
	"log"
	"main/internal/service/Media"
	"main/internal/service/Post"
	"main/internal/service/User"
	"main/internal/service/utils"
	"main/pkg/PostAPIService"
	"main/pkg/UserAPIService"
	"net"
	"net/http"
	"sync"

	"google.golang.org/grpc"
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

	v.Done()
}

func StartPostAPIServer() {
	lis, err := net.Listen("tcp", ":9190")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
		return
	}

	postS := Post.NewServer()
	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(utils.MiddleWareAuth()))
	PostAPIService.RegisterPostAPIServer(grpcServer, postS)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}

	v.Done()
}

func StartMediaAPIServer() {
	lis, err := net.Listen("tcp", ":9290")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
		return
	}

	mediaS := Media.NewServer()

	Media.Init(mediaS)

	if err := http.Serve(lis, nil); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}

	v.Done()
}

var v sync.WaitGroup

func main() {
	println("hello world")

	v.Add(3)

	go StartUserAPIServer()
	go StartPostAPIServer()
	go StartMediaAPIServer()

	v.Wait()
}
