package utils

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/smtp"
	"os"
	"regexp"
	"strconv"
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

		profileIDStr, err := token.Claims.GetSubject()
		if err != nil {
			return nil, status.Error(codes.Internal, "error while extracting profileID")
		}

		profileID, err := strconv.ParseInt(profileIDStr, 10, 64)
		if err != nil {
			return nil, status.Error(codes.Internal, "error while converting userID")
		}

		// Create new context
		newCtx := context.WithValue(ctx, "ProfileID", profileID)

		// Call handler
		return handler(newCtx, req)
	}

}

func ValidateUsername(username string) bool {
	var usernameRegex = regexp.MustCompile(`^[a-zA-z0-9_-]{3,32}$`)
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
	//const charset = `ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789`
	const charset = `0123456789`
	// for front test
	const codeLen = 4
	b := make([]byte, codeLen)
	for i := 0; i < codeLen; i++ {
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

func SendSignUpEmail(email string, code string) {
	username := "mmdhossein.haghdadi@gmail.com"
	password := "xsmtpsib-eb6248a76b82480199faf72cd07e43092f9d8c6ed89357698b5ac6a362171213-sRFDAq53XJx0c9nM"

	from := "no-reply@khanmedia.ir"

	// Receiver email address.
	to := []string{
		email,
	}

	// smtp server configuration.
	smtpHost := "smtp-relay.brevo.com"
	smtpPort := "587"

	// Message.
	message, err := verificationEmail(code)
	if err != nil {
		fmt.Println(err)
		return
	}

	mimeHeaders := "MIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n"

	// Email subject.
	header := fmt.Sprintf("From: no-reply@khanmedia.ir\r\nSubject: Email Verification\r\nTo: %s\r\n", email)

	// Putting together the email message with headers and body content.
	emailMessage := []byte(header + mimeHeaders + "\r\n" + message)

	// Authentication.
	auth := smtp.PlainAuth("", username, password, smtpHost)

	// Sending email.
	err = smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, emailMessage)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func SendResetPassEmail(email string, token string) {
	username := "mmdhossein.haghdadi@gmail.com"
	password := "xsmtpsib-eb6248a76b82480199faf72cd07e43092f9d8c6ed89357698b5ac6a362171213-sRFDAq53XJx0c9nM"

	from := "no-reply@khanmedia.ir"

	// Receiver email address.
	to := []string{
		email,
	}

	// smtp server configuration.
	smtpHost := "smtp-relay.brevo.com"
	smtpPort := "587"

	// Message.
	message, err := forgetPassEmail(fmt.Sprintf("khanmedia.ir/forgetpass?token=%s", token))
	if err != nil {
		fmt.Println(err)
		return
	}

	mimeHeaders := "MIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n"

	// Email subject.
	header := fmt.Sprintf("From: no-reply@khanmedia.ir\r\nSubject: Reset Password\r\nTo: %s\r\n", email)

	// Putting together the email message with headers and body content.
	emailMessage := []byte(header + mimeHeaders + "\r\n" + message)

	// Authentication.
	auth := smtp.PlainAuth("", username, password, smtpHost)

	// Sending email.
	err = smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, emailMessage)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func MediaMiddleware(next http.Handler) http.Handler {
	hmacSecret := []byte(os.Getenv("SECRET_KEY"))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		bearerToken := ""
		if strings.HasPrefix(authHeader, "Bearer ") {
			bearerToken = strings.TrimPrefix(authHeader, "Bearer ")
		} else {
			log.Printf("No Bearer Token found")
			return
		}

		token, err := jwt.Parse(bearerToken, func(token *jwt.Token) (interface{}, error) {
			// Don't forget to validate the alg is what you expect:
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return hmacSecret, nil
		})
		if err != nil {
			return
		}

		profileIDStr, err := token.Claims.GetSubject()
		if err != nil {
			return
		}

		profileID, err := strconv.ParseInt(profileIDStr, 10, 64)
		if err != nil {
			return
		}

		ctx := context.WithValue(r.Context(), "profileID", profileID)
		r = r.WithContext(ctx)

		wrappedWriter := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(wrappedWriter, r.WithContext(ctx))
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func GenerateFileName() string {
	const charset = `ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789`
	// for front test
	const codeLen = 20
	b := make([]byte, codeLen)
	for i := 0; i < codeLen; i++ {
		b[i] = charset[rand.Int()%len(charset)]
	}
	return string(b)
}

func CreateProfilePicToken(profileID string, filename string, key []byte) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &jwt.RegisteredClaims{
		Issuer:    "KhanWeb",
		Subject:   profileID,
		Audience:  jwt.ClaimStrings{"Media ProfilePic", filename},
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 2)),
	})

	return token.SignedString(key)
}

func CreatePostImageToken(profileID string, filename string, key []byte) (string, error) {
	return "", nil
}
