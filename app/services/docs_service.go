package services

import "github.com/rs/zerolog"

type DocumentationService struct {
	Logger zerolog.Logger
}

type DocumentationServiceInterface interface {
}

var _ DocumentationServiceInterface = (*DocumentationService)(nil)
