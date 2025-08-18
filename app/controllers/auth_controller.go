package controllers

import (
	"context"
	"net/http"
	requests "users-service/app/models/requests"
	"users-service/app/models/responses"
	"users-service/app/services"
	"users-service/app/utils"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
)

type AuthController struct {
	AuthService  services.AuthServiceInterface
	UsersService services.UsersServiceInterface
	Upgrader     websocket.Upgrader
	Logger       zerolog.Logger
}

type AuthControllerInterface interface {
	InitWebSocketHelper(ctx *gin.Context) (*requests.WebSocketHelper, error)

	Register(ctx *gin.Context)
	RegisterExternally(ctx *gin.Context)
	Login(ctx *gin.Context)
	LoginExternally(ctx *gin.Context)
	Logout(ctx *gin.Context)

	ActivateAccount(ctx *gin.Context)
	VerifyAccount(ctx *gin.Context)

	ResendActivationCode(ctx *gin.Context)
	ResendResetCode(ctx *gin.Context)
	ResendVerificationCode(ctx *gin.Context)
	ResendSecondFactorCode(ctx *gin.Context)

	RequestPasswordReset(ctx *gin.Context)
	CancelPasswordReset(ctx *gin.Context)
	ConfirmPasswordReset(ctx *gin.Context)
}

var _ AuthControllerInterface = (*AuthController)(nil)
var _ StandardController = (*AuthController)(nil)

func NewAuthController(
	authService services.AuthServiceInterface,
	usersService services.UsersServiceInterface,
	logger zerolog.Logger) *AuthController {

	return &AuthController{
		AuthService:  authService,
		UsersService: usersService,
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
		Logger: logger,
	}

}

func (c *AuthController) Log() *zerolog.Logger {
	return &c.Logger
}

func (c *AuthController) RespondFailure(ctx *gin.Context, statusCode int, errMsg string) {
	requestID := utils.GetRequestID(ctx)

	ctx.JSON(statusCode, responses.ErrorResponse{
		Status:    "failure",
		Error:     errMsg,
		RequestID: requestID,
	})
}

func (c *AuthController) RespondSuccess(ctx *gin.Context, statusCode int, data interface{}) {
	requestID := utils.GetRequestID(ctx)

	ctx.JSON(statusCode, responses.SuccessResponse{
		Status:    "success",
		Data:      data,
		RequestID: requestID,
	})
}

func (c *AuthController) InitWebSocketHelper(ctx *gin.Context) (*requests.WebSocketHelper, error) {
	c.Upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	conn, err := c.Upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		c.Logger.Error().Err(err).Msg("Failed to upgrade connection to WebSocket")
		return nil, err
	}

	c.Logger.Info().Str("remote_addr", conn.RemoteAddr().String()).Msg("New WebSocket connection established")
	return &requests.WebSocketHelper{Conn: conn, Logger: c.Logger}, nil
}

// Register godoc
// @Summary      Register a new user
// @Description  Initiates a WebSocket connection to handle user registration with real-time progress updates.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      requests.Register  true  "User registration data"
// @Success      101   {string}  string "Switching Protocols"
// @Failure      400   {object}  responses.ErrorResponse "Bad Request - Invalid input data"
// @Router       /api/v1/auth/register [post]
func (c *AuthController) Register(ctx *gin.Context) {
	handleStandardRequest(ctx, c,
		func(reqCtx context.Context, req *requests.Register) (*responses.UserPrivate, error) {
			user, err := c.AuthService.Register(reqCtx, req)
			if err != nil {
				return nil, err
			}

			return responses.FromUserPrivate(user), nil
		},
	)
}

// ActivateAccount godoc
// @Summary      Activate user account
// @Description  Activates a user's account using a provided activation code.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      requests.Code      true  "Activation code"
// @Success      200   {object}  responses.SuccessResponse{data=responses.User}  "OK - Account activated successfully"
// @Failure      400   {object}  responses.ErrorResponse "Bad Request - Invalid input data"
// @Failure      422   {object}  responses.ErrorResponse "Unprocessable Entity - Invalid or expired code"
// @Router       /api/v1/auth/activation/account [patch]
func (c *AuthController) ActivateAccount(ctx *gin.Context) {
	handleStandardRequest(ctx, c, func(reqCtx context.Context, req *requests.Code) (*responses.UserPrivate, error) {
		user, err := c.AuthService.ActivateAccount(reqCtx, req)
		if err != nil {
			return nil, err
		}

		return responses.FromUserPrivate(user), nil
	})
}

