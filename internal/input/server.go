package input

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/dontpanicw/DelayedNotifier/internal/port"
)

//go:embed static
var staticFS embed.FS

type Server struct {
	uc  port.Usecases
	mux *http.ServeMux
}

func NewServer(uc port.Usecases) *Server {
	s := &Server{uc: uc, mux: http.NewServeMux()}

	s.mux.HandleFunc("POST /api/notifications", s.handleCreateNotification)
	s.mux.HandleFunc("GET /api/notifications", s.handleListNotifications)
	s.mux.HandleFunc("GET /api/notifications/{id}/status", func(w http.ResponseWriter, r *http.Request) {
		s.handleGetNotificationStatus(w, r, r.PathValue("id"))
	})
	s.mux.HandleFunc("DELETE /api/notifications/{id}", func(w http.ResponseWriter, r *http.Request) {
		s.handleDeleteNotification(w, r, r.PathValue("id"))
	})

	dist, _ := fs.Sub(staticFS, "static")
	s.mux.Handle("/", http.FileServer(http.FS(dist)))

	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}
