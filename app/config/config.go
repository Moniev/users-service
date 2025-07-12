package config

import (
	"context"
	"fmt"
	"strings"
	"time"
	"users-service/app/controllers"
	"users-service/app/infrastructure"
	"users-service/app/middlewares"
	"users-service/app/models/handlers"
	"users-service/app/repositories"
	"users-service/app/routes"
	"users-service/app/services"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/mvrilo/go-redoc"
	ginredoc "github.com/mvrilo/go-redoc/gin"
	"github.com/prometheus/client_golang/prometheus"
)

func RegisterMetrics() {
	prometheus.MustRegister(middlewares.HttpRequestCountWithPath)
	prometheus.MustRegister(middlewares.HttpRequestDuration)
	prometheus.MustRegister(middlewares.HttpRequestStatusCode)
	prometheus.MustRegister(middlewares.HttpRequestSize)
	prometheus.MustRegister(middlewares.HttpResponseSize)
	prometheus.MustRegister(middlewares.ActiveRequests)
	prometheus.MustRegister(middlewares.AuthFailures)
	prometheus.MustRegister(middlewares.BlockedIPs)
	prometheus.MustRegister(middlewares.DBQueryDuration)
	prometheus.MustRegister(middlewares.CacheMisses)
	prometheus.MustRegister(middlewares.CacheHits)
	prometheus.MustRegister(middlewares.CPUUsage)
	prometheus.MustRegister(middlewares.MemoryUsage)
	prometheus.MustRegister(middlewares.DiskUsage)
	prometheus.MustRegister(middlewares.ThreadCount)
	prometheus.MustRegister(middlewares.ErrorRate)
	prometheus.MustRegister(middlewares.SuccessRate)
}

func NewApp(settings *Settings) *gin.Engine {
	ctx := context.Background()
	logger := NewLogger(settings)

	privateKey, publicKey, err := LoadJWTSigningKeys(settings)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to load JWT signing keys")
	}

	entClient, driver, err := NewEntClient(ctx, logger, settings)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to initialize ent client")
	}

	redisClient := NewRedisClient(ctx, logger, settings)

	var topics []string
	if settings.KafkaTopics != "" {
		topics = strings.Split(settings.KafkaTopics, ",")
	}

	if err := CreateTopics(logger, settings, topics); err != nil {
		logger.Fatal().Msg("failed to initialize create kafka topics")
	}

	kafkaProducer, err := NewKafkaProducer(logger, settings)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to initialize kafka producer")
	}

	cacheStore := infrastructure.NewCacheStore(redisClient, logger, settings.EncryptionSecretKey)
	usersRepo := repositories.NewUsersRepository(cacheStore, entClient, driver, logger)
	consumerManager := infrastructure.NewConsumerManager(logger)

	for _, topic := range topics {
		groupID := fmt.Sprintf("users-service-%s-group", topic)
		var handler handlers.MessageHandler

		switch topic {
		case "user.events", "user.actions":
			handler = services.NewUserActionHandler(usersRepo, logger)
		case "user.subscription":
			handler = services.NewSubscriptionActionHandler(usersRepo, logger)
		default:
			logger.Warn().Str("topic", topic).Msg("No handler registered for topic")
		}

		if handler != nil {
			consumerWrapper, err := NewKafkaConsumer(groupID, topic, handler, logger, settings)
			if err != nil {
				logger.Fatal().Err(err).Str("topic", topic).Msg("Failed to create consumer")
			}

			consumerManager.Register(topic, consumerWrapper)
		}
	}

	eventNotifier := infrastructure.NewEventNotifier(kafkaProducer, logger, settings.KafkaNotifierTopic, time.Second*15)

	authService := services.NewAuthService(usersRepo, eventNotifier, logger, privateKey, publicKey, 10)
	diagnosticsService := services.NewDiagnosticsService(usersRepo, cacheStore, eventNotifier, consumerManager, logger)
	usersService := services.NewUsersService(usersRepo, eventNotifier, logger)

	authController := controllers.NewAuthController(authService, usersService, logger)
	diagnosticsController := controllers.NewDiagnosticsController(diagnosticsService, logger)
	usersController := controllers.NewUsersController(usersService, logger)

	middlewaresStore := middlewares.NewMiddlewares(cacheStore, logger, settings.ApiVersion, settings.Environment)
	tracker := middlewares.NewRequestTracker(40)

	router := gin.Default()

	doc := redoc.Redoc{
		Title:       "Factory Chainline, Users Service API",
		Description: "Main users management microservice responsible for use related operations.",
		SpecFile:    "./swagger.json",
		SpecPath:    "/docs/swagger.json",
		DocsPath:    "/redoc",
	}

	router.
		Use(middlewaresStore.RequestAuthenticationMiddleware()).
		Use(middlewaresStore.UpdateCacheHitRatioMetrics(redisClient)).
		Use(middlewaresStore.UpdateServerMetrics()).
		Use(middlewaresStore.UpdateSystemMetrics()).
		Use(ginredoc.New(doc))

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
	RegisterMetrics()

	routes.RegisterAuthRoutes("/auth", api, authController, tracker, middlewaresStore, logger)
	routes.RegisterUserRoutes("/users", api, usersController, authService, tracker, middlewaresStore, logger)
	routes.RegisterDiagnosticsRoutes("/diagnostics", api, diagnosticsController, tracker, middlewaresStore, logger)
	routes.RegisterMonitoringRoutes("/monitoring", api, tracker, middlewaresStore, logger)
	routes.RegisterDocumentationRoutes("/docs", api, authService, tracker, middlewaresStore, logger)
	routes.RegisterSwaggerRoutes("/swagger", api, tracker, middlewaresStore, logger)

	return router
}