// ResendActivationCode godoc
// @Summary      Resend activation code
// @Description  Requests a new activation code to be sent to the user's email address.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      requests.Mail      true  "User's email address"
// @Success      200   {object}  responses.SuccessResponse{data=string}  "OK - Successfully requested a new activation code"
// @Failure      400   {object}  responses.ErrorResponse "Bad Request - Invalid input data"
// @Failure      422   {object}  responses.ErrorResponse "Unprocessable Entity - Service failed to process the request"
// @Router       /api/v1/auth/activation/resend/code [post]
func (c *AuthController) ResendActivationCode(ctx *gin.Context) {
	handleStandardRequest(ctx, c, func(reqCtx context.Context, req *requests.Mail) (string, error) {
		if err := c.AuthService.ResendActivationCode(reqCtx, req); err != nil {
			return "", err
		}

		return "requested activation code resend", nil
	})
}

// VerifyAccount godoc
// @Summary      Verify user account (e.g., phone)
// @Description  Verifies a user's account with a provided code, returning the user and a JWT.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      requests.Code      true  "Verification code"
// @Success      200   {object}  responses.SuccessResponse{data=responses.User}  "OK - Account verified successfully"
// @Failure      400   {object}  responses.ErrorResponse "Bad Request - Invalid input data"
// @Failure      422   {object}  responses.ErrorResponse "Unprocessable Entity - Invalid or expired code"
// @Router       /api/v1/auth/verification/account [patch]
func (c *AuthController) VerifyAccount(ctx *gin.Context) {
	handleStandardRequest(ctx, c, func(reqCtx context.Context, req *requests.Code) (*responses.UserPrivate, error) {
		user, token, err := c.AuthService.VerifyAccount(reqCtx, req)
		if err != nil {
			return nil, err
		}

		return responses.FromUserPrivate(user, token), nil
	})
}

// ResendVerificationCode godoc
// @Summary      Resend verification code
// @Description  Requests a new verification code to be sent to the user.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      requests.Mail      true  "User's email address"
// @Success      200   {object}  responses.SuccessResponse{data=string}  "OK - Successfully requested a new verification code"
// @Failure      400   {object}  responses.ErrorResponse "Bad Request - Invalid input data"
// @Failure      422   {object}  responses.ErrorResponse "Unprocessable Entity - Service failed to process the request"
// @Router       /api/v1/auth/verification/resend/code [post]
func (c *AuthController) ResendVerificationCode(ctx *gin.Context) {
	handleStandardRequest(ctx, c, func(reqCtx context.Context, req *requests.Mail) (string, error) {
		if err := c.AuthService.ResendVerificationCode(reqCtx, req); err != nil {
			return "", err
		}

		return "requested verification code resend", nil
	})
}

// Login godoc
// @Summary      Login a user
// @Description  Initiates a WebSocket connection to handle user login with real-time progress updates.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      requests.Login  true  "User login credentials"
// @Success      101   {string}  string "Switching Protocols"
// @Failure      400   {object}  responses.ErrorResponse "Bad Request - Invalid input data"
// @Router       /api/v1/auth/login [post]
func (c *AuthController) Login(ctx *gin.Context) {
	handleStandardRequest(ctx, c,
		func(reqCtx context.Context, req *requests.Login) (*responses.UserPrivate, error) {
			user, token, err := c.AuthService.Login(reqCtx, req)
			if err != nil {
				return nil, err
			}

			return responses.FromUserPrivate(user, token), nil
		},
	)
}

// ResendSecondFactorCode godoc
// @Summary      Resend second factor code
// @Description  Requests a new second factor code to be sent to the user.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      requests.Mail      true  "User's email address"
// @Success      200   {object}  responses.SuccessResponse{data=string}  "OK - Successfully requested a new 2FA code"
// @Failure      400   {object}  responses.ErrorResponse "Bad Request - Invalid input data"
// @Failure      422   {object}  responses.ErrorResponse "Unprocessable Entity - Service failed to process the request"
// @Router       /api/v1/auth/second-factor/resend/code [post]
func (c *AuthController) ResendSecondFactorCode(ctx *gin.Context) {
	handleStandardRequest(ctx, c, func(reqCtx context.Context, req *requests.Mail) (string, error) {
		if err := c.AuthService.ResendSecondFactorCode(reqCtx, req); err != nil {
			return "", err
		}

		return "requested verification code resend", nil
	})
}

// VerifySecondFactorCode godoc
// @Summary      Resend second factor code
// @Description  Requests a new second factor code to be sent to the user.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      requests.Code      true  "User's email address"
// @Success      200   {object}  responses.SuccessResponse{data=responses.User}  "OK - Successfully requested a new 2FA code"
// @Failure      400   {object}  responses.ErrorResponse "Bad Request - Invalid input data"
// @Failure      422   {object}  responses.ErrorResponse "Unprocessable Entity - Service failed to process the request"
// @Router       /api/v1/auth/second-factor/verification/code [post]
func (c *AuthController) VerifySecondFactorCode(ctx *gin.Context) {
	handleStandardRequest(ctx, c, func(reqCtx context.Context, req *requests.Code) (*responses.UserPrivate, error) {
		user, token, err := c.AuthService.VerifySecondFactor(reqCtx, req)
		if err != nil {
			return nil, err
		}

		return responses.FromUserPrivate(user, token), nil
	})
}

