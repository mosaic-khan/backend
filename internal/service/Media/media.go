package Media

import (
	"main/internal/service/utils"
	"net/http"
	"os"
)

type Server struct {
	hmacSecret []byte
}

func NewServer() *Server {
	return &Server{
		hmacSecret: []byte(os.Getenv("SECRET_KEY")),
	}
}

func Init(s *Server) {
	http.Handle("/KhanAPI.MediaAPI/upload-post-image", utils.MediaMiddleware(http.HandlerFunc(s.UploadPostImagesHandler)))
	http.Handle("/KhanAPI.MediaAPI/upload-profile-image", utils.MediaMiddleware(http.HandlerFunc(s.UploadProfilePicHandler)))
	http.Handle("/KhanAPI.MediaAPI/images/", http.HandlerFunc(s.GetImageHandler))
}
