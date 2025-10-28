package main

import (
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

var (
	cosEndpoint  = os.Getenv("IBM_COS_ENDPOINT")
	cosAccessKey = os.Getenv("IBM_COS_ACCESS_KEY")
	cosSecretKey = os.Getenv("IBM_COS_SECRET_KEY")
	cosBucket    = os.Getenv("IBM_COS_BUCKET")
	cosUseSSL    = func() bool {
		v := os.Getenv("IBM_COS_USE_SSL")
		if v == "" {
			return true
		}
		b, _ := strconv.ParseBool(v)
		return b
	}()

	cloudantURL      = os.Getenv("CLOUDANT_URL")
	cloudantDB       = os.Getenv("CLOUDANT_DB")
	cloudantUser     = os.Getenv("CLOUDANT_USERNAME")
	cloudantPassword = os.Getenv("CLOUDANT_PASSWORD")
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/upload-file", uploadFileHandler)
	mux.HandleFunc("/save-json", saveJSONHandler)

	// wrap mux with simple logging middleware to watch request status
	handler := loggingMiddleware(mux)

	addr := ":8080"
	if a := os.Getenv("PORT"); a != "" {
		addr = ":" + a
	}

	log.Printf("starting server on %s", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

// loggingResponseWriter wraps http.ResponseWriter to capture status code
type loggingResponseWriter struct {
	http.ResponseWriter
	status int
}

func (l *loggingResponseWriter) WriteHeader(code int) {
	l.status = code
	l.ResponseWriter.WriteHeader(code)
}

// loggingMiddleware logs incoming requests and response status codes
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lrw := &loggingResponseWriter{ResponseWriter: w, status: 200}
		log.Printf("%s %s started", r.Method, r.URL.Path)
		next.ServeHTTP(lrw, r)
		log.Printf("%s %s completed with %d", r.Method, r.URL.Path, lrw.status)
	})
}

// uploadFileHandler accepts multipart/form-data with field "file" and optional "objectName"
func uploadFileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()

	if cosEndpoint == "" || cosAccessKey == "" || cosSecretKey == "" || cosBucket == "" {
		http.Error(w, "COS configuration not set", http.StatusInternalServerError)
		return
	}

	if err := r.ParseMultipartForm(32 << 20); err != nil { // 32MB
		http.Error(w, "parse multipart: "+err.Error(), http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "read file: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	objectName := r.FormValue("objectName")
	if objectName == "" {
		objectName = filepath.Base(header.Filename)
	}

	// Read file into memory to determine size (for simplicity). For large files, use temp file streaming.
	data, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "read file data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := UploadToIBMCOS(ctx, cosEndpoint, cosAccessKey, cosSecretKey, cosBucket, objectName, data, cosUseSSL); err != nil {
		http.Error(w, "upload failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, `{"object":"%s"}`+"\n", objectName)
}

// saveJSONHandler accepts application/json and forwards it to Cloudant as a new document
func saveJSONHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if cloudantURL == "" || cloudantDB == "" {
		http.Error(w, "Cloudant configuration not set", http.StatusInternalServerError)
		return
	}

	ctx := r.Context()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "read body: "+err.Error(), http.StatusBadRequest)
		return
	}

	respBody, status, err := SaveJSONToCloudant(ctx, cloudantURL, cloudantDB, cloudantUser, cloudantPassword, body)
	if err != nil {
		http.Error(w, "cloudant save failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(status)
	w.Write(respBody)
}

// Helper to extract file from multipart without using the header too much
func readFormFile(file multipart.File) ([]byte, error) {
	return io.ReadAll(file)
}
