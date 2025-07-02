package registry

import (
	"crypto"
	"crypto/x509"
	"encoding/pem"
	"log"
	"users-service/app/config"
	"users-service/app/controllers"
	"users-service/app/infrastructure"
	"users-service/app/models/ent"
	"users-service/app/repositories"
	"users-service/app/services"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

func NewTestAppContainer(
	testSettings *config.Settings,
	db *ent.Client,
	redis *redis.Client,
	kafkaProducer *kafka.Producer,
	logger zerolog.Logger,
	pubKey crypto.PublicKey,
	privKey crypto.PrivateKey) *AppContainer {

	cache := infrastructure.NewCacheStore(redis, logger, "a-32-byte-long-test-secret-key!!")
	repo := repositories.NewUsersRepository(cache, db, nil, logger)
	notifier := infrastructure.NewEventNotifier(kafkaProducer, logger, "notifications")

	privBytes, err := x509.MarshalPKCS8PrivateKey(privKey)
	if err != nil {
		log.Fatalf("failed to marshal private key: %v", err)
	}
	privPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privBytes})

	pubBytes, err := x509.MarshalPKIXPublicKey(pubKey)
	if err != nil {
		log.Fatalf("failed to marshal public key: %v", err)
	}
	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes})

	serviceContainer := &ServiceContainer{
		AuthService: services.NewAuthService(
			repo,
			notifier,
			logger,
			string(pubPEM),
			string(privPEM),
			10,
		),
		UsersService: services.NewUsersService(repo, notifier, logger),
	}

	controllerContainer := &ControllerContainer{
		AuthController: controllers.NewAuthController(
			serviceContainer.AuthService,
			serviceContainer.UsersService,
			logger,
		),
		UsersController: controllers.NewUsersController(
			serviceContainer.UsersService,
			logger,
		),
	}

	router := gin.New()
	RegisterAllRoutes(router, controllerContainer)

	return &AppContainer{
		Router:      router,
		Services:    serviceContainer,
		Controllers: controllerContainer,
		Functions:   &FunctionRegistry,
	}
}

func RegisterAllRoutes(router *gin.Engine, controllers *ControllerContainer) {
	apiV1 := router.Group("/api/v1")

	authRoutes := apiV1.Group("/auth")
	{
		authRoutes.POST("/register", controllers.AuthController.Register)
		authRoutes.POST("/login", controllers.AuthController.Login)
	}

	usersRoutes := apiV1.Group("/users")
	{
		usersRoutes.PATCH("/update", controllers.UsersController.UpdateUser)
		usersRoutes.PATCH("/details/update", controllers.UsersController.UpdateDetails)
		usersRoutes.PATCH("/settings/update", controllers.UsersController.UpdateSettings)
		usersRoutes.DELETE("/remove", controllers.UsersController.RemoveAccount)
	}
}
