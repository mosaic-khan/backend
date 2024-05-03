package Media

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

const UploadDir = "fileData"

func (s *Server) uploadHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		// Parse the multipart form
		if err := r.ParseMultipartForm(10 << 20); err != nil { // limit your file size to 10MB
			http.Error(w, "The uploaded file is too big.", http.StatusBadRequest)
			return
		}

		file, header, err := r.FormFile("uploadFile")
		if err != nil {
			http.Error(w, "Could not get uploaded file.", http.StatusBadRequest)
			return
		}
		defer file.Close()

		// Create a new file in the uploads directory
		dst, err := os.Create(fmt.Sprintf("%s/%s", UploadDir, header.Filename))
		if err != nil {
			http.Error(w, "Could not create a file.", http.StatusInternalServerError)
			return
		}
		defer dst.Close()

		// Copy the uploaded file to the filesystem
		if _, err := io.Copy(dst, file); err != nil {
			http.Error(w, "Failed to save the uploaded file.", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "File uploaded successfully: %s", header.Filename)
	default:
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
	}
}

// downloadHandler serves the images from the uploads directory
func (s *Server) downloadHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		filePath := UploadDir + r.URL.Path[len("/images/"):]
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			http.Error(w, "File not found.", http.StatusNotFound)
			return
		}

		http.ServeFile(w, r, filePath)
	default:
		http.Error(w, "Only GET method is allowed", http.StatusMethodNotAllowed)
	}
}
