package User

import (
	"context"
	"database/sql"
	"errors"
	"main/internal/service/utils"
	"main/internal/storage/db"
	"main/pkg/UserAPIService"
	"math/rand"
	"strconv"
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
	if err1 != nil && !errors.Is(err1, sql.ErrNoRows) {
		return nil, status.Errorf(codes.Internal, "Error retrieving user %s\n", in.UserNameOrEmail)
	}
	// try to get user by username
	userUsername, err2 := s.query.GetUserByUsername(ctx, in.UserNameOrEmail)
	if err2 != nil && !errors.Is(err2, sql.ErrNoRows) {
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
	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return nil, status.Errorf(codes.InvalidArgument, "Incorrect password")
	} else if err != nil {
		return nil, status.Errorf(codes.Internal, "Error checking password")
	}

	// generate Token
	tokenString, err := utils.CreateLoginToken(strconv.FormatInt(user.ID, 10), time.Hour*12, s.hmacSecret)
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

	user, err := s.query.GetUserByEmail(ctx, in.UserNameOrEmail)
	if err != nil {
		return nil, nil
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "KhanWeb",
		Subject:   strconv.FormatInt(user.ID, 10),
		Audience:  jwt.ClaimStrings{"ForgetPass"},
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Minute)),
		ID:        strconv.Itoa(rand.Int()),
	})
	tokenStr, err := token.SignedString(s.hmacSecret)
	if err != nil {
		return nil, nil
	}

	utils.SendResetPassEmail(tokenStr)

	return nil, nil
}

func (s *Server) NewPasswordWithToken(ctx context.Context, in *UserAPIService.NewPasswordWithTokenRequest) (*emptypb.Empty, error) {
	token, err := jwt.Parse(in.ResetPasswordToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, status.Errorf(codes.Unauthenticated, "unexpected signing method: %v", token.Header["alg"])
		}
		return s.hmacSecret, nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, status.Errorf(codes.Unauthenticated, "invalid token")
	}

	userIDStr, err := token.Claims.GetSubject()
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		return nil, err
	}

	if !utils.ValidatePassword(in.Password) {
		return nil, status.Errorf(codes.InvalidArgument, "invalid password")
	}

	bcryptPass, _ := bcrypt.GenerateFromPassword([]byte(in.Password), 10)

	err = s.query.ResetPassword(ctx, db.ResetPasswordParams{
		ID:       int64(userID),
		Password: string(bcryptPass),
	})
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (s *Server) SignUp(ctx context.Context, in *UserAPIService.SignUpRequest) (*UserAPIService.SignUpResponse, error) {
	if !utils.ValidateEmail(in.Email) {
		return nil, status.Errorf(codes.InvalidArgument, "invalid email")
	}
	if !utils.ValidateUsername(in.Username) {
		return nil, status.Errorf(codes.InvalidArgument, "invalid username")
	}
	if !utils.ValidatePassword(in.Password) {
		return nil, status.Errorf(codes.InvalidArgument, "invalid password")
	}

	EmailCnt, err := s.query.ExistsUserEmail(ctx, in.Email)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	if EmailCnt != 0 {
		return nil, status.Errorf(codes.AlreadyExists, "email already exist")
	}

	UserNameCnt, err := s.query.ExistsUserUsername(ctx, in.Username)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	if UserNameCnt != 0 {
		return nil, status.Errorf(codes.AlreadyExists, "username already exist")
	}

	signUpExpTime := time.Now().Add(5 * time.Minute)
	verificationCode := utils.GenerateVerificationCode()
	bcryptPass, err := bcrypt.GenerateFromPassword([]byte(in.Password), 10)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error hashing password")
	}

	signupID, err := s.query.InsertSignup(ctx, db.InsertSignupParams{
		Email:            in.Email,
		Username:         in.Username,
		Password:         string(bcryptPass),
		VerificationCode: verificationCode,
		Expire:           signUpExpTime,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	go utils.SendSignUpEmail(verificationCode)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(signUpExpTime),
		Issuer:    "KhanWeb",
		Subject:   strconv.Itoa(int(signupID)),
		Audience:  jwt.ClaimStrings{"SignUp"},
	})

	tokenString, err := token.SignedString(s.hmacSecret)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Error creating token")
	}

	return &UserAPIService.SignUpResponse{Token: tokenString}, nil
}

