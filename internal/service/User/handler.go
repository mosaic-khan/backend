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

func (s *Server) RefreshToken(ctx context.Context, _ *emptypb.Empty) (*UserAPIService.RefreshTokenResponse, error) {
	profileID := ctx.Value("ProfileID").(int64)

	tokenString, err := utils.CreateLoginToken(strconv.FormatInt(profileID, 10), time.Minute*5, s.hmacSecret)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error while creating Token")
	}

	return &UserAPIService.RefreshTokenResponse{Token: tokenString}, nil
}

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

	profileID, err := s.query.GetProfileID(ctx, user.ID)

	// generate Token
	tokenString, err := utils.CreateLoginToken(strconv.FormatInt(profileID, 10), time.Hour*12, s.hmacSecret)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Error creating token")
	}

	return &UserAPIService.LoginResponse{
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

	utils.SendResetPassEmail(in.UserNameOrEmail, tokenStr)

	return &emptypb.Empty{}, nil
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

	return &emptypb.Empty{}, nil
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

	go utils.SendSignUpEmail(in.Email, verificationCode)

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

	profileID, err := TXQuery.CreateProfile(ctx, userID)
	if err != nil {
		_ = TX.Rollback()
		return nil, err
	}

	loginToken, err := utils.CreateLoginToken(strconv.Itoa(int(profileID)), time.Hour*12, s.hmacSecret)
	if err != nil {
		_ = TX.Rollback()
		return nil, err
	}

	return &UserAPIService.CodeVerificationResponse{JwtToken: loginToken}, nil
}

func (s *Server) PersonalInfoCompletion(ctx context.Context, in *UserAPIService.PersonalInfoCompletionRequest) (*emptypb.Empty, error) {
	profileID := ctx.Value("ProfileID").(int64)

	t, err := time.Parse("2006-01-02", in.GetBirthDay())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error parsing birthday")
	}
	birthDay := sql.NullTime{
		Time:  t,
		Valid: true,
	}

	err = s.query.UpdateProfileInfo(ctx, db.UpdateProfileInfoParams{
		FirstName: in.FName,
		LastName:  in.LName,
		Gender:    db.Gender(in.Gender),
		BirthDay:  birthDay,
		Bio:       in.Bio,
		CityID:    sql.NullInt16{},
		ID:        profileID,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error while updating profile info")
	}

	return nil, nil
}

func (s *Server) EditProfileInfo(ctx context.Context, in *UserAPIService.EditProfileInfoRequest) (*emptypb.Empty, error) {
	profileID := ctx.Value("ProfileID").(int64)

	userInfo, err := s.query.GetProfileInfo(ctx, profileID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not retrieve user")
	}

	fname := ""
	lname := ""
	gender := db.Gender("")
	birthDay := sql.NullTime{}
	city := sql.NullInt16{}
	bio := ""

	// update first name if one is provided else keep the current first name
	if in.FName != nil {
		fname = in.GetFName()
	} else {
		fname = userInfo.FirstName
	}
	// update last name if one is provided else keep the current last name
	if in.LName != nil {
		lname = in.GetLName()
	} else {
		lname = userInfo.LastName
	}
	// update gender if one is provided else keep the current gender
	if in.Gender != nil {
		gender = db.Gender(*in.Gender)
	} else {
		gender = userInfo.Gender
	}
	// update birthday if one is provided else keep the current birthday
	if in.BirthDay != nil {
		t, err := time.Parse("2006-01-02", in.GetBirthDay())
		if err != nil {
			return nil, status.Errorf(codes.Internal, "error parsing birthday")
		}
		birthDay = sql.NullTime{
			Time:  t,
			Valid: true,
		}
	} else {
		birthDay = userInfo.BirthDay
	}

	if in.Bio != nil {
		bio = *in.Bio
	} else {
		bio = userInfo.Bio
	}

	if in.CityID != nil {
		city = sql.NullInt16{
			Int16: int16(*in.CityID),
			Valid: true,
		}
	} else {
		city = userInfo.CityID
	}

	err = s.query.UpdateProfileInfo(ctx, db.UpdateProfileInfoParams{
		FirstName: fname,
		LastName:  lname,
		Gender:    gender,
		BirthDay:  birthDay,
		Bio:       bio,
		CityID:    city,
		ID:        profileID,
	})

	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not update user profile")
	}

	return &emptypb.Empty{}, nil

}

