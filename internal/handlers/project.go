package handlers

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"log/slog"

	"github.com/labstack/echo/v4"

	"templui/internal/database"
	"templui/internal/jsontools"
	"templui/internal/models"
	"templui/internal/session"
)

type ProjectHandler struct {
	db *database.DB
}

func NewProjectHandler(db *database.DB) *ProjectHandler {
	return &ProjectHandler{db: db}
}

// CreateProjectRequest represents the request body for creating a project
type CreateProjectRequest struct {
	Name           string `json:"name"`
	BaseFile       string `json:"base_file"`       // JSON content as string
	TargetFile     string `json:"target_file"`     // JSON content as string
	BaseLanguage   string `json:"base_language"`   // e.g., "en"
	TargetLanguage string `json:"target_language"` // e.g., "es"
	IsLocked       bool   `json:"is_locked"`       // Whether to lock project with secret key
}

// CreateProject handles POST /api/project
func (h *ProjectHandler) CreateProject(c echo.Context) error {
	var req CreateProjectRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	// Validate base file
	if err := jsontools.ValidateTranslationFile([]byte(req.BaseFile)); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("Invalid base file: %v", err)})
	}

	// If TargetFile not provided, generate an "empty" version with same keys/shape.
	if req.TargetFile == "" {
		var base any
		if err := json.Unmarshal([]byte(req.BaseFile), &base); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": fmt.Sprintf("Invalid base JSON: %v", err),
			})
		}

		empty := makeTranslationSkeleton(base)

		b, err := json.MarshalIndent(empty, "", "  ")
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": fmt.Sprintf("Failed to build target file: %v", err),
			})
		}
		req.TargetFile = string(b)
	}
	// Validate target file
	if err := jsontools.ValidateTranslationFile([]byte(req.TargetFile)); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("Invalid target file: %v", err)})
	}

	// Generate project ID
	projectID := generateID()

	name := "translation project"
	if req.Name != "" {
		name = req.Name
	}

	// Create project
	sessionToken := session.GetSessionToken(c)
	project := &models.Project{
		ID:           projectID,
		Name:         name,
		SessionToken: sessionToken,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Generate secret key if project is locked
	var secretKey string
	if req.IsLocked {
		var keyHash string
		secretKey, keyHash = generateAPIKey()
		project.IsLocked = true
		project.SecretKeyHash = keyHash
		project.SecretKey = secretKey
	}

	if err := h.db.CreateProject(project); err != nil {
		slog.Error("Failed to create project", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create project"})
	}

	// Create base file
	baseFile := &models.TranslationFile{
		ID:           generateID(),
		ProjectID:    projectID,
		FileType:     "base",
		LanguageCode: req.BaseLanguage,
		Content:      req.BaseFile,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if err := h.db.CreateFile(baseFile); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create base file"})
	}

	// Create target file
	targetFile := &models.TranslationFile{
		ID:           generateID(),
		ProjectID:    projectID,
		FileType:     "target",
		LanguageCode: req.TargetLanguage,
		Content:      req.TargetFile,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if err := h.db.CreateFile(targetFile); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create target file"})
	}

	// Generate API key
	apiKey, keyHash := generateAPIKey()
	apiKeyRecord := &models.APIKey{
		ID:          generateID(),
		ProjectID:   projectID,
		KeyHash:     keyHash,
		Permissions: []string{"read", "write"},
		CreatedAt:   time.Now(),
	}
	if err := h.db.CreateAPIKey(apiKeyRecord); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create API key"})
	}

	response := map[string]interface{}{
		"project_id": projectID,
		"api_key":    apiKey,
		"url":        fmt.Sprintf("/project/%s/edit?key=%s", projectID, apiKey),
	}

	// Include secret key in response if project is locked
	if req.IsLocked {
		response["secret_key"] = secretKey
	}

	return c.JSON(http.StatusCreated, response)
}

// GetDiff handles GET /api/project/:id/diff
func (h *ProjectHandler) GetDiff(c echo.Context) error {
	projectID := c.Param("id")

	files, err := h.db.GetFilesByProject(projectID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get files"})
	}

	if len(files) < 2 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Project must have both base and target files"})
	}

	var baseFile, targetFile *models.TranslationFile
	for i := range files {
		if files[i].FileType == "base" {
			baseFile = &files[i]
		} else if files[i].FileType == "target" {
			targetFile = &files[i]
		}
	}

	// Parse JSON files
	baseData, err := jsontools.ParseJSON([]byte(baseFile.Content))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to parse base file"})
	}
	targetData, err := jsontools.ParseJSON([]byte(targetFile.Content))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to parse target file"})
	}

	// Flatten
	baseFlat := jsontools.FlattenJSON(baseData, "")
	targetFlat := jsontools.FlattenJSON(targetData, "")

	// Compare
	diff := jsontools.CompareJSON(baseFlat, targetFlat)
	completion := jsontools.CompletionPercentage(baseFlat, targetFlat)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"missing_keys":     diff.MissingKeys,
		"extra_keys":       diff.ExtraKeys,
		"different_values": diff.DifferentValues,
		"completion":       completion,
	})
}

