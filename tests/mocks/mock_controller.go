package mocks

import (
	"errors"
	"users-service/app/models/requests"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/mock"
)

type MockStandardController struct {
	mock.Mock
}

func (m *MockStandardController) InitWebSocketHelper(ctx *gin.Context) (*requests.WebSocketHelper, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*requests.WebSocketHelper), args.Error(1)
}

func (m *MockStandardController) RespondFailure(ctx *gin.Context, statusCode int, errMsg string) {
	m.Called(ctx, statusCode, errMsg)
	ctx.JSON(statusCode, gin.H{"error": errMsg})
}

func (m *MockStandardController) RespondSuccess(ctx *gin.Context, statusCode int, data interface{}) {
	m.Called(ctx, statusCode, data)
	ctx.JSON(statusCode, gin.H{"data": data})
}

func (m *MockStandardController) Log() *zerolog.Logger {
	args := m.Called()
	return args.Get(0).(*zerolog.Logger)
}

type TestRequest struct {
	Name string `json:"name"`
}

type TestRequestWithValidation struct {
	Name string `json:"name"`
}

func (r *TestRequestWithValidation) Valid() error {
	if r.Name == "invalid" {
		return errors.New("name is invalid")
	}
	return nil
}

type MockWebSocketController struct {
	MockStandardController
	mock.Mock
}

func (m *MockWebSocketController) InitWebSocketHelper(ctx *gin.Context) (*requests.WebSocketHelper, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*requests.WebSocketHelper), args.Error(1)
}

func (m *MockWebSocketController) Log() *zerolog.Logger {
	args := m.Called()
	return args.Get(0).(*zerolog.Logger)
}
