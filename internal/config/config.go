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
	Name      string           `yaml:"name"`
	Fields    []FieldConfig    `yaml:"fields"`
	Relations []RelationConfig `yaml:"relations"`
}

type RelationConfig struct {
	Type   string `yaml:"type"` // belongs_to, has_many
	Entity string `yaml:"entity"`
	Field  string `yaml:"field"` // e.g., user_id
}

type FieldConfig struct {
	Name     string `yaml:"name"`
	Type     string `yaml:"type"` // string, int, bool, text, float
	Required bool   `yaml:"required"`
	Unique   bool   `yaml:"unique"`
	Min      *int   `yaml:"min,omitempty"`
	Max      *int   `yaml:"max,omitempty"`
	Pattern  string `yaml:"pattern,omitempty"`
	Format   string `yaml:"format,omitempty"` // email, uuid, etc.
}
