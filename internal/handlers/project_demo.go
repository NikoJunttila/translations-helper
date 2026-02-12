package handlers

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"templui/internal/jsontools"
	"templui/ui/pages"
)

func (h *ProjectHandler) handleDemoProject(c echo.Context, req map[string]string, projectID string) error {
	for key, value := range req {
		// Basic validation simulation for demo
		// We define a map of base values matching the demo content in ExampleHandler
		demoBase := map[string]string{
			"welcome":         "Welcome to our application",
			"welcome_message": "Welcome back, {user}!",
			"login":           "Log in",
			"signup":          "Sign up",
			"about":           "About Us",
		}

		baseVal := demoBase[key]

		if err := jsontools.ValidatePlaceholders(baseVal, value); err != nil {
			c.Response().Header().Set("HX-Retarget", fmt.Sprintf("#field-%s", key))
			c.Response().Header().Set("HX-Reswap", "outerHTML")
			// Override hx-select to pick the whole field
			c.Response().Header().Set("HX-Reselect", fmt.Sprintf("#field-%s", key))
			return render(c, pages.TranslationField(key, baseVal, value, projectID, err.Error()))
		}

		return render(c, pages.TranslationField(key, baseVal, value, projectID, ""))
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "success"})
}
