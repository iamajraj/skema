# Skema

**Skema** is a powerful Go-powered tool that allows you to instantly spin up a fully functional, production-ready CRUD API server defined by a single YAML file. It handles database migrations, relationships, advanced querying, and serves a beautiful ReDoc documentation UI.

## Features

- **Instant CRUD**: Automatically generates `GET`, `POST`, `GET /id`, `PUT`, and `DELETE` endpoints.
- **Dynamic Database**: Automatically creates SQLite tables and handles Foreign Key constraints.
- **Smart Validation**: Enforce data integrity with `min`, `max`, `pattern` (regex), and `format` constraints.
- **Advanced Querying**: Built-in support for filtering, sorting (`?sort=age:desc`), and pagination (`?limit=10&offset=0`).
- **Intelligent Relationships**: Support for `belongs_to` and `has_many` with on-demand data expansion (`?expand=posts`).
- **Auto-Documentation**: Generates OpenAPI 3.0 specs and serves an interactive **ReDoc UI**.
- **Automated Timestamps**: Every record automatically tracks `created_at` and `updated_at`.

---

## Configuration Guide (`skema.yml`)

The entire server behavior is controlled by one file.

### 1. Server Settings

```yaml
server:
  port: 8080
  name: 'My Awesome API'
```

### 2. Entities & Fields

Define your database tables as entities.

#### Field Types:

- `string`: Standard text.
- `int`: Whole numbers.
- `bool`: True/False values.
- `text`: Long content/descriptions.
- `float`: Decimal numbers.

#### Field Constraints (Validators):

- `required: true`: Field must be present and non-empty.
- `unique: true`: Field value must be unique in the table.
- `min: <int>`: Minimum value for `int` or `float` fields.
- `max: <int>`: Maximum value for `int` or `float` fields.
- `pattern: "<regex>"`: Value must match the provided regular expression.
- `format: "email"`: Validates that the string is a properly formatted email.

### 3. Relationships

Skema handles linkages between your data.

- **`belongs_to`**: Adds a foreign key to the table and enables object expansion.
- **`has_many`**: Enables fetching a collection of related items.

```yaml
relations:
  - type: belongs_to
    entity: User
    field: user_id
```

---

## API Usage

### Getting Started

```bash
go run cmd/skema/main.go --config skema.yml
```

### Standard Response Format

Every response from Skema follows a standardized production-grade structure:

#### For Collections (GET /entities):

```json
{
  "success": true,
  "data": [...],
  "meta": {
    "total": 100,
    "limit": 10,
    "offset": 0
  }
}
```

#### For Single Resources (POST, GET /entities/:id, PUT):

```json
{
  "success": true,
  "data": { ... }
}
```

### Advanced Querying

- **Filtering**: `/users?name=Alice` (String fields use partial matching).
- **Sorting**: `/users?sort=age:desc` or `/users?sort=created_at:asc`.
- **Pagination**: `/users?limit=10&offset=20`.
- **Expansion**: Nested related data using `?expand`.
  - `GET /posts?expand=user` (Singular expansion for `belongs_to`).
  - `GET /users/1?expand=posts` (Plural expansion for `has_many`).

---

## Documentation

Once the server is running, visit:

- **Interactive UI**: `http://localhost:8080/docs`
- **Raw Specification**: `http://localhost:8080/openapi.json`

## Example `skema.yml`

```yaml
server:
  port: 8080
  name: 'Blog Engine'

entities:
  - name: User
    fields:
      - name: name
        type: string
        required: true
        pattern: '^[a-zA-Z ]+$'
      - name: email
        type: string
        unique: true
        format: email
      - name: age
        type: int
        min: 18
    relations:
      - type: has_many
        entity: Post
        field: user_id

  - name: Post
    fields:
      - name: title
        type: string
        required: true
      - name: user_id
        type: int
        required: true
    relations:
      - type: belongs_to
        entity: User
        field: user_id
```

---

## Testing

Skema includes a comprehensive test suite for configuration, database, and API logic.

To run all tests:

```bash
go test ./...
```

## Author

[Muhammad Raj](https://github.com/iamajraj)
