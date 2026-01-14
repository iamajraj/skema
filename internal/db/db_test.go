package db

import (
	"os"
	"testing"

	"github.com/iamajraj/skema/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestInitDB(t *testing.T) {
	dbPath := "test_init.db"
	defer os.Remove(dbPath)

	cfg := &config.Config{
		Entities: []config.EntityConfig{
			{
				Name: "Task",
				Fields: []config.FieldConfig{
					{Name: "title", Type: "string", Required: true},
					{Name: "done", Type: "bool"},
				},
			},
		},
	}

	database, err := InitDB(cfg, dbPath)
	assert.NoError(t, err)
	assert.NotNil(t, database)

	// Check if table exists
	assert.True(t, database.Migrator().HasTable("tasks"))

	// Check columns
	assert.True(t, database.Migrator().HasColumn("tasks", "id"))
	assert.True(t, database.Migrator().HasColumn("tasks", "title"))
	assert.True(t, database.Migrator().HasColumn("tasks", "done"))
	assert.True(t, database.Migrator().HasColumn("tasks", "created_at"))
}

func TestForeignKeys(t *testing.T) {
	dbPath := "test_fk.db"
	defer os.Remove(dbPath)

	cfg := &config.Config{
		Entities: []config.EntityConfig{
			{Name: "User", Fields: []config.FieldConfig{{Name: "name", Type: "string"}}},
			{
				Name:   "Profile",
				Fields: []config.FieldConfig{{Name: "bio", Type: "text"}, {Name: "user_id", Type: "int"}},
				Relations: []config.RelationConfig{
					{Type: "belongs_to", Entity: "User", Field: "user_id"},
				},
			},
		},
	}

	database, err := InitDB(cfg, dbPath)
	assert.NoError(t, err)
	assert.True(t, database.Migrator().HasTable("users"))
	assert.True(t, database.Migrator().HasTable("profiles"))
}