// Logout godoc
// @Summary      Logout a user
// @Description  Logs out the current user by invalidating their token (not implemented).
// @Tags         auth
// @Produce      json
// @Success      501  {object}  responses.ErrorResponse "Not Implemented"
// @Router       /api/v1/auth/logout [delete]
func (c *AuthController) Logout(ctx *gin.Context) {
	c.RespondFailure(ctx, http.StatusNotImplemented, "Not implemented")
}

// RequestPasswordReset godoc
// @Summary      Request password reset
// @Description  Initiates the password reset process by sending a code to the user's email.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      requests.Mail      true  "User's email address"
// @Success      200   {object}  responses.SuccessResponse{data=string}  "OK - Password reset email sent"
// @Failure      400   {object}  responses.ErrorResponse "Bad Request - Invalid input data"
// @Failure      422   {object}  responses.ErrorResponse "Unprocessable Entity - Service failed to process the request"
// @Router       /api/v1/auth/password/reset/request [post]
func (c *AuthController) RequestPasswordReset(ctx *gin.Context) {
	handleStandardRequest(ctx, c, func(reqCtx context.Context, req *requests.Mail) (string, error) {
		if err := c.AuthService.RequestPasswordReset(reqCtx, req); err != nil {
			return "", err
		}

		return "requested reset code resend", nil
	})
}

// CancelPasswordReset godoc
// @Summary      Cancel password reset
// @Description  Cancels an active password reset request (not implemented).
// @Tags         auth
// @Produce      json
// @Success      501  {object}  responses.ErrorResponse "Not Implemented"
// @Router       /api/v1/auth/password/reset/request/cancel [delete]
func (c *AuthController) CancelPasswordReset(ctx *gin.Context) {
	handleStandardRequest(ctx, c, func(reqCtx context.Context, req *requests.Mail) (string, error) {
		if err := c.AuthService.CancelPasswordReset(reqCtx, req); err != nil {
			return "", err
		}

		return "cancelled password resetting procedure", nil
	})
}

// ConfirmPasswordReset godoc
// @Summary      Confirm password reset
// @Description  Sets a new password for the user using a valid reset code.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      requests.ConfirmPasswordReset  true  "Password reset code and new password"
// @Success      200   {object}  responses.SuccessResponse{data=string}  "OK - Password has been reset successfully"
// @Failure      400   {object}  responses.ErrorResponse "Bad Request - Invalid input data"
// @Failure      422   {object}  responses.ErrorResponse "Unprocessable Entity - Invalid code or other error"
// @Router       /api/v1/auth/password/reset/request/confirm [patch]
func (c *AuthController) ConfirmPasswordReset(ctx *gin.Context) {
	handleStandardRequest(ctx, c, func(reqCtx context.Context, req *requests.ConfirmPasswordReset) (string, error) {
		if err := c.AuthService.ConfirmPasswordReset(reqCtx, req); err != nil {
			return "", err
		}

		return "confirmed password reset", nil
	})
}

// ResendResetCode godoc
// @Summary      Resend Password Reset Code
// @Description  Requests a new password reset code to be sent to the user's email address.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      requests.Mail      true  "User's email address"
// @Success      200   {object}  responses.SuccessResponse{data=string}  "OK - Successfully requested a new reset code"
// @Failure      400   {object}  responses.ErrorResponse "Bad Request - Invalid input data"
// @Failure      422   {object}  responses.ErrorResponse "Unprocessable Entity - Service failed to process the request (e.g., user not found)"
// @Router       /api/v1/auth/password/reset/request/resend [post]
func (c *AuthController) ResendResetCode(ctx *gin.Context) {
	handleStandardRequest(ctx, c, func(reqCtx context.Context, req *requests.Mail) (string, error) {
		if err := c.AuthService.ResendResetCode(reqCtx, req); err != nil {
			return "", err
		}

		return "requested reset code resend", nil
	})
}

// RegisterExternally godoc
// @Summary      Register a user via an external provider
// @Description  Handles user registration through an external provider (not implemented).
// @Tags         auth
// @Produce      json
// @Success      501  {object}  responses.ErrorResponse "Not Implemented"
// @Router       /api/v1/auth/register/externally [post]
func (c *AuthController) RegisterExternally(ctx *gin.Context) {
	c.RespondFailure(ctx, http.StatusNotImplemented, "Not implemented")
}

// LoginExternally godoc
// @Summary      Login a user via an external provider
// @Description  Handles user login through an external provider (not implemented).
// @Tags         auth
// @Produce      json
// @Success      501  {object}  responses.ErrorResponse "Not Implemented"
// @Router       /api/v1/auth/login/externally [post]
func (c *AuthController) LoginExternally(ctx *gin.Context) {
	c.RespondFailure(ctx, http.StatusNotImplemented, "Not implemented")
}
