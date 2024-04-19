package utils

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func MiddleWareAuth() func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (midResponse interface{}, midErr error) {
	hmacSecret := []byte(os.Getenv("SECRET_KEY"))

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

		userID, err := token.Claims.GetSubject()
		if err != nil {
			return nil, status.Error(codes.Internal, "error while extracting userID")
		}

		// Create new context
		newCtx := context.WithValue(ctx, "userID", userID)

		// Call handler
		return handler(newCtx, req)
	}

}

func ValidateUsername(username string) bool {
	var usernameRegex = regexp.MustCompile(`^[a-zA-z0-9_-]{8,32}$`)
	return usernameRegex.MatchString(username)
}

func ValidateEmail(mail string) bool {
	var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(mail)
}

func ValidateName(name string) bool {
	var nameRegex = regexp.MustCompile(`^[a-zA-Z ]{3,40}$`)
	return nameRegex.MatchString(name)
}

func ValidatePassword(password string) bool {
	var lowerChar = regexp.MustCompile(`[a-z]`)
	var upperChar = regexp.MustCompile(`[A-Z]`)
	var digit = regexp.MustCompile(`\d`)
	var specialChar = regexp.MustCompile(`[!@#$%^&*_]`)
	var length = regexp.MustCompile(`^.{8,72}$`)
	return lowerChar.MatchString(password) && upperChar.MatchString(password) && digit.MatchString(password) && specialChar.MatchString(password) && length.MatchString(password)
}

func GenerateVerificationCode() string {
	const charset = `ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789`
	b := make([]byte, 6)
	for i := 0; i < 6; i++ {
		b[i] = charset[rand.Int()%len(charset)]
	}
	return string(b)
}

func CreateLoginToken(userID string, duration time.Duration, key []byte) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
		Issuer:    "KhanWeb",
		Subject:   userID,
		Audience:  jwt.ClaimStrings{"Login"},
	})

	return token.SignedString(key)
}

func SendSignUpEmail(code string) {

}