func (s *Server) GetUserInfo(ctx context.Context, _ *emptypb.Empty) (*UserAPIService.GetUserInfoResponse, error) {
	profileID := ctx.Value("ProfileID").(int64)

	userInfo, err := s.query.GetProfileInfo(ctx, profileID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not retrieve user")
	}

	return &UserAPIService.GetUserInfoResponse{User: &UserAPIService.User{
		FName:         userInfo.FirstName,
		LName:         userInfo.LastName,
		Username:      userInfo.Username,
		Email:         userInfo.Email,
		BirthDay:      userInfo.BirthDay.Time.String(),
		Gender:        string(userInfo.Gender),
		ProfilePicUrl: userInfo.ProfilePicAddress,
		City:          userInfo.CityName.String,
		Bio:           userInfo.Bio,
	}}, nil
}

func (s *Server) ChangeUsername(context.Context, *UserAPIService.ChangeUsernameRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ChangeUsername not implemented")
}

func (s *Server) ChangeEmail(context.Context, *UserAPIService.ChangeEmailRequest) (*UserAPIService.ChangeEmailResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ChangeEmail not implemented")
}

func (s *Server) ConfirmChangeEmail(context.Context, *UserAPIService.ConfirmChangeEmailRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ConfirmChangeEmail not implemented")
}

func (s *Server) ChangePassword(ctx context.Context, in *UserAPIService.ChangePasswordRequest) (*emptypb.Empty, error) {
	profileID := ctx.Value("ProfileID").(int64)

	userID, err := s.query.GetUserIDbyProfileID(ctx, profileID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error while fetching userID")
	}

	user, err := s.query.GetUserByID(ctx, userID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "error fetching user by user ID")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(in.OldPassword))
	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return nil, status.Errorf(codes.InvalidArgument, "Incorrect password")
	} else if err != nil {
		return nil, status.Errorf(codes.Internal, "Error checking password")
	}

	bcryptPass, err := bcrypt.GenerateFromPassword([]byte(in.NewPassword), 10)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error while converting to bcrypt")
	}

	err = s.query.ResetPassword(ctx, db.ResetPasswordParams{
		ID:       userID,
		Password: string(bcryptPass),
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error while changing password")
	}

	return nil, nil
}

func (s *Server) ChangeProfilePic(context.Context, *UserAPIService.ChangeProfilePicRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ChangeProfilePic not implemented")
}

func (s *Server) GetProfile(ctx context.Context, in *UserAPIService.GetProfileRequests) (*UserAPIService.GetProfileResponse, error) {
	profileID := ctx.Value("ProfileID").(int64)

	var profile *UserAPIService.Profile
	var gender string

	if in.Username == nil {
		profileDB, err := s.query.GetProfileByProfileID(ctx, profileID)
		if err != nil {
			return nil, err
		}

		profile = &UserAPIService.Profile{
			Id:            profileDB.ID,
			Name:          profileDB.Name.(string),
			Username:      profileDB.Username,
			Pronouns:      "",
			Bio:           profileDB.Bio,
			City:          profileDB.CityName.String,
			ProfilePicUrl: profileDB.ProfilePicAddress,
		}

		gender = string(profileDB.Gender)

	} else {
		profileDB, err := s.query.GetProfileByUsername(ctx, in.GetUsername())
		if err != nil {
			return nil, err
		}

		profile = &UserAPIService.Profile{
			Id:            profileDB.ID,
			Name:          profileDB.Name.(string),
			Username:      profileDB.Username,
			Pronouns:      "",
			Bio:           profileDB.Bio,
			City:          profileDB.CityName.String,
			ProfilePicUrl: profileDB.ProfilePicAddress,
		}

		gender = string(profileDB.Gender)
	}

	if gender == "male" {
		profile.Pronouns = "He/Him"
	} else if gender == "female" {
		profile.Pronouns = "She/Her"
	}

	return &UserAPIService.GetProfileResponse{Profile: profile}, nil
}

func (s *Server) GetCities(ctx context.Context, in *UserAPIService.GetCitiesRequest) (*UserAPIService.GetCitiesResponse, error) {

	citiesDB, err := s.query.GetCities(ctx, sql.NullString{String: in.CityPattern, Valid: true})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error while retriving cities")
	}

	cities := make([]*UserAPIService.City, len(citiesDB))
	for i, city := range citiesDB {
		cities[i] = &UserAPIService.City{
			Id:   int32(city.ID),
			Name: city.Name,
		}
	}

	return &UserAPIService.GetCitiesResponse{Cities: cities}, nil
}

func (s *Server) DeleteAccount(ctx context.Context, in *emptypb.Empty) (*emptypb.Empty, error) {

	profileId := ctx.Value("ProfileID").(int64)

	accountID, err := s.query.GetProfileUserID(ctx, profileId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not get profile's account id")
	}

	err = s.query.DeleteUser(ctx, accountID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not delete account")
	}

	return &emptypb.Empty{}, nil
}
