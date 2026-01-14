package config

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Entities []EntityConfig `yaml:"entities"`
}

type ServerConfig struct {
	Port int    `yaml:"port"`
	Name string `yaml:"name"`
}

type EntityConfig struct {
	Name   string        `yaml:"name"`
	Fields []FieldConfig `yaml:"fields"`
}

type FieldConfig struct {
	Name     string `yaml:"name"`
	Type     string `yaml:"type"` // string, int, bool, text, float
	Required bool   `yaml:"required"`
	Unique   bool   `yaml:"unique"`
}
