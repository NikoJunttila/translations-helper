package handlers

import (
	"encoding/json"
	"net/http"
	"sort"

	"github.com/labstack/echo/v4"

	"templui/internal/database"
	"templui/internal/jsontools"
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
