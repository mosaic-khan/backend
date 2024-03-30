package User

import (
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"main/pkg/UserAPIService"
)

func (s *Server) Login(ctx context.Context, in *UserAPIService.LoginRequest) (*UserAPIService.LoginResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Login not implemented")
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