func (s *Server) CodeVerification(ctx context.Context, in *UserAPIService.CodeVerificationRequest) (*UserAPIService.CodeVerificationResponse, error) {
	token, err := jwt.Parse(in.SignUpToken, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, status.Errorf(codes.Unauthenticated, "unexpected signing method: %v", token.Header["alg"])
		}
		return s.hmacSecret, nil
	})
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, err.Error())
	}

	aud, _ := token.Claims.GetAudience()
	if len(aud) != 1 || aud[0] != "SignUp" {
		return nil, status.Errorf(codes.InvalidArgument, "invalid token")
	}

	signUpIDStr, _ := token.Claims.GetSubject()
	signUpID, err := strconv.Atoi(signUpIDStr)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	signUpRow, err := s.query.GetSignUpData(ctx, int32(signUpID))
	if err != nil {
		return nil, err
	}

	if signUpRow.VerificationCode != in.Code {
		return nil, status.Errorf(codes.InvalidArgument, "Wrong Code")
	}

	err = s.query.DeleteSignup(ctx, signUpRow.ID)
	if err != nil {
		return nil, err
	}

	TX, err := s.conn.Begin()
	defer func(TX *sql.Tx) {
		_ = TX.Commit()
	}(TX)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	TXQuery := s.query.WithTx(TX)

	usernameCnt, err := TXQuery.ExistsUserUsername(ctx, signUpRow.Username)
	if err != nil || usernameCnt != 0 {
		_ = TX.Rollback()
		return nil, err
	}

	emailCnt, err := TXQuery.ExistsUserEmail(ctx, signUpRow.Email)
	if err != nil || emailCnt != 0 {
		_ = TX.Rollback()
		return nil, err
	}

	userID, err := TXQuery.InsertUser(ctx, db.InsertUserParams{
		Email:    signUpRow.Email,
		Username: signUpRow.Username,
		Password: signUpRow.Password,
	})
	if err != nil {
		_ = TX.Rollback()
		return nil, err
	}

	loginToken, err := utils.CreateLoginToken(strconv.Itoa(int(userID)), time.Hour*12, s.hmacSecret)
	if err != nil {
		_ = TX.Rollback()
		return nil, err
	}

	return &UserAPIService.CodeVerificationResponse{JwtToken: loginToken}, nil
}

func (s *Server) PersonalInfoCompletion(ctx context.Context, in *UserAPIService.PersonalInfoCompletionRequest) (*UserAPIService.PersonalInfoCompletionRequest, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PersonalInfoCompletion not implemented")
}

func (s *Server) EditProfileInfo(ctx context.Context, in *UserAPIService.EditProfileInfoRequest) (*UserAPIService.EditProfileInfoResponse, error) {

	userIDStr := ctx.Value("userID").(string)
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	user, err := s.query.GetUserByID(ctx, userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not retrieve user")
	}

	fname := sql.NullString{}
	lname := sql.NullString{}
	gender := db.NullGender{}
	birthDay := sql.NullTime{}

	// update first name if one is provided else keep the current first name
	if in.FName != nil {
		fname = sql.NullString{
			String: *in.FName,
			Valid:  true,
		}
	} else {
		fname = user.FirstName
	}
	// update last name if one is provided else keep the current last name
	if in.LName != nil {
		lname = sql.NullString{
			String: *in.LName,
			Valid:  true,
		}
	} else {
		lname = user.LastName
	}
	// update gender if one is provided else keep the current gender
	if in.Gender != nil {
		gender = db.NullGender{
			Gender: db.Gender(*in.Gender),
			Valid:  true,
		}
	} else {
		gender = user.Gender
	}
	// update birthday if one is provided else keep the current birthday
	if in.BirthDay != nil {
		t, err := time.Parse("2006-01-02", *in.BirthDay)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "error parsing birthday")
		}
		birthDay = sql.NullTime{
			Time:  t,
			Valid: true,
		}
	} else {
		birthDay = user.BirthDay
	}

	err = s.query.UpdateUserInfo(ctx, db.UpdateUserInfoParams{
		FirstName: fname,
		LastName:  lname,
		Gender:    gender,
		BirthDay:  birthDay,
		ID:        userID,
	})

	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not update user profile")
	}

	return &UserAPIService.EditProfileInfoResponse{Ok: true}, nil

}
