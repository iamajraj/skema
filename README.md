# Skema

Skema is a Go-powered tool that allows you to instantly spin up a fully functional CRUD API server just by defining your schema in a YAML file.

## Features

- **Instant CRUD**: Automatically generates GET, POST, PUT, and DELETE endpoints for your entities.
- **Dynamic Database**: Automatically creates and manages SQLite tables based on your schema.
- **Auto Documentation**: Generates OpenAPI 3.0 specs and serves them via an interactive ReDoc UI.
- **Automatic Timestamps**: Handles `created_at` and `updated_at` for every record.
- **Zero Configuration**: Just one YAML file is all you need.

## Quick Start

### 1. Define your schema (`skema.yml`)

```yaml
server:
  port: 8080
  name: 'My Awesome API'

entities:
  - name: User
    fields:
      - name: name
        type: string
        required: true
      - name: email
        type: string
        unique: true
      - name: age
        type: int

  - name: Post
    fields:
      - name: title
        type: string
        required: true
      - name: content
        type: text
      - name: published
        type: bool
```

### 2. Run the server

```bash
go run cmd/skema/main.go --config skema.yml
```

### 3. Explore APIs & Docs

- **API Base**: `http://localhost:8080`
- **Interactive Docs**: `http://localhost:8080/docs`
- **OpenAPI Spec**: `http://localhost:8080/openapi.json`

## Supported Types

- `string`
- `int`
- `bool`
- `text`
- `float`

## API Endpoints

For each entity (e.g., `User` -> `/users`):

- `GET /users` - List all records
- `POST /users` - Create a new record
- `GET /users/:id` - Get a specific record
- `PUT /users/:id` - Update a record
- `DELETE /users/:id` - Delete a record

## Author
[Muhammad Raj](https://github.com/iamajraj)