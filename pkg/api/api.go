package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// R is a global router reference used for testing or external access.
var R *chi.Mux

// AuthMiddleware is a function type that wraps an http.HandlerFunc to add authentication.
// It is used to protect routes that require a valid JWT token.
type AuthMiddleware func(http.HandlerFunc) http.HandlerFunc

// Init initializes the API routes on the provided chi router.
// It registers public routes that do not require authentication:
//   - GET  /api/nextdate – calculate next date for repeating tasks
//   - POST /api/signin   – authenticate and receive a JWT token
//
// Protected routes are registered separately via RegisterProtectedRoutes.
func Init(r *chi.Mux) {
	// Public routes (no authentication required)
	r.Get("/api/nextdate", nextDayHandler)
	r.Post("/api/signin", SigninHandler)

	// Protected routes (require authentication)
	// We'll register these separately with auth middleware
	// The actual registration with auth happens in server.go
	R = r
}

// RegisterProtectedRoutes registers all routes that require authentication.
// If auth is not nil, each route is wrapped with the authentication middleware.
// If auth is nil (e.g., in development when no password is set), routes are registered directly.
// Protected routes include:
//   - POST   /api/task      – create a new task
//   - GET    /api/tasks     – list tasks (with optional search parameters)
//   - GET    /api/task      – retrieve a single task by ID
//   - PUT    /api/task      – update an existing task
//   - DELETE /api/task      – delete a task
//   - POST   /api/task/done – mark a repeating task as done and compute next date
func RegisterProtectedRoutes(r *chi.Mux, auth AuthMiddleware) {
	if auth != nil {
		r.Post("/api/task", auth(addTaskHandler))
		r.Get("/api/tasks", auth(tasksHandler))
		r.Get("/api/task", auth(taskHandler))
		r.Put("/api/task", auth(taskHandler))
		r.Delete("/api/task", auth(taskHandler))
		r.Post("/api/task/done", auth(taskDoneHandler))
	} else {
		// No auth, register directly
		r.Post("/api/task", addTaskHandler)
		r.Get("/api/tasks", tasksHandler)
		r.Get("/api/task", taskHandler)
		r.Put("/api/task", taskHandler)
		r.Delete("/api/task", taskHandler)
		r.Post("/api/task/done", taskDoneHandler)
	}
}
