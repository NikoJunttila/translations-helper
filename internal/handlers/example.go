package handlers

import (
	"sort"
	"templui/ui/pages"

	"github.com/labstack/echo/v4"
)

type ExampleHandler struct{}

func NewExampleHandler() *ExampleHandler {
	return &ExampleHandler{}
}

func (h *ExampleHandler) Example(c echo.Context) error {
	// Dummy data for the example
	base := map[string]string{
		"welcome":         "Welcome to our application",
		"welcome_message": "Welcome back, {user}!",
		"login":           "Log in",
		"signup":          "Sign up",
		"about":           "About Us",
	}

	target := map[string]string{
		"welcome":         "Bienvenido a nuestra aplicación",
		"welcome_message": "¡Bienvenido de nuevo, {user}!",
		"login":           "", // Missing translation
		"signup":          "Registrarse",
		"about":           "", // Missing translation
	}

	// Filter based on view
	view := c.QueryParam("view")
	showingMissingOnly := view == "missing"

	// Sort keys for consistent order
	var keys []string
	for k := range base {
		if showingMissingOnly && target[k] != "" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)

	return render(c, pages.Example(keys, base, target, showingMissingOnly))
}
