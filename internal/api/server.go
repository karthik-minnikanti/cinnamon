package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/karthik-minnikanti/cinnamon/internal/models"
	"github.com/karthik-minnikanti/cinnamon/internal/storage"
)

type Server struct {
	router    *mux.Router
	storage   storage.Storage
	staticDir string
}

func NewServer(storage storage.Storage, staticDir string) *Server {
	s := &Server{
		router:    mux.NewRouter(),
		storage:   storage,
		staticDir: staticDir,
	}
	s.setupRoutes()
	return s
}

func (s *Server) Router() http.Handler {
	return s.router
}

func (s *Server) setupRoutes() {
	// API routes
	s.router.HandleFunc("/api/connections", s.handlePostConnection).Methods("POST")
	s.router.HandleFunc("/api/connections", s.handleGetConnections).Methods("GET")
	s.router.HandleFunc("/api/connections/error/{error}", s.handleGetConnectionsByError).Methods("GET")
	s.router.HandleFunc("/api/connections/process/{pid}", s.handleGetConnectionsByProcess).Methods("GET")

	// Serve static files
	s.router.PathPrefix("/").Handler(http.FileServer(http.Dir(s.staticDir)))
}

func (s *Server) handlePostConnection(w http.ResponseWriter, r *http.Request) {
	var conn models.Connection
	if err := json.NewDecoder(r.Body).Decode(&conn); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := s.storage.StoreConnection(&conn); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleGetConnections(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	if limit <= 0 {
		limit = 100
	}

	connections, err := s.storage.GetConnections(limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(connections)
}

func (s *Server) handleGetConnectionsByError(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	errorType := models.ConnectionError(vars["error"])

	connections, err := s.storage.GetConnectionsByError(errorType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(connections)
}

func (s *Server) handleGetConnectionsByProcess(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pid, err := strconv.Atoi(vars["pid"])
	if err != nil {
		http.Error(w, "Invalid process ID", http.StatusBadRequest)
		return
	}

	connections, err := s.storage.GetConnectionsByProcess(pid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(connections)
}
