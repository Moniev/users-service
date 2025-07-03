package controllers

import (
	"net/http"
	"users-service/app/models/responses"
	"users-service/app/services"
	"users-service/app/utils"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

type DiagnosticsController struct {
	DiagnosticsService services.DiagnosticsServiceInterface
	Logger             zerolog.Logger
}

type DiagnosticsControllerInterface interface {
	RespondFailure(ctx *gin.Context, statusCode int, errMsg string)
	RespondSuccess(ctx *gin.Context, statusCode int, data interface{})
	ReadinessProbe(ctx *gin.Context)
	Log() *zerolog.Logger
	HealthProbe(ctx *gin.Context)
}

var _ DiagnosticsControllerInterface = (*DiagnosticsController)(nil)
var _ StandardController = (*DiagnosticsController)(nil)

func NewDiagnosticsController(
	diagnosticsService services.DiagnosticsServiceInterface,
	logger zerolog.Logger) *DiagnosticsController {

	return &DiagnosticsController{
		DiagnosticsService: diagnosticsService,
		Logger:             logger,
	}
}

func (c *DiagnosticsController) Log() *zerolog.Logger {
	return &c.Logger
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

// ReadinessProbe godoc
// @Summary      Readiness Probe
// @Description  Checks if the application is ready to serve traffic. This typically involves
// @Description  verifying connections to critical dependencies like databases or message queues.
// @Tags         diagnostics, health
// @Produce      json
// @Success      200 {object} responses.SuccessResponse{data=object{message=string}} "OK - Service is ready"
// @Failure      503 {object} responses.ErrorResponse "Service Unavailable - Service not ready"
// @Router       /api/v1/diagnostics/readiness [get]
func (c *DiagnosticsController) ReadinessProbe(ctx *gin.Context) {
	err := c.DiagnosticsService.CheckReadiness(ctx)
	if err != nil {
		c.Logger.Error().Err(err).Msg("Readiness probe failed")
		c.RespondFailure(ctx, http.StatusServiceUnavailable, "Service not ready: "+err.Error())
		return
	}

	c.RespondSuccess(ctx, http.StatusOK, gin.H{"message": "Service is ready"})
}

// HealthProbe godoc
// @Summary      Health Probe
// @Description  Checks if the application is alive and running. This is typically a lightweight check
// @Description  to ensure the process is active and responsive.
// @Tags         diagnostics, health
// @Produce      json
// @Success      200 {object} responses.SuccessResponse{data=object{message=string}} "OK - Service is healthy"
// @Failure      503 {object} responses.ErrorResponse "Service Unavailable - Service unhealthy"
// @Router       /api/v1/diagnostics/health [get]
func (c *DiagnosticsController) HealthProbe(ctx *gin.Context) {
	err := c.DiagnosticsService.CheckHealth(ctx)
	if err != nil {
		c.Logger.Error().Err(err).Msg("Health probe failed")
		c.RespondFailure(ctx, http.StatusServiceUnavailable, "Service unhealthy: "+err.Error())
		return
	}

	c.RespondSuccess(ctx, http.StatusOK, gin.H{"message": "Service is healthy"})
}
