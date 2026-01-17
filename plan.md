# JSON Translation Editor - Project Plan

## Project Overview
A web-based JSON editor specifically designed for managing translation files, with a focus on identifying and filling missing translations across language files.

## Tech Stack
- **Backend**: Go (Golang)
- **Frontend**: Templ + TemplUI components + HTMX + TailwindCSS
tempui guide in llms.txt file
- **Database**: TursoDB (libSQL/SQLite fork)
- **Migrations**: Goose
- **Auth**: API key-based or shareable links
- **HTTP framework**: "github.com/labstack/echo/v4"


## Core Features
1. Upload/manage two JSON files (base English + target language)
2. Compare files and identify missing keys
3. Form-based editing that prevents JSON syntax errors
4. Support for nested objects/structures
5. Toggle between "missing only" and "full view"
6. Generate shareable edit links with API keys

---

## Project Structure

```
translation-editor/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── handlers/
│   │   ├── home.go
│   │   ├── project.go
│   │   ├── editor.go
│   │   └── api.go
│   ├── models/
│   │   ├── project.go
│   │   ├── translation.go
│   │   └── apikey.go
│   ├── database/
│   │   ├── db.go
│   │   └── queries.go
│   ├── jsontools/
│   │   ├── parser.go
│   │   ├── differ.go
│   │   └── flattener.go
│   └── middleware/
│       └── auth.go
├── templates/
│   ├── layout.templ
│   ├── home.templ
│   ├── editor.templ
│   └── components/
│       ├── json_form.templ
│       └── nav.templ
├── migrations/
│   └── 001_initial_schema.sql
├── static/
│   ├── css/
│   └── js/
├── go.mod
└── go.sum
```

---

## Phase 1: Setup & Foundation

### 1.1 Initialize Project
```bash
go mod init translation-editor
go get github.com/a-h/templ
go get github.com/tursodatabase/libsql-client-go
go get github.com/pressly/goose/v3
go get github.com/labstack/echo/v4 
```

### 1.2 Database Schema (Goose Migration)
```sql
-- migrations/001_initial_schema.sql
-- +goose Up
CREATE TABLE projects (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE files (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    file_type TEXT NOT NULL, -- 'base' or 'target'
    language_code TEXT NOT NULL,
    content TEXT NOT NULL, -- JSON as text
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);

CREATE TABLE api_keys (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    key_hash TEXT NOT NULL UNIQUE,
    permissions TEXT DEFAULT 'read,write', -- JSON or comma-separated
    expires_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);

CREATE INDEX idx_api_keys_hash ON api_keys(key_hash);
CREATE INDEX idx_files_project ON files(project_id);

-- +goose Down
DROP TABLE api_keys;
DROP TABLE files;
DROP TABLE projects;
```

### 1.3 Environment Configuration
```env
DATABASE_URL=libsql://your-turso-db.turso.io
DATABASE_AUTH_TOKEN=your-token
PORT=8080
```

---

## Phase 2: Core JSON Processing Logic

### 2.1 JSON Tools Module

**Key Functions Needed:**

1. **Flattener** - Convert nested JSON to flat key-value pairs
```go
// Example: {"user": {"name": "John"}} → {"user.name": "John"}
func FlattenJSON(data map[string]interface{}, prefix string) map[string]string
```

2. **Differ** - Compare two flattened JSONs
```go
type Difference struct {
    MissingKeys []string
    ExtraKeys   []string
    DifferentValues map[string]struct{Base, Target string}
}

func CompareJSON(base, target map[string]string) Difference
```

3. **Reconstructor** - Rebuild nested JSON from flat structure
```go
func UnflattenJSON(flat map[string]string) map[string]interface{}
```

4. **Validator** - Ensure JSON structure is valid
```go
func ValidateTranslationFile(content []byte) error
```

### 2.2 Data Models

```go
type Project struct {
    ID        string
    Name      string
    CreatedAt time.Time
    UpdatedAt time.Time
}

type TranslationFile struct {
    ID           string
    ProjectID    string
    FileType     string // "base" or "target"
    LanguageCode string // "en", "es", "fr", etc.
    Content      string // JSON as string
    ParsedData   map[string]interface{} // In-memory only
}

type APIKey struct {
    ID          string
    ProjectID   string
    KeyHash     string
    Permissions []string
    ExpiresAt   *time.Time
}
```

---

## Phase 3: Backend Implementation

### 3.1 Handlers

**Home Handler** (`/`)
- Display recent projects
- Upload new project (2 files)
- Generate shareable link

**Project Handler** (`/project/:id`)
- Show project overview
- Display file info
- Generate/manage API keys

**Editor Handler** (`/project/:id/edit`)
- Main editing interface
- Toggle "missing only" vs "full view"
- Form-based JSON editing
- HTMX for dynamic updates

**API Handlers**
- `POST /api/project` - Create new project
- `POST /api/project/:id/files` - Upload files
- `GET /api/project/:id/diff` - Get missing keys
- `PATCH /api/project/:id/translations` - Update translation
- `GET /api/project/:id/export` - Export updated JSON

### 3.2 Auth Middleware

