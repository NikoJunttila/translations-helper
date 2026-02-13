package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"templui/assets"
	"templui/internal/database"
	"templui/internal/handlers"
	"templui/internal/metrics"
	"templui/internal/session"
	"templui/migrations"
	"templui/utils"
)

func main() {
	InitDotEnv()

	// Initialize slog
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Initialize database
	db, err := database.NewDB()
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := db.Close(); err != nil {
			slog.Error("Failed to close database", "error", err)
		}
	}()
	if err := migrations.RunMigrations(db.GetConn()); err != nil {
		slog.Error("Failed to run migrations", "error", err)
		os.Exit(1)
	}

	mPort := os.Getenv("METRICS_PORT")
	if mPort == "" {
		slog.Error("failed to find metrics port")
		os.Exit(1)
	}
	if err := metrics.StartMetrics(fmt.Sprintf(":%s", mPort)); err != nil {
		slog.Error("failed to start metrics", "error", err)
		os.Exit(1)
	}

	// Initialize Tracing
	otlpEndpoint := os.Getenv("OTLP_ENDPOINT")
	if otlpEndpoint != "" {
		shutdown, err := metrics.InitTracer(context.Background(), otlpEndpoint)
		if err != nil {
			slog.Error("failed to initialize tracer", "error", err)
		} else {
			defer func() {
				if err := shutdown(context.Background()); err != nil {
					slog.Error("failed to shutdown tracer", "error", err)
				}
			}()
			slog.Info("Tracer initialized", "endpoint", utils.SanitizeLog(otlpEndpoint)) // #nosec G706
		}
	} else {
		slog.Warn("OTLP_ENDPOINT not set, tracing disabled")
	}

	e := echo.New()

	// ── Global middleware ──────────────────────────────────────────────
	// ── Global middleware ──────────────────────────────────────────────
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus:   true,
		LogURI:      true,
		LogMethod:   true,
		LogRemoteIP: true,
		LogLatency:  true,
		LogError:    true,
		HandleError: true, // forwards error to the HTTPErrorHandler
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			if v.Error == nil {
				slog.LogAttrs(context.Background(), slog.LevelInfo, "request",
					slog.String("method", v.Method),
					slog.String("uri", v.URI),
					slog.Int("status", v.Status),
					slog.String("remote_ip", v.RemoteIP),
					slog.Duration("latency", v.Latency),
				)
			} else {
				slog.LogAttrs(context.Background(), slog.LevelError, "request",
					slog.String("method", v.Method),
					slog.String("uri", v.URI),
					slog.Int("status", v.Status),
					slog.String("remote_ip", v.RemoteIP),
					slog.Duration("latency", v.Latency),
					slog.String("err", v.Error.Error()),
				)
			}
			return nil
		},
	}))
	e.Use(middleware.Recover())
	e.Use(session.SessionMiddleware())

	metrics.SetupEcho(e)

	SetupAssetsRoutes(e)

	// ── Silence common browser/dev noise ──────────────────────────────
	e.GET("/favicon.ico", func(c echo.Context) error {
		isDevelopment := os.Getenv("GO_ENV") != "production"
		if isDevelopment {
			return c.File("assets/favicon.ico")
		}
		file, err := assets.Assets.ReadFile("favicon.ico")
		if err != nil {
			return c.NoContent(http.StatusNotFound)
		}
		return c.Blob(http.StatusOK, "image/x-icon", file)
	})
	e.GET("/.well-known/*", func(c echo.Context) error { return c.NoContent(http.StatusNoContent) })

	// Initialize handlers
	homeHandler := handlers.NewHomeHandler(db)
	projectHandler := handlers.NewProjectHandler(db)
	editorHandler := handlers.NewEditorHandler(db)
	exampleHandler := handlers.NewExampleHandler()

	// ── Routes ──────────────────────────────────────────────────────
	// Home
	e.GET("/", homeHandler.Home)
	e.GET("/example", exampleHandler.Example)

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
		api.POST("/project/:id/translate", projectHandler.AutoTranslate)
		api.GET("/project/:id/export", projectHandler.ExportFile)

		// Session/User routes
		api.GET("/user/projects", projectHandler.GetUserProjects)
		api.GET("/user/templates", projectHandler.GetBaseTemplates)
		api.GET("/project/:id/base", projectHandler.GetProjectBaseFile)
	}

	portStr := os.Getenv("PORT")
	if portStr == "" {
		portStr = "8090"
	}
	// Parse as int to remove taint for logging
	port, err := strconv.Atoi(portStr)
	if err != nil {
		port = 8090
	}

	slog.Info("Server is running", "url", fmt.Sprintf("http://localhost:%d", port))
	e.HideBanner = true
	if err := e.Start(fmt.Sprintf(":%d", port)); err != nil && err != http.ErrServerClosed {
		slog.Error("Server shutdown", "error", err)
	}
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
