package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Jiruu246/rms/internal/config"
	"github.com/Jiruu246/rms/internal/ent"
	"github.com/Jiruu246/rms/internal/handler"
	"github.com/Jiruu246/rms/internal/repos"
	"github.com/Jiruu246/rms/internal/services"
	"github.com/gin-gonic/gin"
)

type Middlewares struct {
	RestrictiveCORS func(origins []string) gin.HandlerFunc
	CORS            func() gin.HandlerFunc
	JWTMiddleware   func(secretKey []byte) gin.HandlerFunc
}
type Server struct {
	cfg         *config.Config
	client      *ent.Client
	engine      *gin.Engine
	srv         *http.Server
	middlewares Middlewares
}

func New(cfg *config.Config, client *ent.Client, middlewares Middlewares) *Server {
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()
	engine.Use(gin.Recovery())

	// Add CORS middleware at the top level
	if cfg.Env == "production" {
		// Use restrictive CORS for production with configured origins
		engine.Use(middlewares.RestrictiveCORS(cfg.AllowedOrigins))
	} else {
		// Use permissive CORS for development
		engine.Use(middlewares.CORS())
	}

	s := &Server{
		cfg:         cfg,
		client:      client,
		engine:      engine,
		middlewares: middlewares,
	}

	s.routes()

	s.srv = &http.Server{
		Addr:         fmt.Sprintf(":"+"%d", cfg.Port),
		Handler:      engine,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}
	return s
}

