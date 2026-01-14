package docs

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/iamajraj/skema/internal/config"
)

func RegisterSwagger(r *chi.Mux, cfg *config.Config) {
	spec := generateOpenAPI(cfg)
	specJSON, _ := json.MarshalIndent(spec, "", "  ")

	r.Get("/openapi.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(specJSON)
	})

	r.Get("/docs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
  <head>
    <title>%s Docs</title>
    <meta charset="utf-8"/>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link href="https://fonts.googleapis.com/css?family=Montserrat:300,400,700|Roboto:300,400,700" rel="stylesheet">
    <style>body { margin: 0; padding: 0; }</style>
  </head>
  <body>
    <div id="redoc-container"></div>
    <script src="https://cdn.redoc.ly/redoc/latest/bundles/redoc.standalone.js"></script>
    <script>
      Redoc.init('/openapi.json', {
        scrollYOffset: 50
      }, document.getElementById('redoc-container'))
    </script>
  </body>
</html>`, cfg.Server.Name)
	})

	r.Get("/swagger", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>%s - Swagger UI</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css" >
    <style>
        html { box-sizing: border-box; overflow: -moz-scrollbars-vertical; overflow-y: scroll; }
        *, *:before, *:after { box-sizing: inherit; }
        body { margin:0; background: #fafafa; }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"> </script>
    <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-standalone-preset.js"> </script>
    <script>
    window.onload = function() {
      const ui = SwaggerUIBundle({
        url: "/openapi.json",
        dom_id: '#swagger-ui',
        deepLinking: true,
        presets: [
          SwaggerUIBundle.presets.apis,
          SwaggerUIStandalonePreset
        ],
        plugins: [
          SwaggerUIBundle.plugins.DownloadUrl
        ],
        layout: "StandaloneLayout"
      })
      window.ui = ui
    }
    </script>
</body>
</html>`, cfg.Server.Name)
	})
}

