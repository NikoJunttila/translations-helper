package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"templui/assets"
	"templui/internal/database"
	"templui/internal/handlers"
	"templui/internal/metrics"
	"templui/migrations"
)

func main() {
	InitDotEnv()

	// Initialize database
	db, err := database.NewDB()
	if err != nil {
		fmt.Printf("Failed to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()
	if err := migrations.RunMigrations(db.GetConn()); err != nil {
		log.Fatal(err)
	}
	if err := metrics.StartMetrics(":8081"); err != nil {
		log.Fatal(err)
	}

	e := echo.New()

	// ── Global middleware ──────────────────────────────────────────────
	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())

	metrics.SetupEcho(e)

	SetupAssetsRoutes(e)

	// ── Silence common browser/dev noise ──────────────────────────────
	e.GET("/favicon.ico", func(c echo.Context) error { return c.NoContent(http.StatusNoContent) })
	e.GET("/.well-known/*", func(c echo.Context) error { return c.NoContent(http.StatusNoContent) })

	// Initialize handlers
	homeHandler := handlers.NewHomeHandler(db)
	projectHandler := handlers.NewProjectHandler(db)
	editorHandler := handlers.NewEditorHandler(db)

	// ── Routes ──────────────────────────────────────────────────────
	// Home
	e.GET("/", homeHandler.Home)

	// Editor
	e.GET("/project/:id/edit", editorHandler.Editor)
	e.GET("/project/:id/auth", editorHandler.ProjectAuth)
	e.POST("/project/:id/auth", editorHandler.VerifyProjectKey)

	// API Routes
	api := e.Group("/api")
	{
		api.POST("/project", projectHandler.CreateProject)
		api.GET("/project/:id/diff", projectHandler.GetDiff)
		api.POST("/project/:id/translations", projectHandler.UpdateTranslation)
		api.GET("/project/:id/export", projectHandler.ExportFile)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8090"
	}

	fmt.Printf("Server is running on http://localhost:%s\n", port)
	e.Logger.Fatal(e.Start(":" + port))
}

func InitDotEnv() {
	if err := godotenv.Load(); err != nil {
		fmt.Println("No .env file found, using environment variables")
	}
}

func SetupAssetsRoutes(e *echo.Echo) {
	isDevelopment := os.Getenv("GO_ENV") != "production"

	if isDevelopment {
		// Dev: serve local assets with no cache
		e.GET("/assets/*", echo.WrapHandler(
			http.StripPrefix("/assets/",
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Cache-Control", "no-store")
					http.FileServer(http.Dir("./assets")).ServeHTTP(w, r)
				}),
			),
		))
	} else {
		// Prod: serve embedded assets
		e.GET("/assets/*", echo.WrapHandler(
			http.StripPrefix(
				"/assets/",
				http.FileServer(http.FS(assets.Assets)),
			),
		))
	}
}
