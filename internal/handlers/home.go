package handlers

import (
	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"

	"templui/internal/database"
	"templui/ui/pages"
)

type HomeHandler struct {
	db *database.DB
}

func NewHomeHandler(db *database.DB) *HomeHandler {
	return &HomeHandler{db: db}
}

// Home handles GET /
func (h *HomeHandler) Home(c echo.Context) error {
	projects, _ := h.db.ListProjects(10)
	return render(c, pages.Home(projects))
}

// render is a helper to render templ components
func render(c echo.Context, component templ.Component) error {
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
	return component.Render(c.Request().Context(), c.Response())
}
