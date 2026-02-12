package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"

	"templui/internal/ai"
	"templui/internal/jsontools"
	"templui/internal/models"
)

func (h *ProjectHandler) performAutoTranslate(c echo.Context, projectID string) error {
	// Get files
	files, err := h.db.GetFilesByProject(projectID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get files"})
	}

	var baseFile, targetFile *models.TranslationFile
	for i := range files {
		if files[i].FileType == "base" {
			baseFile = &files[i]
		} else if files[i].FileType == "target" {
			targetFile = &files[i]
		}
	}

	if baseFile == nil || targetFile == nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Files not found"})
	}

	// Parse and flatten
	baseFlat, targetFlat, err := parseAndFlatten(baseFile, targetFile)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	// Identify missing translations
	missing := identifyMissing(baseFlat, targetFlat)
	slog.Info("Found missing translations", "count", len(missing), "project_id", projectID)

	if len(missing) == 0 {
		return c.JSON(http.StatusOK, map[string]string{"message": "Nothing to translate"})
	}

	if len(missing) > 50 {
		slog.Warn("Large batch of translations", "count", len(missing))
	}

	// Call AI
	aiClient := ai.NewOpenAIClient()
	translations, err := aiClient.Translate(baseFile.LanguageCode, targetFile.LanguageCode, missing)
	if err != nil {
		slog.Error("AI Translation failed", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("AI Translation failed: %v", err)})
	}

	// Update and save
	if err := h.updateAndSave(targetFile, targetFlat, translations); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to save translations"})
	}

	c.Response().Header().Set("HX-Refresh", "true")
	return c.JSON(http.StatusOK, map[string]string{"status": "success"})
}

func parseAndFlatten(baseFile, targetFile *models.TranslationFile) (map[string]string, map[string]string, error) {
	baseData, err := jsontools.ParseJSON([]byte(baseFile.Content))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse base file")
	}
	targetData, err := jsontools.ParseJSON([]byte(targetFile.Content))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse target file")
	}

	return jsontools.FlattenJSON(baseData, ""), jsontools.FlattenJSON(targetData, ""), nil
}

func identifyMissing(baseFlat, targetFlat map[string]string) map[string]string {
	missing := make(map[string]string)
	for k, v := range baseFlat {
		if targetFlat[k] == "" {
			missing[k] = v
		}
	}
	return missing
}

func (h *ProjectHandler) updateAndSave(targetFile *models.TranslationFile, targetFlat map[string]string, translations map[string]string) error {
	for k, v := range translations {
		targetFlat[k] = v
	}

	updatedData := jsontools.UnflattenJSON(targetFlat)
	updatedJSON, err := json.MarshalIndent(updatedData, "", "  ")
	if err != nil {
		return err
	}

	return h.db.UpdateFile(targetFile.ID, string(updatedJSON))
}