func (s *Server) routes() {
	// initialize repositories
	categoryRepo := repos.NewEntCategoryRepository(s.client)
	userRepo := repos.NewEntUserRepository(s.client)
	refreshTokenRepo := repos.NewEntRefreshTokenRepository(s.client)
	restaurantRepo := repos.NewEntRestaurantRepository(s.client)
	menuitemRepo := repos.NewEntMenuItemRepository(s.client)
	modifierRepo := repos.NewEntModifierRepository(s.client)
	modifierOptionRepo := repos.NewEntModifierOptionRepository(s.client)
	orderRepo := repos.NewEntOrderRepository(s.client)

	// initialize services
	categoryService := services.NewCategoryService(categoryRepo)
	authService := services.NewAuthService(userRepo, refreshTokenRepo)
	userService := services.NewUserService(userRepo)
	restaurantService := services.NewRestaurantService(restaurantRepo)
	menuItemService := services.NewMenuItemService(menuitemRepo)
	modifierService := services.NewModifierService(modifierRepo)
	modifierOptionService := services.NewModifierOptionService(modifierOptionRepo)
	orderService := services.NewOrderService(orderRepo, menuitemRepo, modifierOptionRepo)

	// initialize handlers
	categoryHandler := handler.NewCategoryHandler(categoryService)
	authHandler := handler.NewAuthHandler(authService, []byte(s.cfg.JWTSecret))
	userHandler := handler.NewUserHandler(userService)
	restaurantHandler := handler.NewRestaurantHandler(restaurantService)
	menuItemHandler := handler.NewMenuItemHandler(menuItemService)
	modifierHandler := handler.NewModifierHandler(modifierService)
	modifierOptionHandler := handler.NewModifierOptionHandler(modifierOptionService)
	orderHandler := handler.NewOrderHandler(orderService)

	// API routes

	api := s.engine.Group("/api")
	{
		//TODO: Not a great pattern, refactor later
		public := api.Group("/public")
		{
			public.POST("/order", orderHandler.CreateOrderPub)
		}

		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.Refresh)
			auth.POST("/logout", authHandler.Logout)

			google := auth.Group("/google")
			{
				google.POST("/login", authHandler.GoogleLogin)
				google.POST("/callback", authHandler.GoogleCallback)
			}
		}

		categories := api.Group("/categories")
		{
			// Apply JWT middleware to all category routes
			categories.Use(s.middlewares.JWTMiddleware([]byte(s.cfg.JWTSecret)))

			categories.POST("", categoryHandler.CreateCategory)
			categories.GET("", categoryHandler.GetCategories)
			categories.GET("/:id", categoryHandler.GetCategory)
			categories.PUT("/:id", categoryHandler.UpdateCategory)
			categories.DELETE("/:id", categoryHandler.DeleteCategory)
		}

		users := api.Group("/users")
		users.Use(s.middlewares.JWTMiddleware([]byte(s.cfg.JWTSecret)))
		{
			users.GET("/profile", userHandler.GetProfile)
			users.PUT("/profile", userHandler.UpdateProfile)
			users.DELETE("/profile", userHandler.DeleteAccount)
		}

		restaurants := api.Group("/restaurants")
		{
			// Apply JWT middleware to all restaurant routes
			restaurants.Use(s.middlewares.JWTMiddleware([]byte(s.cfg.JWTSecret)))

			restaurants.POST("", restaurantHandler.CreateRestaurant)
			restaurants.GET("", restaurantHandler.GetRestaurants)
			restaurants.GET("/:id", restaurantHandler.GetRestaurant)
			restaurants.PUT("/:id", restaurantHandler.UpdateRestaurant)
			restaurants.DELETE("/:id", restaurantHandler.DeleteRestaurant)
		}

		menuItems := api.Group("/menu-items")
		{
			// Apply JWT middleware to all menu item routes
			menuItems.Use(s.middlewares.JWTMiddleware([]byte(s.cfg.JWTSecret)))

			menuItems.POST("", menuItemHandler.CreateMenuItem)
			menuItems.GET("", menuItemHandler.GetMenuItems)
			menuItems.GET("/:id", menuItemHandler.GetMenuItem)
			menuItems.PUT("/:id", menuItemHandler.UpdateMenuItem)
			menuItems.DELETE("/:id", menuItemHandler.DeleteMenuItem)
		}

		modifiers := api.Group("/modifiers")
		{
			// Apply JWT middleware to all modifier routes
			modifiers.Use(s.middlewares.JWTMiddleware([]byte(s.cfg.JWTSecret)))

			modifiers.POST("", modifierHandler.CreateModifier)
			modifiers.GET("", modifierHandler.GetAllModifiers)
			modifiers.GET(":id", modifierHandler.GetModifier)
			modifiers.PATCH(":id", modifierHandler.UpdateModifier)
			modifiers.DELETE(":id", modifierHandler.DeleteModifier)

			// Modifier options endpoints
			options := modifiers.Group("/options")
			{
				options.POST("", modifierOptionHandler.CreateModifierOption)
				options.GET("", modifierOptionHandler.GetAllModifierOptions)
				options.GET(":id", modifierOptionHandler.GetModifierOption)
				options.PATCH(":id", modifierOptionHandler.UpdateModifierOption)
				options.DELETE(":id", modifierOptionHandler.DeleteModifierOption)
			}
		}

		orders := api.Group("/orders")
		{
			orders.Use(s.middlewares.JWTMiddleware([]byte(s.cfg.JWTSecret)))

			orders.POST("", orderHandler.CreateOrderPub)
			orders.GET("/:id", orderHandler.GetOrder)
			orders.GET("", orderHandler.GetOrders)
			orders.PATCH("/:id", orderHandler.UpdateOrder)
			orders.DELETE("/:id", orderHandler.DeleteOrder)
		}
	}
}

// Start runs the HTTP server.
func (s *Server) Start() error {
	fmt.Printf("listening on %s\n", s.srv.Addr)
	return s.srv.ListenAndServe()
}

// Shutdown gracefully stops the server.
func (s *Server) Shutdown(ctx context.Context) error {
	// allow in-flight requests a short time to finish
	ctxTimeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	return s.srv.Shutdown(ctxTimeout)
}

// Engine returns the gin engine for testing purposes
func (s *Server) Engine() *gin.Engine {
	return s.engine
}
