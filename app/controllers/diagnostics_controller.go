package controllers

import (
	"users-service/app/models/responses"
	"users-service/app/services"
	"users-service/app/utils"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

type DiagnosticsController struct {
	AuthService  services.AuthServiceInterface
	UsersService services.UsersServiceInterface
	Logger       zerolog.Logger
}

type DiagnosticsControllerInterface interface {
	RespondFailure(ctx *gin.Context, statusCode int, errMsg string)
	RespondSuccess(ctx *gin.Context, statusCode int, data interface{})
	ReadinessProbe(ctx *gin.Context)
	HealthProbe(ctx *gin.Context)
}

var _ DiagnosticsControllerInterface = (*DiagnosticsController)(nil)

func NewDiagnosticsController(
	authService services.AuthServiceInterface,
	usersService services.UsersServiceInterface,
	diagnosticsService services.DiagnosticsServiceInterface,
	logger zerolog.Logger) *DiagnosticsController {

	return &DiagnosticsController{
		AuthService:  authService,
		UsersService: usersService,
		Logger:       logger,
	}
}

func (c *DiagnosticsController) RespondFailure(ctx *gin.Context, statusCode int, errMsg string) {
	requestID := utils.GetRequestID(ctx)

	ctx.JSON(statusCode, responses.ErrorResponse{
		Status:    "failure",
		Error:     errMsg,
		RequestID: requestID,
	})
}

func (c *DiagnosticsController) RespondSuccess(ctx *gin.Context, statusCode int, data interface{}) {
	requestID := utils.GetRequestID(ctx)

	ctx.JSON(statusCode, responses.SuccessResponse{
		Status:    "success",
		Data:      data,
		RequestID: requestID,
	})
}

func (c *DiagnosticsController) ReadinessProbe(ctx *gin.Context) {

}

func (c *DiagnosticsController) HealthProbe(ctx *gin.Context) {

}
