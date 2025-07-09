package main

import (
	"github.com/labstack/echo/v4"
	"course_management/services"
)

// ServiceMiddleware creates Echo middleware that injects services into context
func ServiceMiddleware(container services.ServiceContainer) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Inject services into context
			c.Set("services", container)
			c.Set("courseService", container.CourseService())
			c.Set("authService", container.AuthService())
			c.Set("sessionService", container.SessionService())
			c.Set("reviewService", container.ReviewService())
			
			return next(c)
		}
	}
}

// Helper functions to extract services from Echo context
func GetCourseService(c echo.Context) services.CourseService {
	service, ok := c.Get("courseService").(services.CourseService)
	if !ok {
		panic("CourseService not found in context - ensure ServiceMiddleware is used")
	}
	return service
}

func GetAuthService(c echo.Context) services.AuthService {
	service, ok := c.Get("authService").(services.AuthService)
	if !ok {
		panic("AuthService not found in context - ensure ServiceMiddleware is used")
	}
	return service
}

func GetSessionService(c echo.Context) services.SessionService {
	service, ok := c.Get("sessionService").(services.SessionService)
	if !ok {
		panic("SessionService not found in context - ensure ServiceMiddleware is used")
	}
	return service
}

func GetReviewService(c echo.Context) services.ReviewService {
	service, ok := c.Get("reviewService").(services.ReviewService)
	if !ok {
		panic("ReviewService not found in context - ensure ServiceMiddleware is used")
	}
	return service
}

func GetServices(c echo.Context) services.ServiceContainer {
	servicesContainer, ok := c.Get("services").(services.ServiceContainer)
	if !ok {
		panic("ServiceContainer not found in context - ensure ServiceMiddleware is used")
	}
	return servicesContainer
}

// Integration example for main.go
func ExampleServiceIntegration() {
	/*
	
	// In your main.go file:
	
	func main() {
		// Load configuration
		config := LoadConfig()
		
		// Initialize database
		if err := InitDatabase(); err != nil {
			log.Fatal("Failed to initialize database:", err)
		}
		
		// Create service configuration
		serviceConfig := services.ServiceConfig{
			DatabaseURL: config.DatabaseURL,
			AuthConfig: services.AuthConfig{
				GoogleClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
				GoogleClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
				RedirectURL:        os.Getenv("GOOGLE_REDIRECT_URL"),
			},
		}
		
		// Initialize service container
		container := services.NewServiceContainer(GetDB(), serviceConfig)
		defer container.Close()
		
		// Create Echo instance
		e := echo.New()
		e.Renderer = NewTemplates()
		
		// Add middleware
		e.Use(middleware.Logger())
		e.Use(middleware.Recover())
		e.Use(ServiceMiddleware(container))
		
		// Create handlers (using either old or new pattern)
		handlers := NewHandlers() // or NewServiceLayerHandlers()
		
		// Setup routes
		e.GET("/", handlers.Home)
		e.POST("/create-course", handlers.CreateCourse)
		// ... other routes
		
		// Start server
		e.Logger.Fatal(e.Start(":" + config.Port))
	}
	
	*/
}