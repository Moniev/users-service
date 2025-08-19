package controllers

import (
	"context"
	requests "users-service/app/models/requests"
	"users-service/app/models/responses"
	"users-service/app/services"
	"users-service/app/utils"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

type UsersController struct {
	UsersService services.UsersServiceInterface
	Logger       zerolog.Logger
}

type UsersControllerInterface interface {
	RespondFailure(ctx *gin.Context, statusCode int, errMsg string)
	RespondSuccess(ctx *gin.Context, statusCode int, data interface{})

	UpdateUser(ctx *gin.Context)
	UpdateDetails(ctx *gin.Context)
	UpdateSettings(ctx *gin.Context)

	RemoveAccount(ctx *gin.Context)
}

var _ UsersControllerInterface = (*UsersController)(nil)
var _ StandardController = (*UsersController)(nil)

func NewUsersController(
	usersService services.UsersServiceInterface,
	logger zerolog.Logger) *UsersController {

	return &UsersController{
		UsersService: usersService,
		Logger:       logger,
	}
}

func (c *UsersController) Log() *zerolog.Logger {
	return &c.Logger
}

func (c *UsersController) RespondFailure(ctx *gin.Context, statusCode int, errMsg string) {
	requestID := utils.GetRequestID(ctx)

	ctx.JSON(statusCode, responses.ErrorResponse{
		Status:    "failure",
		Error:     errMsg,
		RequestID: requestID,
	})
}

func (c *UsersController) RespondSuccess(ctx *gin.Context, statusCode int, data interface{}) {
	requestID := utils.GetRequestID(ctx)

	ctx.JSON(statusCode, responses.SuccessResponse{
		Status:    "success",
		Data:      data,
		RequestID: requestID,
	})
}

// UpdateUser godoc
// @Summary      Update user's basic information
// @Description  Updates the authenticated user's basic profile information (e.g., name, email).
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        Authorization header    string            true  "Bearer token"
// @Param        body      body      requests.User  true  "User data to update"
// @Success      200       {object}  responses.SuccessResponse{data=responses.UserPrivate}  "OK - User updated successfully"
// @Failure      400       {object}  responses.ErrorResponse "Bad Request - Invalid input data"
// @Failure      401       {object}  responses.ErrorResponse "Unauthorized - Invalid or missing token"
// @Failure      422       {object}  responses.ErrorResponse "Unprocessable Entity - Update failed"
// @Router       /api/v1/users [patch]
func (c *UsersController) UpdateUser(ctx *gin.Context) {
	handleAuthenticatedRequest(ctx, c, func(reqCtx context.Context, userID int, req *requests.User) (*responses.UserPrivate, error) {
		user, err := c.UsersService.UpdateUser(reqCtx, userID, req)
		if err != nil {
			return nil, err
		}

		return responses.FromUserPrivate(user), nil
	})
}

// UpdateDetails godoc
// @Summary      Update user's details
// @Description  Updates the authenticated user's additional details.
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        Authorization header    string            true  "Bearer token"
// @Param        body      body      requests.Details  true  "User details to update"
// @Success      200       {object}  responses.SuccessResponse{data=responses.UserPrivate}  "OK - User details updated successfully"
// @Failure      400       {object}  responses.ErrorResponse "Bad Request - Invalid input data"
// @Failure      401       {object}  responses.ErrorResponse "Unauthorized - Invalid or missing token"
// @Failure      422       {object}  responses.ErrorResponse "Unprocessable Entity - Update failed"
// @Router       /api/v1/users/details [patch]
func (c *UsersController) UpdateDetails(ctx *gin.Context) {
	handleAuthenticatedRequest(ctx, c, func(reqCtx context.Context, userID int, req *requests.Details) (*responses.UserPrivate, error) {
		user, err := c.UsersService.UpdateDetails(reqCtx, userID, req)
		if err != nil {
			return nil, err
		}

		return responses.FromUserPrivate(user), nil
	})
}

// UpdateSettings godoc
// @Summary      Update user's settings
// @Description  Updates the authenticated user's application settings (e.g., 2FA, night mode).
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        Authorization header    string            true  "Bearer token"
// @Param        body      body      requests.Settings  true  "User settings to update"
// @Success      200       {object}  responses.SuccessResponse{data=responses.UserPrivate}  "OK - User settings updated successfully"
// @Failure      400       {object}  responses.ErrorResponse "Bad Request - Invalid input data"
// @Failure      401       {object}  responses.ErrorResponse "Unauthorized - Invalid or missing token"
// @Failure      422       {object}  responses.ErrorResponse "Unprocessable Entity - Update failed"
// @Router       /api/v1/users/settings [patch]
func (c *UsersController) UpdateSettings(ctx *gin.Context) {
	handleAuthenticatedRequest(ctx, c, func(reqCtx context.Context, userID int, req *requests.Settings) (*responses.UserPrivate, error) {
		user, err := c.UsersService.UpdateSettings(reqCtx, userID, req)
		if err != nil {
			return nil, err
		}

		return responses.FromUserPrivate(user), nil
	})
}

// RemoveAccount godoc
// @Summary      Remove user account
// @Description  Permanently removes the authenticated user's account. This action is irreversible.
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        Authorization header    string            true  "Bearer token"
// @Success      200       {object}  responses.SuccessResponse{data=string}  "OK - Account removed successfully"
// @Failure      401       {object}  responses.ErrorResponse "Unauthorized - Invalid or missing token"
// @Failure      422       {object}  responses.ErrorResponse "Unprocessable Entity - Removal failed"
// @Router       /api/v1/users [delete]
func (c *UsersController) RemoveAccount(ctx *gin.Context) {
	handleAuthenticatedRequest(ctx, c, func(reqCtx context.Context, userID int, req *requests.Empty) (string, error) {
		err := c.UsersService.RemoveAccount(reqCtx, userID)
		if err != nil {
			return "failed to remove user", err
		}

		return "successfully removed user", nil
	})
}

// UpdateEntrepreneurDetails godoc
// @Summary      Updates user entrepreneur details
// @Description  Updates existing entrepreneur details if already created. Other way creates new EntrepreneurDetails object.
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        Authorization header    string            true  "Bearer token"
// @Success      200       {object}  responses.SuccessResponse{data=string}  "OK - Account removed successfully"
// @Failure      401       {object}  responses.ErrorResponse "Unauthorized - Invalid or missing token"
// @Failure      422       {object}  responses.ErrorResponse "Unprocessable Entity - Update failed"
// @Router       /api/v1/users/details/entrepreneur [put]
func (c *UsersController) UpdateEntrepreneurDetails(ctx *gin.Context) {
	handleAuthenticatedRequest(ctx, c, func(reqCtx context.Context, userID int, req *requests.EntrepreneurDetails) (*responses.UserPrivate, error) {
		user, err := c.UsersService.UpdateEntrepreneurDetails(reqCtx, userID, req)
		if err != nil {
			return nil, err
		}

		return responses.FromUserPrivate(user), nil
	})
}

// UpdateLocation godoc
// @Summary      Updates user's location
// @Description  Updates existing location if already created. Other way creates new location object.
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        Authorization header    string            true  "Bearer token"
// @Success      200       {object}  responses.SuccessResponse{data=string}  "OK - location updated successfully"
// @Failure      401       {object}  responses.ErrorResponse "Unauthorized - Invalid or missing token"
// @Failure      422       {object}  responses.ErrorResponse "Unprocessable Entity - Update failed"
// @Router       /api/v1/users/details/location [put]
func (c *UsersController) UpdateLocation(ctx *gin.Context) {
	handleAuthenticatedRequest(ctx, c, func(reqCtx context.Context, userID int, req *requests.Location) (*responses.UserPrivate, error) {
		user, err := c.UsersService.UpdateLocation(reqCtx, userID, req)
		if err != nil {
			return nil, err
		}

		return responses.FromUserPrivate(user), nil
	})
}

// Show godoc
// @Summary      Show given user public data
// @Description  Shows public user's data. Returns UserPublic object if user found.
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        Authorization header    string            true  "Bearer token"
// @Success      200       {object}  responses.SuccessResponse{data=responses.UserPublic}  "OK - User fetched successfully"
// @Failure      401       {object}  responses.ErrorResponse "Unauthorized - Invalid or missing token"
// @Failure      422       {object}  responses.ErrorResponse "Unprocessable Entity - fetch failed"
// @Router       /api/v1/users [get]
func (c *UsersController) Show(ctx *gin.Context) {
	handleAuthenticatedRequest(ctx, c, func(reqCtx context.Context, userID int, req *requests.UserPublic) (*responses.UserPublic, error) {
		user, err := c.UsersService.GetUserPublic(reqCtx, req.ID)
		if err != nil {
			return nil, err
		}

		return responses.FromUserPublic(user), nil
	})
}

// Me godoc
// @Summary      Show given user public data
// @Description  Shows user's private data. Returns UserPrivate object if user is authenticated.
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        Authorization header    string            true  "Bearer token"
// @Success      200       {object}  responses.SuccessResponse{data=responses.UserPublic}  "OK - User fetched successfully"
// @Failure      401       {object}  responses.ErrorResponse "Unauthorized - Invalid or missing token"
// @Failure      422       {object}  responses.ErrorResponse "Unprocessable Entity - fetch failed"
// @Router       /api/v1/users/me [get]
func (c *UsersController) Me(ctx *gin.Context) {
	handleAuthenticatedRequest(ctx, c, func(reqCtx context.Context, userID int, req *requests.Empty) (*responses.UserPrivate, error) {
		user, err := c.UsersService.GetUserPrivate(reqCtx, userID)
		if err != nil {
			return nil, err
		}

		return responses.FromUserPrivate(user), nil
	})
}

// Index godoc
// @Summary      Show user's public data for users in choosed range
// @Description  Returns array of UserPublic objects.
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        Authorization header    string            true  "Bearer token"
// @Success      200       {object}  responses.SuccessResponse{data=responses.Users}  "OK - Users fetched successfully"
// @Failure      401       {object}  responses.ErrorResponse "Unauthorized - Invalid or missing token"
// @Failure      422       {object}  responses.ErrorResponse "Unprocessable Entity - fetch failed"
// @Router       /api/v1/users/index [get]
func (c *UsersController) Index(ctx *gin.Context) {
	handleAuthenticatedRequest(ctx, c, func(reqCtx context.Context, userID int, req *requests.Index) (*responses.Users, error) {
		users, err := c.UsersService.Index(ctx, req.Page, req.PageSize)
		if err != nil {
			return nil, err
		}

		return responses.FromUsers(users), nil
	})
}

// AddRole godoc
// @Summary      Adds role for choosen user
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        Authorization header    string            true  "Bearer token"
// @Success      200       {object}  responses.SuccessResponse{data=string}  "OK - Successfully added role for user"
// @Failure      401       {object}  responses.ErrorResponse "Unauthorized - Invalid or missing token"
// @Failure      422       {object}  responses.ErrorResponse "Unprocessable Entity - fetch failed"
// @Router       /api/v1/users/role [post]
func (c *UsersController) AddRole(ctx *gin.Context) {
	handleAuthenticatedRequest(ctx, c, func(reqCtx context.Context, userID int, req *requests.Role) (string, error) {
		err := c.UsersService.AddRole(ctx, req.UserID, req.RoleID)
		if err != nil {
			return "Failed to add role for user", err
		}

		return "Successfully added role for user", nil
	})
}

// RevokeRole godoc
// @Summary      Revokes role from chosen user
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        Authorization header    string            true  "Bearer token"
// @Success      200       {object}  responses.SuccessResponse{data=string}  "OK - Successfully revoked role from user
// @Failure      401       {object}  responses.ErrorResponse "Unauthorized - Invalid or missing token"
// @Failure      422       {object}  responses.ErrorResponse "Unprocessable Entity - fetch failed"
// @Router       /api/v1/users/role [patch]
func (c *UsersController) RevokeRole(ctx *gin.Context) {
	handleAuthenticatedRequest(ctx, c, func(reqCtx context.Context, userID int, req *requests.Role) (string, error) {
		err := c.UsersService.RevokeRole(ctx, req.UserID, req.RoleID)
		if err != nil {
			return "Failed to revoke role from user", err
		}

		return "Successfully revoked role from user", nil
	})
}
