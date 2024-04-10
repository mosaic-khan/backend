package User

import (
	"context"
	"database/sql"
	"main/internal/storage/db"
	"main/pkg/UserAPIService"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Server) Login(ctx context.Context, in *UserAPIService.LoginRequest) (*UserAPIService.LoginResponse, error) {

	// verify user
	// try to get user by email
	userEmail, err1 := s.query.GetUserByEmail(ctx, in.UserNameOrEmail)
	if err1 != nil && err1 != sql.ErrNoRows {
		return nil, status.Errorf(codes.Internal, "Error retrieving user %s\n", in.UserNameOrEmail)
	}
	// try to get user by username
	userUsername, err2 := s.query.GetUserByUsername(ctx, in.UserNameOrEmail)
	if err2 != nil && err2 != sql.ErrNoRows {
		return nil, status.Errorf(codes.Internal, "Error retrieving user %s\n", in.UserNameOrEmail)
	}
	// user doesn't exist
	if err1 != nil && err2 != nil {
		return nil, status.Errorf(codes.NotFound, "No such user %s\n", in.UserNameOrEmail)
	}

	var user db.Account

	if err1 == nil {
		user = userEmail
	} else {
		user = userUsername
	}
	// if not verified raise error
	// compare password
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(in.Password))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return nil, status.Errorf(codes.InvalidArgument, "Incorrect password")
	} else if err != nil {
		return nil, status.Errorf(codes.Internal, "Error checking password")
	}

	// get userID

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &jwt.MapClaims{
		"ExpiresAt": jwt.NewNumericDate(time.Now().Add(time.Hour * 12)),
		"Issuer":    "Khan",
		"ID":        user.ID,
	})

	tokenString, err := token.SignedString(s.hmacSecret)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Error creating token")
	}

	return &UserAPIService.LoginResponse{
		User: &UserAPIService.User{
			FName:         user.FirstName.String,
			LName:         user.LastName.String,
			Username:      user.Username,
			Email:         user.Email,
			BirthDay:      user.BirthDay.Time.String(),
			Gender:        string(user.Gender.Gender),
			ProfilePicUrl: "",
		},
		JwtToken: tokenString,
	}, nil

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
