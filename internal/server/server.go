package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/iamajraj/skema/internal/config"
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
			query := s.DB.Table(tableName)

			// 1. Filtering
			for _, field := range entity.Fields {
				val := r.URL.Query().Get(field.Name)
				if val != "" {
					if field.Type == "string" || field.Type == "text" {
						query = query.Where(fmt.Sprintf("%s LIKE ?", field.Name), "%"+val+"%")
					} else {
						query = query.Where(fmt.Sprintf("%s = ?", field.Name), val)
					}
				}
			}

			// 2. Sorting
			sort := r.URL.Query().Get("sort") // format: field:asc or field:desc
			if sort != "" {
				parts := strings.Split(sort, ":")
				if len(parts) == 2 {
					query = query.Order(fmt.Sprintf("%s %s", parts[0], parts[1]))
				} else {
					query = query.Order(parts[0])
				}
			} else {
				query = query.Order("created_at desc")
			}

			// 3. Pagination
			limitStr := r.URL.Query().Get("limit")
			offsetStr := r.URL.Query().Get("offset")

			limit := 100
			if limitStr != "" {
				fmt.Sscanf(limitStr, "%d", &limit)
			}
			offset := 0
			if offsetStr != "" {
				fmt.Sscanf(offsetStr, "%d", &offset)
			}

			results := []map[string]interface{}{}
			if err := query.Limit(limit).Offset(offset).Find(&results).Error; err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			s.expandData(entity, results, r.URL.Query().Get("expand"))

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

			if errMsg := s.validateData(entity, data); errMsg != "" {
				http.Error(w, errMsg, http.StatusBadRequest)
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

		// Get by ID
		r.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
			id := chi.URLParam(r, "id")
			result := make(map[string]interface{})
			dbRes := s.DB.Table(tableName).Where("id = ?", id).Scan(&result)
			if dbRes.Error != nil || dbRes.RowsAffected == 0 {
				http.Error(w, "Not Found", http.StatusNotFound)
				return
			}

			results := []map[string]interface{}{result}
			s.expandData(entity, results, r.URL.Query().Get("expand"))

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(results[0])
		})

		// Update by ID
		r.Put("/{id}", func(w http.ResponseWriter, r *http.Request) {
			id := chi.URLParam(r, "id")
			var data map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
				http.Error(w, "Invalid JSON", http.StatusBadRequest)
				return
			}

			if errMsg := s.validateData(entity, data); errMsg != "" {
				http.Error(w, errMsg, http.StatusBadRequest)
				return
			}

			data["updated_at"] = time.Now()

			if err := s.DB.Table(tableName).Where("id = ?", id).Updates(data).Error; err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			result := make(map[string]interface{})
			s.DB.Table(tableName).Where("id = ?", id).Scan(&result)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(result)
		})

		// Delete by ID
		r.Delete("/{id}", func(w http.ResponseWriter, r *http.Request) {
			id := chi.URLParam(r, "id")
			if err := s.DB.Table(tableName).Where("id = ?", id).Delete(nil).Error; err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		})
	})
}

func (s *Server) validateData(entity config.EntityConfig, data map[string]interface{}) string {
	for _, field := range entity.Fields {
		val, exists := data[field.Name]

		// Required check
		if field.Required && (!exists || val == nil || val == "") {
			return fmt.Sprintf("field '%s' is required", field.Name)
		}

		if exists && val != nil {
			// Min/Max for numbers
			if field.Type == "int" || field.Type == "float" {
				var num float64
				switch v := val.(type) {
				case int:
					num = float64(v)
				case float64:
					num = v
				case int64:
					num = float64(v)
				}

				if field.Min != nil && num < float64(*field.Min) {
					return fmt.Sprintf("field '%s' must be at least %d", field.Name, *field.Min)
				}
				if field.Max != nil && num > float64(*field.Max) {
					return fmt.Sprintf("field '%s' must be at most %d", field.Name, *field.Max)
				}
			}

			// Pattern check for strings
			if field.Pattern != "" {
				res, _ := regexp.MatchString(field.Pattern, fmt.Sprintf("%v", val))
				if !res {
					return fmt.Sprintf("field '%s' does not match pattern '%s'", field.Name, field.Pattern)
				}
			}

			// Format checks
			if field.Format == "email" {
				emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
				if !emailRegex.MatchString(fmt.Sprintf("%v", val)) {
					return fmt.Sprintf("field '%s' must be a valid email", field.Name)
				}
			}
		}
	}

	// Relationship Validation
	for _, rel := range entity.Relations {
		if rel.Type == "belongs_to" {
			val, exists := data[rel.Field]
			if exists && val != nil {
				targetTable := strings.ToLower(rel.Entity) + "s"
				var count int64
				s.DB.Table(targetTable).Where("id = ?", val).Count(&count)
				if count == 0 {
					return fmt.Sprintf("related %s with id %v does not exist", rel.Entity, val)
				}
			}
		}
	}

	return ""
}

func (s *Server) expandData(entity config.EntityConfig, results []map[string]interface{}, expandParam string) {
	if expandParam == "" {
		return
	}

	expands := strings.Split(expandParam, ",")
	for _, exp := range expands {
		var relation *config.RelationConfig
		for _, r := range entity.Relations {
			// Match singular or plural (e.g., expand=post or expand=posts)
			match := strings.EqualFold(r.Entity, exp) ||
				strings.EqualFold(r.Entity+"s", exp)
			if match {
				relation = &r
				break
			}
		}

		if relation != nil {
			targetTable := strings.ToLower(relation.Entity) + "s"
			for i := range results {
				if relation.Type == "belongs_to" {
					targetID := results[i][relation.Field]
					if targetID != nil {
						targetData := make(map[string]interface{})
						dbRes := s.DB.Table(targetTable).Where("id = ?", targetID).Scan(&targetData)
						if dbRes.Error == nil && dbRes.RowsAffected > 0 {
							results[i][strings.ToLower(relation.Entity)] = targetData
						}
					}
				} else if relation.Type == "has_many" {
					currentID := results[i]["id"]
					if currentID != nil {
						targetRecords := []map[string]interface{}{}
						if err := s.DB.Table(targetTable).Where(fmt.Sprintf("%s = ?", relation.Field), currentID).Find(&targetRecords).Error; err == nil {
							key := strings.ToLower(relation.Entity) + "s"
							results[i][key] = targetRecords
						}
					}
				}
			}
		}
	}
}

func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.Config.Server.Port)
	fmt.Printf("Server starting on %s\n", addr)
	return http.ListenAndServe(addr, s.Router)
}
