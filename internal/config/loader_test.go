package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	yamlContent := `
server:
  port: 9090
  name: "Test API"
entities:
  - name: User
    fields:
      - name: name
        type: string
        required: true
`
	tmpfile, err := os.CreateTemp("", "skema_test_*.yml")
	assert.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write([]byte(yamlContent))
	assert.NoError(t, err)
	tmpfile.Close()

	cfg, err := LoadConfig(tmpfile.Name())
	assert.NoError(t, err)
	assert.Equal(t, 9090, cfg.Server.Port)
	assert.Equal(t, "Test API", cfg.Server.Name)
	assert.Len(t, cfg.Entities, 1)
	assert.Equal(t, "User", cfg.Entities[0].Name)
	assert.Equal(t, "name", cfg.Entities[0].Fields[0].Name)
	assert.True(t, cfg.Entities[0].Fields[0].Required)
}

func TestDefaultConfig(t *testing.T) {
	yamlContent := `
entities: []
`
	tmpfile, err := os.CreateTemp("", "skema_test_default_*.yml")
	assert.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write([]byte(yamlContent))
	assert.NoError(t, err)
	tmpfile.Close()

	cfg, err := LoadConfig(tmpfile.Name())
	assert.NoError(t, err)
	assert.Equal(t, 8080, cfg.Server.Port)
	assert.Equal(t, "Skema API", cfg.Server.Name)
}
