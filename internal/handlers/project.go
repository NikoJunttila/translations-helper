package handlers

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"

	"templui/internal/database"
	"templui/internal/jsontools"
	"templui/internal/models"
	"templui/ui/pages"
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
	project := &models.Project{
		ID:        projectID,
		Name:      name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := h.db.CreateProject(project); err != nil {
		log.Error(err)
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

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"project_id": projectID,
		"api_key":    apiKey,
		"url":        fmt.Sprintf("/project/%s/edit?key=%s", projectID, apiKey),
	})
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
	baseData, _ := jsontools.ParseJSON([]byte(baseFile.Content))
	targetData, _ := jsontools.ParseJSON([]byte(targetFile.Content))

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

	// Get target file
	files, err := h.db.GetFilesByProject(projectID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get files"})
	}

	var targetFile *models.TranslationFile
	var baseFlat map[string]string
	for i := range files {
		if files[i].FileType == "target" {
			targetFile = &files[i]
		} else if files[i].FileType == "base" {
			// Get base file for rendering
			var baseData map[string]interface{}
			json.Unmarshal([]byte(files[i].Content), &baseData)
			baseFlat = jsontools.FlattenJSON(baseData, "")
		}
	}

	if targetFile == nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Target file not found"})
	}

	// Parse existing content
	var targetData map[string]interface{}
	json.Unmarshal([]byte(targetFile.Content), &targetData)

	// Flatten
	targetFlat := jsontools.FlattenJSON(targetData, "")

	// Update with new values
	for key, value := range req {
		// Validate placeholders
		baseVal := baseFlat[key]
		if err := jsontools.ValidatePlaceholders(baseVal, value); err != nil {
			// Return field with error
			// Since we want to update the whole field to show error, we need correct hx-swap-oob or just replace the field target?
			// The current hx-target on input is "previous .translation-label", which is JUST the label.
			// But for error message we need to show it below input or on input border.
			// The updated template now renders error message in the field component.
			// BUT the HTMX request on the input has hx-target="previous .translation-label".
			// If we return the whole field, HTMX will try to put the whole field INSIDE the label!
			// We need to change the hx-target if there is an error? Impossible.
			// We need OOB swap if valid vs invalid?

			// Fix: If we have an error, we want to replace the whole field container to show the error message and red border.
			// But the input on the page is configured to target THE LABEL.

			// We can use OOB swap to replace the whole field by ID.
			// field-KEY id is on the wrapper.
			c.Response().Header().Set("HX-Retarget", fmt.Sprintf("#field-%s", key))
			c.Response().Header().Set("HX-Reswap", "outerHTML")
			c.Response().Header().Set("HX-Reselect", fmt.Sprintf("#field-%s", key)) // Override hx-select to pick the whole field
			return render(c, pages.TranslationField(key, baseVal, value, projectID, err.Error()))
		}
		targetFlat[key] = value
	}

	// Unflatten back
	updatedData := jsontools.UnflattenJSON(targetFlat)

	// Convert to JSON
	updatedJSON, _ := json.MarshalIndent(updatedData, "", "  ")

	// Save to database
	if err := h.db.UpdateFile(targetFile.ID, string(updatedJSON)); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update file"})
	}

	// Return updated field HTML (just the label part if success, but we can't easily partially render)
	// Actually, if success, we configured hx-target="previous .translation-label".
	// The template renders the WHOLE field.
	// If we return the whole field, it puts the whole field inside the label. BAD.

	// We need to fix the template logic or return logic.
	// Since we changed hx-target to label, we should theoretically ONLY return the label part on success.
	// OR we use OOB for everything.

	for key := range req {
		// On success, we also use OOB to be safe and consistent, OR we assume the template change I made earlier was correct?
		// Wait, if I changed hx-target to label, but I return pages.TranslationField (whole component),
		// then yes, it puts the component inside the label. That's a BUG in my previous step that user verified worked?
		// Ah, previous step I used hx-select=".translation-label".
		// That SELECTS only the label from the response. So it works!

		// So on success: Return whole field, HTMX picks label, puts in label. Perfect.

		// On ERROR: We want to replace the INPUT (to show red border) AND show error text.
		// Taking just the label won't show red border on input or error text below it.
		// So for error, we must replace the WHOLE field.
		// We can override target using HX-Retarget header!

		return render(c, pages.TranslationField(key, baseFlat[key], targetFlat[key], projectID, ""))
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "success"})
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

// Helper functions

func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func generateAPIKey() (string, string) {
	b := make([]byte, 32)
	rand.Read(b)
	key := hex.EncodeToString(b)

	// Hash the key for storage
	hash := sha256.Sum256([]byte(key))
	keyHash := hex.EncodeToString(hash[:])

	return key, keyHash
}
