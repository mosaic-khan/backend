package User

import (
	"context"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"main/pkg/UserAPIService"
	"strings"
	"time"
)

func (s *Server) Login(ctx context.Context, in *UserAPIService.LoginRequest) (*UserAPIService.LoginResponse, error) {

	// verify user
	// if not verified raise error
	// get userID

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 1000)),
		ID:        "givenID",
	})

	tokenString, err := token.SignedString(s.hmacSecret)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error while creating token")
	}

	return &UserAPIService.LoginResponse{
		User:     nil,
		JwtToken: tokenString,
	}, nil

}

func (s *Server) AuthMiddleware(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (midResponse interface{}, midErr error) {

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
	println(md)
	if !ok {
		return nil, status.Error(codes.Internal, "failed to extract metadata")
	}

	// Extract token
	auth := md.Get("Authorization")
	println(auth)
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
		return s.hmacSecret, nil
	})

	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}

	userID, err := token.Claims.GetIssuer()
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, "error while extracting userID")
	}

	// Create new context
	newCtx := context.WithValue(ctx, "userID", userID)

	// Call handler
	return handler(newCtx, req)
}

func (s *Server) ForgetPassword(ctx context.Context, in *UserAPIService.ForgetPasswordRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ForgetPassword not implemented")
}

func (s *Server) NewPasswordWithToken(ctx context.Context, in *UserAPIService.NewPasswordWithTokenRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method NewPasswordWithToken not implemented")
}

func (s *Server) SignUp(ctx context.Context, in *UserAPIService.SignUpRequest) (*UserAPIService.SignUpResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SignUp not implemented")
}

func (s *Server) CodeVerification(ctx context.Context, in *UserAPIService.CodeVerificationRequest) (*UserAPIService.CodeVerificationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CodeVerification not implemented")
}

func (s *Server) PersonalInfoCompletion(ctx context.Context, in *UserAPIService.PersonalInfoCompletionRequest) (*UserAPIService.PersonalInfoCompletionRequest, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PersonalInfoCompletion not implemented")
}