// UpdateTranslation handles POST /api/project/:id/translations
func (h *ProjectHandler) UpdateTranslation(c echo.Context) error {
	projectID := c.Param("id")

	// Parse form data manually since key names are dynamic
	if err := c.Request().ParseForm(); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	// Create a map from the form values
	req := make(map[string]string)
	for key, values := range c.Request().PostForm {
		if len(values) > 0 {
			req[key] = values[0]
		}
	}

	// INTERCEPT DEMO PROJECT
	if projectID == "demo-project" {
		return h.handleDemoProject(c, req, projectID)
	}

	return h.handleTranslationUpdate(c, projectID, req)
}

// ExportFile handles GET /api/project/:id/export
func (h *ProjectHandler) ExportFile(c echo.Context) error {
	projectID := c.Param("id")
	lang := c.QueryParam("lang")

	files, err := h.db.GetFilesByProject(projectID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get files"})
	}

	var targetFile *models.TranslationFile
	for i := range files {
		if files[i].LanguageCode == lang || (lang == "" && files[i].FileType == "target") {
			targetFile = &files[i]
			break
		}
	}

	if targetFile == nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "File not found"})
	}

	c.Response().Header().Set("Content-Type", "application/json")
	c.Response().Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.json", targetFile.LanguageCode))

	return c.String(http.StatusOK, targetFile.Content)
}

// AutoTranslate handles POST /api/project/:id/translate
func (h *ProjectHandler) AutoTranslate(c echo.Context) error {
	projectID := c.Param("id")

	return h.performAutoTranslate(c, projectID)
}

// Helper functions

func generateID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "" // or panic?
	}
	return hex.EncodeToString(b)
}

func generateAPIKey() (string, string) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", ""
	}
	key := hex.EncodeToString(b)

	// Hash the key for storage
	hash := sha256.Sum256([]byte(key))
	keyHash := hex.EncodeToString(hash[:])

	return key, keyHash
}

// makeTranslationSkeleton preserves the JSON structure but blanks out leaf values.
// - objects: recurse into fields
// - arrays: recurse into elements
// - primitives (string/number/bool/null): becomes "" (empty string)
func makeTranslationSkeleton(v any) any {
	switch t := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(t))
		for k, vv := range t {
			out[k] = makeTranslationSkeleton(vv)
		}
		return out
	case []any:
		out := make([]any, len(t))
		for i := range t {
			out[i] = makeTranslationSkeleton(t[i])
		}
		return out
	default:
		// leaf value placeholder for translations
		return ""
	}
}

// GetUserProjects handles GET /api/user/projects
func (h *ProjectHandler) GetUserProjects(c echo.Context) error {
	token := session.GetSessionToken(c)
	if token == "" {
		return c.JSON(http.StatusOK, []models.Project{})
	}

	projects, err := h.db.GetProjectsBySession(token, 10)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch projects"})
	}

	return c.JSON(http.StatusOK, projects)
}

// GetBaseTemplates handles GET /api/user/templates
func (h *ProjectHandler) GetBaseTemplates(c echo.Context) error {
	token := session.GetSessionToken(c)
	if token == "" {
		return c.JSON(http.StatusOK, []models.BaseTemplate{})
	}

	templates, err := h.db.GetBaseTemplatesBySession(token)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch templates"})
	}

	return c.JSON(http.StatusOK, templates)
}

// GetProjectBaseFile handles GET /api/project/:id/base - retrieves base file for reuse
func (h *ProjectHandler) GetProjectBaseFile(c echo.Context) error {
	projectID := c.Param("id")
	token := session.GetSessionToken(c)

	project, err := h.db.GetProject(projectID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Project not found"})
	}

	// Only allow reuse if user owns the project (session match) or if we implement public reusing later
	// For now, restrict to session owner for privacy if it's not locked, or just simple session check
	if project.SessionToken != token {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "Unauthorized"})
	}

	files, err := h.db.GetFilesByProject(projectID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get files"})
	}

	for _, f := range files {
		if f.FileType == "base" {
			return c.JSON(http.StatusOK, map[string]string{
				"content":       f.Content,
				"language_code": f.LanguageCode,
			})
		}
	}

	return c.JSON(http.StatusNotFound, map[string]string{"error": "Base file not found"})
}
