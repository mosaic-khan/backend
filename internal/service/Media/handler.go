package Media

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"io"
	"main/internal/service/utils"
	"main/pkg/PostAPIService"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const UploadDir = "fileData"

func (s *Server) UploadProfilePicHandler(w http.ResponseWriter, r *http.Request) {
	profileID := r.Context().Value("profileID").(int64)

	if r.Method != "POST" {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse the multipart form, limiting file size to 10MB
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "The uploaded file is too big.", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("uploadFile")
	if err != nil {
		http.Error(w, "Could not get uploaded file.", http.StatusBadRequest)
		return
	}
	defer file.Close()

	if !strings.HasPrefix(header.Header.Get("Content-Type"), "image/") {
		http.Error(w, "The uploaded file must be an image.", http.StatusBadRequest)
		return
	}

	filename := utils.GenerateFileName()

	// Create and write the file
	dst, err := os.Create(fmt.Sprintf("%s/%s", UploadDir, filename))
	if err != nil {
		http.Error(w, "Could not create a file.", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, "Failed to save the uploaded file.", http.StatusInternalServerError)
		return
	}

	profilePicToken, err := utils.CreateProfilePicToken(strconv.Itoa(int(profileID)), filename, s.hmacSecret)
	if err != nil {
		http.Error(w, "error while creating Token", http.StatusInternalServerError)
	}

	_, _ = w.Write([]byte(profilePicToken))
	w.WriteHeader(http.StatusOK)
}

func (s *Server) UploadPostImagesHandler(w http.ResponseWriter, r *http.Request) {
	profileID := r.Context().Value("profileID").(int64)
	token := r.Context().Value("token").(string)

	if r.Method != "POST" {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse the multipart form, limiting file size to 10MB
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "The uploaded file is too big.", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("uploadFile")
	if err != nil {
		http.Error(w, "Could not get uploaded file.", http.StatusBadRequest)
		return
	}
	defer file.Close()

	if !strings.HasPrefix(header.Header.Get("Content-Type"), "image/") {
		http.Error(w, "The uploaded file must be an image.", http.StatusBadRequest)
		return
	}

	postIDStr := r.PostFormValue("postID")
	if postIDStr == "" {
		http.Error(w, "postID is required.", http.StatusBadRequest)
		return
	}
	postID, err := strconv.ParseInt(postIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid postID format.", http.StatusBadRequest)
		return
	}

	filename := utils.GenerateFileName()

	// Create and write the file
	dst, err := os.Create(fmt.Sprintf("%s/%s", UploadDir, filename))
	if err != nil {
		http.Error(w, "Could not create a file.", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, "Failed to save the uploaded file.", http.StatusInternalServerError)
		return
	}

	postImageToken, err := utils.CreatePostImageToken(strconv.Itoa(int(profileID)), strconv.Itoa(int(postID)), filename, s.hmacSecret)
	if err != nil {
		http.Error(w, "error while creating Token", http.StatusInternalServerError)
	}

	conn, _ := grpc.Dial("localhost:9190", grpc.WithTransportCredentials(insecure.NewCredentials()))
	defer func(conn *grpc.ClientConn) {
		_ = conn.Close()
	}(conn)

	client := PostAPIService.NewPostAPIClient(conn)

	ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("authorization", token))

	_, _ = client.AddImageForPost(ctx, &PostAPIService.AddImageForPostRequest{PostImageToken: postImageToken})

	_, _ = w.Write([]byte(postImageToken))
	w.WriteHeader(http.StatusOK)
}

func (s *Server) GetImageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Only GET method is allowed", http.StatusMethodNotAllowed)
		return
	}

	filePath := UploadDir + "/" + r.URL.Path[len("/KhanAPI.MediaAPI/images/"):]
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, "File not found.", http.StatusNotFound)
		return
	}

	http.ServeFile(w, r, filePath)
}
