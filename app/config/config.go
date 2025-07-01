package config

import (
	"context"
	"time"
	"users-service/app/controllers"
	"users-service/app/infrastructure"
	"users-service/app/middlewares"
	"users-service/app/repositories"
	"users-service/app/routes"
	"users-service/app/services"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func InitApp(settings *Settings) *gin.Engine {
	ctx := context.Background()

	logger := NewLogger(settings)

	entClient, driver, err := NewEntClient(ctx, logger, settings)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to initialize ent client")
	}

	redisClient := NewRedisClient(ctx, logger, settings)

	if err := CreateTopics(logger, settings, settings.KafkaTopics); err != nil {
		logger.Fatal().Msg("failed to initialize create kafka topics")
	}

	kafkaProducer, err := NewKafkaProducer(logger, settings)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to initialize kafka producer")
	}

	eventNotifier := infrastructure.NewEventNotifier(kafkaProducer, logger, settings.KafkaNotifierTopic)

	cacheStore := infrastructure.NewCacheStore(redisClient, logger, settings.EncryptionSecretKey)
	usersRepo := repositories.NewUsersRepository(cacheStore, entClient, driver, logger)

	authService := services.NewAuthService(usersRepo, eventNotifier, logger, settings.JwtPrivateKeyBase64, settings.JwtPublicKeyBase64, 10)
	diagnosticsService := services.NewDiagnosticsService(usersRepo, logger)
	usersService := services.NewUsersService(usersRepo, eventNotifier, logger)

	authController := controllers.NewAuthController(authService, usersService, logger)
	diagnosticsController := controllers.NewDiagnosticsController(authService, usersService, diagnosticsService, logger)
	usersController := controllers.NewUsersController(usersService, logger)

	middlewaresStore := middlewares.NewMiddlewares(cacheStore, logger, settings.ApiVersion, settings.Environment)
	tracker := middlewares.NewRequestTracker(40)

	router := gin.Default()

	router.
		Use(middlewaresStore.RequestAuthenticationMiddleware()).
		Use(middlewaresStore.UpdateCacheHitRatioMetrics(redisClient)).
		Use(middlewaresStore.UpdateServerMetrics()).
		Use(middlewaresStore.UpdateSystemMetrics())

	router.Use(cors.New(cors.Config{
		AllowOrigins:        []string{"*"},
		AllowMethods:        []string{"GET", "POST", "PUT", "DELETE", "PATCH"},
		AllowHeaders:        []string{"*"},
		ExposeHeaders:       []string{"*"},
		AllowCredentials:    true,
		MaxAge:              12 * time.Hour,
		AllowPrivateNetwork: false,
		AllowWebSockets:     true,
	}))

	api := router.Group("")

	routes.RegisterAuthRoutes("/auth", api, authController, tracker, middlewaresStore, logger)
	routes.RegisterUserRoutes("/users", api, usersController, authService, tracker, middlewaresStore, logger)
	routes.RegisterDiagnosticsRoutes("/diagnostics", api, diagnosticsController, tracker, middlewaresStore, logger)

	return router
}
