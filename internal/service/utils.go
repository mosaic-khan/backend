package utils

import (
	"context"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"os"
	"strings"
)

func MiddleWareAuth() func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (midResponse interface{}, midErr error) {
	hmacSecret := []byte(os.Getenv("hmacSecret"))

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (midResponse interface{}, midErr error) {

		allowedMethods := []string{
			"/KhanAPI.UserAPI/Login",
			"/KhanAPI.UserAPI/ForgetPassword",
			"/KhanAPI.UserAPI/NewPasswordWithToken",
			"/KhanAPI.UserAPI/SignUp",
			"/KhanAPI.UserAPI/CodeVerification",
		}

		for _, method := range allowedMethods {
			if info.FullMethod == method {
				return handler(ctx, req)
			}
		}

		// Get metadata
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Internal, "failed to extract metadata")
		}

		auth := md.Get("Authorization")
		if len(auth) == 0 {
			return nil, status.Error(codes.Unauthenticated, "missing token")
		}

		tokenStr := strings.Split(auth[0], " ")[1]

		// Validate token
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			// Don't forget to validate the alg is what you expect:
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return hmacSecret, nil
		})

		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}

		userID, err := token.Claims.GetIssuer()
		if err != nil {
			return nil, status.Error(codes.Internal, "error while extracting userID")
		}

		// Create new context
		newCtx := context.WithValue(ctx, "userID", userID)

		// Call handler
		return handler(newCtx, req)
	}

}
