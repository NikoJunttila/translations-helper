package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"sort"
	"time"

	"github.com/labstack/echo/v4"

	"templui/internal/database"
	"templui/internal/jsontools"
	"templui/internal/models"
	"templui/ui/pages"
)

type EditorHandler struct {
	db *database.DB
}

func NewEditorHandler(db *database.DB) *EditorHandler {
	return &EditorHandler{db: db}
}

// Editor handles GET /project/:id/edit
func (h *EditorHandler) Editor(c echo.Context) error {
	projectID := c.Param("id")
	viewMode := c.QueryParam("view") // "missing" or "full"

	// Get project
	project, err := h.db.GetProject(projectID)
	if err != nil {
		return c.String(http.StatusNotFound, "Project not found")
	}

	// Check if project is locked
	if project.IsLocked {
		// Check for valid authentication
		if !h.isAuthenticated(c, project) {
			// Redirect to auth page
			return c.Redirect(http.StatusFound, "/project/"+projectID+"/auth")
		}
	}

	// Get files
	files, err := h.db.GetFilesByProject(projectID)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load files")
	}

	if len(files) < 2 {
		return c.String(http.StatusBadRequest, "Project must have both base and target files")
	}

	var baseFile, targetFile *map[string]interface{}
	var baseLang, targetLang string

	for i := range files {
		var data map[string]interface{}
		json.Unmarshal([]byte(files[i].Content), &data)

		if files[i].FileType == "base" {
			baseFile = &data
			baseLang = files[i].LanguageCode
		} else {
			targetFile = &data
			targetLang = files[i].LanguageCode
		}
	}

	// Flatten both
	baseFlat := jsontools.FlattenJSON(*baseFile, "")
	targetFlat := jsontools.FlattenJSON(*targetFile, "")

	// Compare
	diff := jsontools.CompareJSON(baseFlat, targetFlat)

	// Prepare sorted keys for template
	var sortedKeys []string
	if viewMode == "missing" {
		// Show only missing keys (sorted)
		sortedKeys = diff.MissingKeys
		sort.Strings(sortedKeys)
	} else {
		// Show all keys (sorted)
		sortedKeys = make([]string, 0, len(baseFlat))
		for key := range baseFlat {
			sortedKeys = append(sortedKeys, key)
		}
		sort.Strings(sortedKeys)
	}

	// Reconstruct JSON for raw view
	nested := jsontools.UnflattenJSON(targetFlat)
	rawJSONBytes, _ := json.MarshalIndent(nested, "", "  ")
	rawJSON := string(rawJSONBytes)

	return render(c, pages.Editor(project, sortedKeys, baseFlat, targetFlat, rawJSON, baseLang, targetLang, viewMode == "missing"))
}

// ProjectAuth handles GET /project/:id/auth - shows auth page for locked projects
func (h *EditorHandler) ProjectAuth(c echo.Context) error {
	projectID := c.Param("id")

	// Get project
	project, err := h.db.GetProject(projectID)
	if err != nil {
		return c.String(http.StatusNotFound, "Project not found")
	}

	// If project is not locked, redirect to editor
	if !project.IsLocked {
		return c.Redirect(http.StatusFound, "/project/"+projectID+"/edit")
	}

	// Get error message if any
	errorMsg := c.QueryParam("error")

	return render(c, pages.ProjectAuth(project, errorMsg))
}

// VerifyProjectKey handles POST /project/:id/auth - verifies secret key
func (h *EditorHandler) VerifyProjectKey(c echo.Context) error {
	projectID := c.Param("id")
	secretKey := c.FormValue("secret_key")

	// Get project
	project, err := h.db.GetProject(projectID)
	if err != nil {
		return c.Redirect(http.StatusFound, "/project/"+projectID+"/auth?error=Project+not+found")
	}

	// Hash the provided key
	hash := sha256.Sum256([]byte(secretKey))
	providedHash := hex.EncodeToString(hash[:])

	// Compare with stored hash
	if providedHash != project.SecretKeyHash {
		return c.Redirect(http.StatusFound, "/project/"+projectID+"/auth?error=Invalid+secret+key")
	}

	// Set authentication cookie
	cookie := &http.Cookie{
		Name:     "project_auth_" + projectID,
		Value:    providedHash,
		Path:     "/",
		Expires:  time.Now().Add(24 * time.Hour * 30), // 30 days
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}
	c.SetCookie(cookie)

	// Redirect to editor
	return c.Redirect(http.StatusFound, "/project/"+projectID+"/edit")
}

// isAuthenticated checks if the user is authenticated for a locked project
func (h *EditorHandler) isAuthenticated(c echo.Context, project *models.Project) bool {
	projectID := project.ID

	// Check cookie
	cookie, err := c.Cookie("project_auth_" + projectID)
	if err == nil && cookie.Value == project.SecretKeyHash {
		return true
	}

	// Check query parameter (for direct links with key)
	secretKey := c.QueryParam("key")
	if secretKey != "" {
		hash := sha256.Sum256([]byte(secretKey))
		providedHash := hex.EncodeToString(hash[:])
		if providedHash == project.SecretKeyHash {
			// Set cookie for future requests
			cookie := &http.Cookie{
				Name:     "project_auth_" + projectID,
				Value:    providedHash,
				Path:     "/",
				Expires:  time.Now().Add(24 * time.Hour * 30),
				HttpOnly: true,
				SameSite: http.SameSiteStrictMode,
			}
			c.SetCookie(cookie)
			return true
		}
	}

	return false
}
