package db

import (
	"fmt"
	"strings"

	"github.com/iamajraj/skema/internal/config"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func InitDB(cfg *config.Config, path string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	for _, entity := range cfg.Entities {
		err := createTable(db, entity)
		if err != nil {
			return nil, err
		}
	}

	return db, nil
}

func createTable(db *gorm.DB, entity config.EntityConfig) error {
	tableName := strings.ToLower(entity.Name) + "s"

	// Check if table exists
	if db.Migrator().HasTable(tableName) {
		return nil
	}

	var columns []string
	columns = append(columns, "id INTEGER PRIMARY KEY AUTOINCREMENT")

	for _, field := range entity.Fields {
		sqlType := ""
		switch field.Type {
		case "string":
			sqlType = "TEXT"
		case "int":
			sqlType = "INTEGER"
		case "bool":
			sqlType = "BOOLEAN"
		case "text":
			sqlType = "TEXT"
		case "float":
			sqlType = "REAL"
		default:
			sqlType = "TEXT"
		}

		colDef := fmt.Sprintf("%s %s", field.Name, sqlType)
		if field.Required {
			colDef += " NOT NULL"
		}
		if field.Unique {
			colDef += " UNIQUE"
		}
		columns = append(columns, colDef)
	}

	columns = append(columns, "created_at DATETIME", "updated_at DATETIME")

	// Add Foreign Keys
	for _, rel := range entity.Relations {
		if rel.Type == "belongs_to" {
			targetTable := strings.ToLower(rel.Entity) + "s"
			columns = append(columns, fmt.Sprintf("FOREIGN KEY (%s) REFERENCES %s(id)", rel.Field, targetTable))
		}
	}

	query := fmt.Sprintf("CREATE TABLE %s (%s)", tableName, strings.Join(columns, ", "))
	return db.Exec(query).Error
}
