package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/iamajraj/skema/internal/config"
	"github.com/iamajraj/skema/internal/db"
	"github.com/stretchr/testify/assert"
)

func TestServerCRUD(t *testing.T) {
	// Setup mock config
	minAge := 18
	cfg := &config.Config{
		Server: config.ServerConfig{Name: "Test API", Port: 8080},
		Entities: []config.EntityConfig{
			{
				Name: "User",
				Fields: []config.FieldConfig{
					{Name: "name", Type: "string", Required: true},
					{Name: "email", Type: "string", Unique: true, Format: "email"},
					{Name: "age", Type: "int", Min: &minAge},
				},
			},
		},
	}

	// Initialize DB
	os.Remove("test_skema.db")
	database, err := db.InitDB(cfg, "test_skema.db")
	assert.NoError(t, err)

	srv := NewServer(cfg, database)

	// Test Case 1: Create User (Success)
	userData := map[string]interface{}{
		"name":  "Test User",
		"email": "test@example.com",
		"age":   25,
	}
	body, _ := json.Marshal(userData)
	req := httptest.NewRequest("POST", "/users", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	srv.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var createResp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &createResp)
	assert.True(t, createResp["success"].(bool))

	// Test Case 2: Validation Fail (Age < 18)
	userData["age"] = 10
	body, _ = json.Marshal(userData)
	req = httptest.NewRequest("POST", "/users", bytes.NewBuffer(body))
	w = httptest.NewRecorder()
	srv.Router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "must be at least 18")

	// Test Case 3: List Users
	req = httptest.NewRequest("GET", "/users", nil)
	w = httptest.NewRecorder()
	srv.Router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var listResp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &listResp)
	assert.True(t, listResp["success"].(bool))
	assert.Len(t, listResp["data"].([]interface{}), 1)

	// Clean up
	os.Remove("test_skema.db")
}
