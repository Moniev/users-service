package registry

import (
	"users-service/app/controllers"
	"users-service/app/middlewares"
	"users-service/app/models/events"
	"users-service/app/models/requests"
	"users-service/app/models/responses"
	utilsModels "users-service/app/models/utils"
	"users-service/app/repositories"
	"users-service/app/services"

	"github.com/gin-gonic/gin"
)

type AppContainer struct {
	Router      *gin.Engine
	Middlewares *middlewares.Middlewares
	Services    *ServiceContainer
	Controllers *ControllerContainer
	Models      *ModelsContainer
	Functions   *FunctionContainer
}

type ServiceContainer struct {
	AuthService          services.AuthServiceInterface
	UsersService         services.UsersServiceInterface
	DiagnosticsService   services.DiagnosticsServiceInterface
	DocumentationService services.DocumentationServiceInterface
}

type ControllerContainer struct {
	AuthController          controllers.AuthControllerInterface
	UsersController         controllers.UsersControllerInterface
	DiagnosticsController   controllers.DiagnosticsControllerInterface
	DocumentationController controllers.DocumentationControllerInterface
}

type RepositoryContainer struct {
	UsersRepository *repositories.UsersRepository
}

type ModelsContainer struct {
	BaseEvent          events.BaseEvent
	RegistrationEvent  events.RegistrationEvent
	SecondFactorEvent  events.SecondFactorEvent
	LoginEvent         events.LoginEvent
	VerificationEvent  events.VerificationEvent
	NotificationEvent  events.NotificationEvent
	UserActionEvent    events.UserActionEvent
	ResetPasswordEvent events.ResetPasswordEvent

	Device               requests.Device
	Register             requests.Register
	Login                requests.Login
	Code                 requests.Code
	Mail                 requests.Mail
	ConfirmPasswordReset requests.ConfirmPasswordReset
	Empty                requests.Empty
	Details              requests.Details
	UserRequest          requests.User
	Settings             requests.Settings
	Password             requests.Password

	Error   responses.ErrorResponse
	Success responses.SuccessResponse
	Message responses.Message

	UserResponse responses.User

	ProgressReporter responses.ProgressReporter

	PermissionInfo      utilsModels.PermissionInfo
	UserRoleInfo        utilsModels.UserRoleInfo
	Claims              utilsModels.Claims
	JWTValidationResult utilsModels.JWTValidationResult
	JWTResult           utilsModels.JWTResult
	HashResult          utilsModels.HashResult
}