```go
func APIKeyAuth(next echo.HandlerFunc) echo.HandlerFunc {
    return func(c echo.Context) error {
        key := c.QueryParam("key")
        // or c.Request().Header.Get("X-API-Key")
        
        // Validate key against database
        // Check permissions
        // Set project context
        
        return next(c)
    }
}
```

---

## Phase 4: Frontend (Templ + HTMX)

### 4.1 Key Templates

**Layout Template**
- Base HTML structure
- TailwindCSS/TempUI styles
- HTMX script inclusion

**Editor Template**
- Nested form structure
- Each translation key as editable field
- Visual nesting indicators (indentation, grouping)
- Toggle switch for view mode

**Components**
- JSON field component (recursive for nested objects)
- Missing key badge
- Save button with HTMX
- Export button

### 4.2 HTMX Interactions

```html
<!-- Toggle view mode -->
<button hx-get="/project/123/edit?view=missing" 
        hx-target="#editor-content"
        hx-swap="innerHTML">
    Show Missing Only
</button>

<!-- Update translation -->
<input type="text" 
       name="user.profile.name"
       hx-post="/api/project/123/translations"
       hx-trigger="blur"
       hx-swap="none"
       value="">

<!-- Export -->
<button hx-get="/api/project/123/export?lang=es"
        hx-swap="none">
    Download Spanish JSON
</button>
```

### 4.3 Form Structure for Nested Objects

```html
<!-- Visual representation -->
<div class="nested-group">
    <h3>user</h3>
    <div class="nested-content ml-4">
        <h4>profile</h4>
        <div class="ml-4">
            <label>name</label>
            <input name="user.profile.name" value="">
            
            <label>email</label>
            <input name="user.profile.email" value="">
        </div>
    </div>
</div>
```

---

## Phase 5: Features Implementation

### 5.1 File Upload Flow
1. User uploads base (English) JSON
2. User uploads target language JSON
3. Backend validates both files
4. Store in database
5. Run comparison
6. Generate shareable link with API key

### 5.2 Missing Keys Detection
1. Flatten both JSON files
2. Compare keys
3. Identify missing in target
4. Display only missing keys in form
5. Allow toggle to full view

### 5.3 Live Editing
1. Render form with current values
2. On blur/change → HTMX PATCH request
3. Update database
4. Return success indicator
5. No page reload

### 5.4 Export
1. Get current state from DB
2. Merge updates into original structure
3. Unflatten to nested JSON
4. Return as downloadable file

---

## Phase 6: Nice-to-Have Features

1. **Progress Indicator** - Show % of translations completed
2. **Search/Filter** - Find specific keys
3. **Bulk Operations** - Mark multiple as complete
4. **History** - Track changes over time
5. **Collaboration** - Multiple editors with different API keys
6. **JSON Schema Support** - Validate against schema
7. **Pluralization Support** - Handle `{count}` variables
8. **Comments** - Add notes to translations
9. **Auto-translate** - Integration with translation APIs
10. **Git Integration** - Commit changes directly

---

## Implementation Timeline

**Week 1: Foundation**
- Project setup
- Database schema & migrations
- Basic models and database layer

**Week 2: Core Logic**
- JSON processing tools (flatten, diff, unflatten)
- File upload handling
- Basic CRUD operations

**Week 3: Frontend**
- Templ templates
- HTMX integration
- Editor UI with nested forms

**Week 4: Features & Polish**
- API key authentication
- Missing keys toggle
- Export functionality
- Testing & bug fixes

**Week 5: Enhancement**
- Progress indicators
- Better UX/UI
- Documentation
- Deployment

---

## Testing Strategy

1. **Unit Tests**
   - JSON flattening/unflattening
   - Diff algorithm
   - Key generation

2. **Integration Tests**
   - Database operations
   - File upload/parsing
   - API endpoints

3. **E2E Tests**
   - Complete workflow from upload to export
   - HTMX interactions

---

## Deployment Considerations

- **TursoDB Setup**: Create database, get connection URL
- **Server**: Deploy to Fly.io, Railway, or similar
- **Static Assets**: Serve via server or CDN
- **Environment Variables**: Secure storage of DB credentials
- **Monitoring**: Basic logging and error tracking

---

## Security Considerations

1. API key hashing (bcrypt/argon2)
2. Rate limiting on endpoints
3. File size limits for uploads
4. JSON validation to prevent injection
5. CORS configuration
6. XSS protection in Templ templates (automatic)

---

## Getting Started

```bash
# 1. Initialize project
mkdir translation-editor && cd translation-editor
go mod init translation-editor

# 2. Install dependencies
go get github.com/a-h/templ
go get github.com/tursodatabase/libsql-client-go/libsql
go get github.com/pressly/goose/v3
go get github.com/labstack/echo/v4

# 3. Setup TursoDB
turso db create translation-editor
turso db show translation-editor --url
turso db tokens create translation-editor

# 4. Run migrations
goose -dir migrations postgres $DATABASE_URL up

# 5. Generate Templ files
templ generate

# 6. Run server
go run cmd/server/main.go
```

This plan provides a complete roadmap. Would you like me to start implementing any specific part first?
