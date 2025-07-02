package controllers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	requests "users-service/app/models/requests"
	"users-service/app/models/responses"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

type StandardController interface {
	RespondFailure(ctx *gin.Context, statusCode int, errMsg string)
	RespondSuccess(ctx *gin.Context, statusCode int, data interface{})
	Log() *zerolog.Logger
}

type WebSocketController interface {
	StandardController
	InitWebSocketHelper(ctx *gin.Context) (*requests.WebSocketHelper, error)
}

func handleStandardRequest[T any, R any](
	ctx *gin.Context,
	c StandardController,
	serviceCall func(reqCtx context.Context, request *T) (R, error),
) {
	var req T
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.Log().Error().Err(err).Msg("Failed to bind JSON request")
		c.RespondFailure(ctx, http.StatusBadRequest, "invalid body request")
		return
	}

	type validatable interface {
		Valid() error
	}
	if v, ok := any(&req).(validatable); ok {
		if err := v.Valid(); err != nil {
			c.Log().Warn().Err(err).Msg("Invalid request data")
			c.RespondFailure(ctx, http.StatusBadRequest, "invalid request data: "+err.Error())
			return
		}
	}

	result, err := serviceCall(ctx.Request.Context(), &req)
	if err != nil {
		c.Log().Error().Err(err).Msg("Service call failed")
		c.RespondFailure(ctx, http.StatusUnprocessableEntity, err.Error())
		return
	}

	c.RespondSuccess(ctx, http.StatusOK, result)
}

func handleAuthenticatedRequest[T any, R any](
	ctx *gin.Context,
	c StandardController,
	serviceCall func(reqCtx context.Context, userID int, request *T) (R, error),
) {
	var req T
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.Log().Error().Err(err).Msg("Failed to bind JSON request")
		c.RespondFailure(ctx, http.StatusBadRequest, "invalid body request")
		return
	}

	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		c.Log().Error().Err(err).Msg("Failed to get UserID from context")
		c.RespondFailure(ctx, http.StatusUnauthorized, "invalid authentication context")
		return
	}

	type validatable interface {
		Valid() error
	}
	if v, ok := any(&req).(validatable); ok {
		if err := v.Valid(); err != nil {
			c.Log().Warn().Err(err).Msg("Invalid request data")
			c.RespondFailure(ctx, http.StatusBadRequest, "invalid request data: "+err.Error())
			return
		}
	}

	result, err := serviceCall(ctx.Request.Context(), userID, &req)
	if err != nil {
		c.Log().Error().Err(err).Msg("Service call failed")
		c.RespondFailure(ctx, http.StatusUnprocessableEntity, err.Error())
		return
	}

	c.RespondSuccess(ctx, http.StatusOK, result)
}

func handleWebSocketRequest[T any, R any](
	ctx *gin.Context,
	c WebSocketController,
	serviceCall func(reqCtx context.Context, request *T, reporter responses.Reporter) (R, error),
) {
	ws, err := c.InitWebSocketHelper(ctx)
	if err != nil {
		return
	}
	defer ws.Conn.Close()

	var req T
	if err := ws.ReadRequest(&req); err != nil {
		return
	}

	type validatable interface {
		Valid() error
	}
	if v, ok := any(&req).(validatable); ok {
		if err := v.Valid(); err != nil {
			ws.RespondFailure("Invalid credentials provided", err.Error())
			c.Log().Warn().Err(err).Msg("Received invalid request data via WebSocket")
			return
		}
	}

	progressChan := make(chan responses.Message, 10)
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		ws.ForwardProgress(progressChan)
	}()

	reporter := responses.NewProgressReporter(progressChan)

	result, err := serviceCall(ctx.Request.Context(), &req, reporter)
	if err != nil {
		c.Log().Error().Err(err).Msg("WebSocket process finished with an error")
	} else {
		c.Log().Info().Interface("result", result).Msg("WebSocket process completed successfully")
	}

	close(progressChan)
	wg.Wait()
}

func getUserIDFromContext(ctx *gin.Context) (int, error) {
	userIDVal, exists := ctx.Get("UserID")
	if !exists {
		return 0, errors.New("UserID not found in context")
	}

	userID, ok := userIDVal.(int)
	if !ok {
		return 0, fmt.Errorf("UserID in context is not of type int, got %T", userIDVal)
	}

	return userID, nil
}
