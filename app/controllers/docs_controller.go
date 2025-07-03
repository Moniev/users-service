package controllers

import (
	"users-service/app/services"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

type DocumentationController struct {
	DocumentationService services.DocumentationServiceInterface
	Logger               zerolog.Logger
}

type DocumentationControllerInterface interface {
}

var _ DocumentationControllerInterface = (*DocumentationController)(nil)

func (c *DocumentationController) Swagger(ctx *gin.Context) {

}

func (c *DocumentationController) OpenAPI(ctx *gin.Context) {

}

func (c *DocumentationController) Docs(ctx *gin.Context) {

}
