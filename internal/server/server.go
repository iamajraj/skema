package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/iamajraj/skema/internal/config"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"gorm.io/gorm"
)

type Server struct {
	Config *config.Config
	DB     *gorm.DB
	Router *chi.Mux
}

func NewServer(cfg *config.Config, db *gorm.DB) *Server {
	s := &Server{
		Config: cfg,
		DB:     db,
		Router: chi.NewRouter(),
	}

	s.setupMiddleware()
	s.setupRoutes()

	return s
}

func (s *Server) setupMiddleware() {
	s.Router.Use(middleware.Logger)
	s.Router.Use(middleware.Recoverer)
}

func (s *Server) setupRoutes() {
	s.Router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": fmt.Sprintf("Welcome to %s", s.Config.Server.Name)})
	})

	for _, entity := range s.Config.Entities {
		s.setupEntityRoutes(entity)
	}
}

func (s *Server) setupEntityRoutes(entity config.EntityConfig) {
	path := "/" + strings.ToLower(entity.Name) + "s"
	tableName := strings.ToLower(entity.Name) + "s"

	s.Router.Route(path, func(r chi.Router) {
		// List
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			results := []map[string]interface{}{}
			if err := s.DB.Table(tableName).Find(&results).Error; err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(results)
		})

		// Create
		r.Post("/", func(w http.ResponseWriter, r *http.Request) {
			var data map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
				http.Error(w, "Invalid JSON", http.StatusBadRequest)
				return
			}

			now := time.Now()
			data["created_at"] = now
			data["updated_at"] = now

			if err := s.DB.Table(tableName).Create(&data).Error; err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(data)
		})

		// Get, Update, Delete
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", func(w http.ResponseWriter, r *http.Request) {
				id := chi.URLParam(r, "id")
				var result map[string]interface{}
				if err := s.DB.Table(tableName).First(&result, id).Error; err != nil {
					http.Error(w, "Not Found", http.StatusNotFound)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(result)
			})

			r.Put("/", func(w http.ResponseWriter, r *http.Request) {
				id := chi.URLParam(r, "id")
				var data map[string]interface{}
				if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
					http.Error(w, "Invalid JSON", http.StatusBadRequest)
					return
				}

				data["updated_at"] = time.Now()

				if err := s.DB.Table(tableName).Where("id = ?", id).Updates(data).Error; err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				var result map[string]interface{}
				s.DB.Table(tableName).First(&result, id)
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(result)
			})

			r.Delete("/", func(w http.ResponseWriter, r *http.Request) {
				id := chi.URLParam(r, "id")
				if err := s.DB.Table(tableName).Delete(nil, id).Error; err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				w.WriteHeader(http.StatusNoContent)
			})
		})
	})
}

func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.Config.Server.Port)
	fmt.Printf("Server starting on %s\n", addr)
	return http.ListenAndServe(addr, s.Router)
}
