package registry

import (
	"reflect"
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

var ModelTypeRegistry = make(map[string]reflect.Type)

func init() {
	ModelTypeRegistry["requests.Device"] = reflect.TypeOf(&requests.Device{})
	ModelTypeRegistry["requests.Register"] = reflect.TypeOf(&requests.Register{})
	ModelTypeRegistry["requests.Login"] = reflect.TypeOf(&requests.Login{})
	ModelTypeRegistry["requests.Code"] = reflect.TypeOf(&requests.Code{})
	ModelTypeRegistry["requests.Mail"] = reflect.TypeOf(&requests.Mail{})
	ModelTypeRegistry["requests.ConfirmPasswordReset"] = reflect.TypeOf(&requests.ConfirmPasswordReset{})
	ModelTypeRegistry["requests.Empty"] = reflect.TypeOf(&requests.Empty{})
	ModelTypeRegistry["requests.Details"] = reflect.TypeOf(&requests.Details{})
	ModelTypeRegistry["requests.User"] = reflect.TypeOf(&requests.User{})
	ModelTypeRegistry["requests.Settings"] = reflect.TypeOf(&requests.Settings{})
	ModelTypeRegistry["requests.Password"] = reflect.TypeOf(&requests.Password{})

	ModelTypeRegistry["responses.ErrorResponse"] = reflect.TypeOf(&responses.ErrorResponse{})
	ModelTypeRegistry["responses.SuccessResponse"] = reflect.TypeOf(&responses.SuccessResponse{})
	ModelTypeRegistry["responses.Message"] = reflect.TypeOf(&responses.Message{})
	ModelTypeRegistry["responses.UserResponse"] = reflect.TypeOf(&responses.User{})
	ModelTypeRegistry["responses.ProgressReporter"] = reflect.TypeOf(&responses.ProgressReporter{})

	ModelTypeRegistry["events.BaseEvent"] = reflect.TypeOf(&events.BaseEvent{})
	ModelTypeRegistry["events.RegistrationEvent"] = reflect.TypeOf(&events.RegistrationEvent{})
	ModelTypeRegistry["events.SecondFactorEvent"] = reflect.TypeOf(&events.SecondFactorEvent{})
	ModelTypeRegistry["events.LoginEvent"] = reflect.TypeOf(&events.LoginEvent{})
	ModelTypeRegistry["events.VerificationEvent"] = reflect.TypeOf(&events.VerificationEvent{})
	ModelTypeRegistry["events.NotificationEvent"] = reflect.TypeOf(&events.NotificationEvent{})
	ModelTypeRegistry["events.UserActionEvent"] = reflect.TypeOf(&events.UserActionEvent{})
	ModelTypeRegistry["events.ResetPasswordEvent"] = reflect.TypeOf(&events.ResetPasswordEvent{})

	ModelTypeRegistry["utils.PermissionInfo"] = reflect.TypeOf(&utilsModels.PermissionInfo{})
	ModelTypeRegistry["utils.UserRoleInfo"] = reflect.TypeOf(&utilsModels.UserRoleInfo{})
	ModelTypeRegistry["utils.Claims"] = reflect.TypeOf(&utilsModels.Claims{})
	ModelTypeRegistry["utils.JWTValidationResult"] = reflect.TypeOf(&utilsModels.JWTValidationResult{})
	ModelTypeRegistry["utils.JWTResult"] = reflect.TypeOf(&utilsModels.JWTResult{})
	ModelTypeRegistry["utils.HashResult"] = reflect.TypeOf(&utilsModels.HashResult{})
}
