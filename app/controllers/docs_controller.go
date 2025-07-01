package controllers

import (
	"users-service/app/services"

	"github.com/rs/zerolog"
)

type DocumentationController struct {
	DocumentationService services.DocumentationServiceInterface
	Logger               zerolog.Logger
}

type DocumentationControllerInterface interface {
}

var _ DocumentationControllerInterface = (*DocumentationController)(nil)