func generateOpenAPI(cfg *config.Config) map[string]interface{} {
	paths := make(map[string]interface{})
	components := make(map[string]interface{})
	schemas := make(map[string]interface{})
	var tags []map[string]interface{}

	for _, entity := range cfg.Entities {
		name := entity.Name
		lowerName := strings.ToLower(name)
		collectionPath := "/" + lowerName + "s"
		itemPath := collectionPath + "/{id}"

		tags = append(tags, map[string]interface{}{
			"name":        name,
			"description": "Operations for " + name,
		})

		// Schema definition
		schemaProperties := make(map[string]interface{})
		schemaProperties["id"] = map[string]interface{}{"type": "integer"}
		for _, field := range entity.Fields {
			prop := map[string]interface{}{"type": mapType(field.Type)}
			schemaProperties[field.Name] = prop
		}
		schemaProperties["created_at"] = map[string]interface{}{"type": "string", "format": "date-time"}
		schemaProperties["updated_at"] = map[string]interface{}{"type": "string", "format": "date-time"}

		schemas[name] = map[string]interface{}{
			"type":       "object",
			"properties": schemaProperties,
		}

		// Paths
		collectionParams := []interface{}{
			map[string]interface{}{"name": "limit", "in": "query", "schema": map[string]interface{}{"type": "integer", "default": 100}, "description": "Limit the number of records"},
			map[string]interface{}{"name": "offset", "in": "query", "schema": map[string]interface{}{"type": "integer", "default": 0}, "description": "Number of records to skip"},
			map[string]interface{}{"name": "sort", "in": "query", "schema": map[string]interface{}{"type": "string"}, "description": "Sort order (e.g., name:asc, created_at:desc)"},
		}

		// Relationships for expansion
		var expandable []string
		for _, rel := range entity.Relations {
			if rel.Type == "belongs_to" {
				expandable = append(expandable, strings.ToLower(rel.Entity))
			} else {
				expandable = append(expandable, strings.ToLower(rel.Entity)+"s")
			}
		}
		if len(expandable) > 0 {
			collectionParams = append(collectionParams, map[string]interface{}{
				"name":        "expand",
				"in":          "query",
				"schema":      map[string]interface{}{"type": "string"},
				"description": "Expand related data: " + strings.Join(expandable, ", "),
			})
		}

		// Dynamic filters
		for _, field := range entity.Fields {
			collectionParams = append(collectionParams, map[string]interface{}{
				"name":        field.Name,
				"in":          "query",
				"schema":      map[string]interface{}{"type": mapType(field.Type)},
				"description": "Filter by " + field.Name,
			})
		}

		paths[collectionPath] = map[string]interface{}{
			"get": map[string]interface{}{
				"tags":       []string{name},
				"summary":    "List all " + lowerName + "s",
				"parameters": collectionParams,
				"responses": map[string]interface{}{
					"200": map[string]interface{}{
						"description": "A list of " + lowerName + "s",
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"type": "object",
									"properties": map[string]interface{}{
										"success": map[string]interface{}{"type": "boolean"},
										"data": map[string]interface{}{
											"type": "array",
											"items": map[string]interface{}{
												"$ref": "#/components/schemas/" + name,
											},
										},
										"meta": map[string]interface{}{
											"type": "object",
											"properties": map[string]interface{}{
												"total":  map[string]interface{}{"type": "integer"},
												"limit":  map[string]interface{}{"type": "integer"},
												"offset": map[string]interface{}{"type": "integer"},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"post": map[string]interface{}{
				"tags":    []string{name},
				"summary": "Create a new " + lowerName,
				"requestBody": map[string]interface{}{
					"required": true,
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{"$ref": "#/components/schemas/" + name},
						},
					},
				},
				"responses": map[string]interface{}{
					"201": map[string]interface{}{
						"description": "Created",
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"type": "object",
									"properties": map[string]interface{}{
										"success": map[string]interface{}{"type": "boolean"},
										"data":    map[string]interface{}{"$ref": "#/components/schemas/" + name},
									},
								},
							},
						},
					},
				},
			},
		}

		paths[itemPath] = map[string]interface{}{
			"parameters": []interface{}{
				map[string]interface{}{
					"name":     "id",
					"in":       "path",
					"required": true,
					"schema":   map[string]interface{}{"type": "integer"},
				},
				map[string]interface{}{
					"name":        "expand",
					"in":          "query",
					"schema":      map[string]interface{}{"type": "string"},
					"description": "Expand related data: " + strings.Join(expandable, ", "),
				},
			},
			"get": map[string]interface{}{
				"tags":    []string{name},
				"summary": "Get " + lowerName + " by ID",
				"responses": map[string]interface{}{
					"200": map[string]interface{}{
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"type": "object",
									"properties": map[string]interface{}{
										"success": map[string]interface{}{"type": "boolean"},
										"data":    map[string]interface{}{"$ref": "#/components/schemas/" + name},
									},
								},
							},
						},
					},
				},
			},
			"put": map[string]interface{}{
				"tags":    []string{name},
				"summary": "Update " + lowerName + " by ID",
				"requestBody": map[string]interface{}{
					"required": true,
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{"$ref": "#/components/schemas/" + name},
						},
					},
				},
				"responses": map[string]interface{}{
					"200": map[string]interface{}{
						"description": "Updated",
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"type": "object",
									"properties": map[string]interface{}{
										"success": map[string]interface{}{"type": "boolean"},
										"data":    map[string]interface{}{"$ref": "#/components/schemas/" + name},
									},
								},
							},
						},
					},
				},
			},
			"delete": map[string]interface{}{
				"tags":    []string{name},
				"summary": "Delete " + lowerName + " by ID",
				"responses": map[string]interface{}{
					"204": map[string]interface{}{"description": "Deleted"},
				},
			},
		}
	}

	components["schemas"] = schemas

	return map[string]interface{}{
		"openapi": "3.0.0",
		"info": map[string]interface{}{
			"title":   cfg.Server.Name,
			"version": "1.0.0",
		},
		"tags":       tags,
		"paths":      paths,
		"components": components,
	}
}

func mapType(t string) string {
	switch t {
	case "string", "text":
		return "string"
	case "int":
		return "integer"
	case "bool":
		return "boolean"
	case "float":
		return "number"
	default:
		return "string"
	}
}
