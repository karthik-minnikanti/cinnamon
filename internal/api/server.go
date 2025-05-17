package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/karthik-minnikanti/cinnamon/internal/models"
	"github.com/karthik-minnikanti/cinnamon/internal/storage"
)

type Server struct {
	router  *mux.Router
	storage storage.Storage
}

func NewServer(storage storage.Storage) *Server {
	s := &Server{
		router:  mux.NewRouter(),
		storage: storage,
	}
	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	s.router.HandleFunc("/api/connections", s.handleConnections).Methods("GET", "POST")
	s.router.HandleFunc("/api/connections/stats", s.handleStats).Methods("GET")
	s.router.HandleFunc("/api/connections/{id}", s.handleConnectionDetails).Methods("GET")
	s.router.HandleFunc("/api/services", s.handleServices).Methods("GET")
	s.router.HandleFunc("/api/errors", s.handleErrors).Methods("GET")
	s.router.HandleFunc("/api/environments", s.handleEnvironments).Methods("GET")

	// Serve static files
	s.router.PathPrefix("/").Handler(http.FileServer(http.Dir("static")))
}

func (s *Server) handleConnections(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		s.getConnections(w, r)
	case "POST":
		s.createConnection(w, r)
	}
}

func (s *Server) getConnections(w http.ResponseWriter, r *http.Request) {
	// Get filter parameters
	service := r.URL.Query().Get("service")
	errorType := r.URL.Query().Get("error")
	environment := r.URL.Query().Get("environment")
	search := r.URL.Query().Get("search")

	// Get connections with filters
	connections, err := s.storage.GetConnections(service, errorType, environment, search)
	if err != nil {
		log.Printf("Error getting connections: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Create response object
	response := struct {
		Connections []*models.Connection `json:"connections"`
	}{
		Connections: connections,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding connections: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (s *Server) createConnection(w http.ResponseWriter, r *http.Request) {
	var conn models.Connection
	if err := json.NewDecoder(r.Body).Decode(&conn); err != nil {
		log.Printf("Error decoding connection: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Set timestamp if not provided
	if conn.Timestamp.IsZero() {
		conn.Timestamp = time.Now()
	}

	if err := s.storage.StoreConnection(&conn); err != nil {
		log.Printf("Error storing connection: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	// Get time range
	startTime := time.Now().Add(-24 * time.Hour) // Default to last 24 hours
	endTime := time.Now()

	// Get statistics
	stats, err := s.storage.GetStats(startTime, endTime)
	if err != nil {
		log.Printf("Error getting stats: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(stats); err != nil {
		log.Printf("Error encoding stats: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (s *Server) handleConnectionDetails(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	conn, err := s.storage.GetConnectionByID(id)
	if err != nil {
		log.Printf("Error getting connection details: %v", err)
		http.Error(w, "Connection not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(conn); err != nil {
		log.Printf("Error encoding connection details: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (s *Server) handleServices(w http.ResponseWriter, r *http.Request) {
	services, err := s.storage.GetServices()
	if err != nil {
		log.Printf("Error getting services: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(services); err != nil {
		log.Printf("Error encoding services: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (s *Server) handleErrors(w http.ResponseWriter, r *http.Request) {
	errors, err := s.storage.GetErrors()
	if err != nil {
		log.Printf("Error getting errors: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(errors); err != nil {
		log.Printf("Error encoding errors: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (s *Server) handleEnvironments(w http.ResponseWriter, r *http.Request) {
	environments, err := s.storage.GetEnvironments()
	if err != nil {
		log.Printf("Error getting environments: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(environments); err != nil {
		log.Printf("Error encoding environments: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (s *Server) Start(addr string) error {
	return http.ListenAndServe(addr, s.router)
}
