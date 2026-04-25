package server

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/EugeneDFox/go_final_project/pkg/api"
	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
)

// webDir is the directory containing static web assets.
const webDir = "web"

// defaultPort is the default port if TODO_PORT is not set.
const defaultPort = "7540"

// Server represents an HTTP server that serves static files.
type Server struct {
	logger *log.Logger
	server *http.Server
}

// getAddr returns the address string (host:port) for the server.
// It reads the TODO_PORT environment variable; if not set or invalid, uses defaultPort.
func getAddr() string {
	port := os.Getenv("TODO_PORT")
	if port == "" {
		port = defaultPort
	}
	// Ensure port is numeric (basic validation)
	if _, err := strconv.Atoi(port); err != nil {
		port = defaultPort
	}
	return ":" + port
}

// NewServer creates a new Server instance with the given logger.
// The server will listen on the port defined by TODO_PORT environment variable
// (default :7540) and serve files from the "web" directory.
func NewServer(logger *log.Logger) *Server {
	r := chi.NewRouter()

	// Initialize API routes (public routes)
	api.Init(r)

	// Register protected routes with authentication middleware
	api.RegisterProtectedRoutes(r, Auth)

	// Static file handler (catch-all, should be last)
	r.Handle("/*", http.FileServer(http.Dir(webDir)))

	addr := getAddr()
	server := &http.Server{
		Addr:         addr,
		Handler:      r,
		ErrorLog:     logger,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	return &Server{
		logger: logger,
		server: server,
	}
}

// Start starts the HTTP server and blocks until the server stops.
func (s *Server) Start() error {
	s.logger.Printf("Server starting on %s", s.server.Addr)
	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the server with the given context.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

// CheckHashPassword computes the SHA‑256 hash of a document (string) and compares it
// with the provided hash (hex‑encoded). Returns true if they match.
func CheckHashPassword(document string, hash string) bool {
	documentBytes := []byte(document)
	result := sha256.Sum256(documentBytes)
	h := hex.EncodeToString(result[:])
	return h == hash
}

// Auth is a middleware that validates JWT tokens for protected routes.
// If TODO_PASSWORD environment variable is set, the middleware expects a "token" cookie
// containing a valid JWT signed with that password. If no password is set, authentication
// is bypassed (useful for development).
// The middleware returns HTTP 401 Unauthorized if the token is missing, invalid, or expired.
func Auth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pass := os.Getenv("TODO_PASSWORD")
		if len(pass) > 0 {
			cookie, err := r.Cookie("token")
			if err != nil {
				http.Error(w, "Authentication required", http.StatusUnauthorized)
				return
			}
			jwtToken, err := jwt.Parse(cookie.Value, func(t *jwt.Token) (interface{}, error) {
				return []byte(pass), nil
			})
			if err != nil {
				http.Error(w, "Authentication required", http.StatusUnauthorized)
				return
			}
			if !jwtToken.Valid {
				http.Error(w, "Authentication required", http.StatusUnauthorized)
				return
			}
		}
		next(w, r)
	})
}
