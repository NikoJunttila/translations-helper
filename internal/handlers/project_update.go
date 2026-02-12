package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"

	"templui/internal/jsontools"
	"templui/internal/models"
	"templui/ui/pages"
)

func (h *ProjectHandler) handleTranslationUpdate(c echo.Context, projectID string, req map[string]string) error {
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
			var baseData map[string]interface{}
			if err := json.Unmarshal([]byte(files[i].Content), &baseData); err != nil {
				slog.Warn("Failed to parse base file", "error", err)
			}
			baseFlat = jsontools.FlattenJSON(baseData, "")
		}
	}

	if targetFile == nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Target file not found"})
	}

	// Parse existing content
	var targetData map[string]interface{}
	if err := json.Unmarshal([]byte(targetFile.Content), &targetData); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to parse target file"})
	}

	// Flatten
	targetFlat := jsontools.FlattenJSON(targetData, "")

	// Update with new values
	for key, value := range req {
		// Validate placeholders
		baseVal := baseFlat[key]
		if err := jsontools.ValidatePlaceholders(baseVal, value); err != nil {
			// interpolated string error here. This works but when redoing it we skip this.
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
	updatedJSON, err := json.MarshalIndent(updatedData, "", "  ")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to marshal JSON"})
	}

	// Save to database
	if err := h.db.UpdateFile(targetFile.ID, string(updatedJSON)); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update file"})
	}

	for key := range req {
		// can't we just swap the whole input or div around it? this way we can easily replace old error message.
		return render(c, pages.TranslationField(key, baseFlat[key], targetFlat[key], projectID, ""))
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "success"})
}
