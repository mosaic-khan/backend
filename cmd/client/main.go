package main

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"main/pkg/UserAPIService"
)

func main() {
	conn, _ := grpc.Dial("82.115.13.61:9090", grpc.WithTransportCredentials(insecure.NewCredentials()))
	defer func(conn *grpc.ClientConn) {
		_ = conn.Close()
	}(conn)

	client := UserAPIService.NewUserAPIClient(conn)
	ctx := context.Background()

	loginResponse, err := client.Login(ctx, &UserAPIService.LoginRequest{
		UserNameOrEmail: "hjhj",
		Password:        "fhfhg",
	})
	if err != nil {
		return
	}
	println(loginResponse.JwtToken)
}
