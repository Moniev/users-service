//go:build unit

package controllers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"users-service/app/models/ent"
	"users-service/app/models/requests"
	"users-service/tests/mocks"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/mock"
)

func setupAuthController(t *testing.T) (
	*gin.Context,
	*httptest.ResponseRecorder,
	*gin.Engine,
	*mocks.MockAuthService,
	*mocks.MockUsersService,
	*AuthController) {

	gin.SetMode(gin.TestMode)

	mockAuthService := new(mocks.MockAuthService)
	mockUsersService := new(mocks.MockUsersService)
	logger := zerolog.Nop()

	authController := NewAuthController(mockAuthService, mockUsersService, logger)

	router := gin.New()
	router.POST("/api/v1/auth/register", authController.Register)
	router.POST("/api/v1/auth/login", authController.Login)
	router.POST("/api/v1/auth/register/externally", authController.RegisterExternally)
	router.POST("/api/v1/auth/login/externally", authController.LoginExternally)
	router.DELETE("/api/v1/auth/logout", authController.Logout)
	router.PATCH("/api/v1/auth/activation/account", authController.ActivateAccount)
	router.PATCH("/api/v1/auth/verification/account", authController.VerifyAccount)
	router.POST("/api/v1/auth/activation/resend/code", authController.ResendActivationCode)
	router.POST("/api/v1/auth/verification/resend/code", authController.ResendVerificationCode)
	router.POST("/api/v1/auth/second-factor/resend/code", authController.ResendSecondFactorCode)
	router.POST("/api/v1/auth/second-factor/verification/code", authController.VerifySecondFactorCode)
	router.POST("/api/v1/auth/password/reset/request", authController.RequestPasswordReset)
	router.DELETE("/api/v1/auth/password/reset/request/cancel", authController.CancelPasswordReset)
	router.PATCH("/api/v1/auth/password/reset/request/confirm", authController.ConfirmPasswordReset)
	router.POST("/api/v1/auth/password/reset/request/resend", authController.ResendResetCode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	return c, w, router, mockAuthService, mockUsersService, authController
}

func TestAuthController_Register(t *testing.T) {
	testCases := []struct {
		name               string
		requestBody        requests.Register
		setupMocks         func(m *mocks.MockAuthService)
		expectedStatusCode int
		expectedResponse   string
	}{
		{
			name: "Success",
			requestBody: requests.Register{
				Name:     "Stec",
				Mail:     "random@mail.com",
				Password: "StrongPassword!3@3",
				Device: requests.Device{
					DeviceToken:    "token123",
					IPAddress:      "192.168.1.1",
					UserAgent:      "Mozilla/5.0",
					OSName:         "Windows",
					OSVersion:      "10",
					BrowserName:    "Chrome",
					BrowserVersion: "120.0",
				},
			},
			setupMocks: func(m *mocks.MockAuthService) {
				m.On("Register", mock.Anything, &requests.Register{
					Name:     "Stec",
					Mail:     "random@mail.com",
					Password: "StrongPassword!3@3",
					Device: requests.Device{
						DeviceToken:    "token123",
						IPAddress:      "192.168.1.1",
						UserAgent:      "Mozilla/5.0",
						OSName:         "Windows",
						OSVersion:      "10",
						BrowserName:    "Chrome",
						BrowserVersion: "120.0",
					},
				}).Return(&ent.User{Mail: "random@mail.com"}, nil)
			},
			expectedStatusCode: http.StatusOK,

			expectedResponse: `{"status":"success", "data":{"active":false,"blacklisted":false,"created_at":"0001-01-01T00:00:00Z","edges":{"user_details":null,"user_devices":null,"user_roles":null,"user_settings":null},"id":0,"mail":"random@mail.com","organization_ids":null,"phone":"","removed":false,"subscription_ids":null,"team_ids":null,"updated_at":"0001-01-01T00:00:00Z","verified":false}}`,
		},
		{
			name: "Invalid Credentials 1.1",
			requestBody: requests.Register{
				Name:     "Stec",
				Mail:     "random@mail.com",
				Password: "easypassword",
				Device: requests.Device{
					DeviceToken:    "token123",
					IPAddress:      "192.168.1.1",
					UserAgent:      "Mozilla/5.0",
					OSName:         "Windows",
					OSVersion:      "10",
					BrowserName:    "Chrome",
					BrowserVersion: "120.0",
				},
			},
			setupMocks:         func(m *mocks.MockAuthService) {},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   `{"status":"failure","error":"invalid request data: password does not meet complexity requirements"}`,
		},
		{
			name: "Invalid Credentials 1.2",
			requestBody: requests.Register{
				Name:     "Stec",
				Mail:     "invalid-mail.com",
				Password: "easypassword",
				Device: requests.Device{
					DeviceToken:    "token123",
					IPAddress:      "192.168.1.1",
					UserAgent:      "Mozilla/5.0",
					OSName:         "Windows",
					OSVersion:      "10",
					BrowserName:    "Chrome",
					BrowserVersion: "120.0",
				},
			},
			setupMocks:         func(m *mocks.MockAuthService) {},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   `{"status":"failure","error":"invalid body request"}`,
		},
		{
			name: "Invalid Input Data",
			requestBody: requests.Register{
				Name:     "Alfa",
				Mail:     "random@mail.com",
				Password: "StrongPassword!3@3",
				Device: requests.Device{
					DeviceToken:    "token123",
					IPAddress:      "192.168.1.1",
					UserAgent:      "Mozilla/5.0",
					OSName:         "Windows",
					OSVersion:      "10",
					BrowserName:    "Chrome",
					BrowserVersion: "120.0",
				},
			},
			setupMocks: func(m *mocks.MockAuthService) {
				m.On("Register", mock.Anything, &requests.Register{
					Name:     "Alfa",
					Mail:     "random@mail.com",
					Password: "StrongPassword!3@3",
					Device: requests.Device{
						DeviceToken:    "token123",
						IPAddress:      "192.168.1.1",
						UserAgent:      "Mozilla/5.0",
						OSName:         "Windows",
						OSVersion:      "10",
						BrowserName:    "Chrome",
						BrowserVersion: "120.0",
					},
				}).Return(nil, errors.New("failed to register"))
			},
			expectedStatusCode: http.StatusUnprocessableEntity,
			expectedResponse:   `{"error":"failed to register","status":"failure"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c, w, _, mockAuthService, _, authController := setupAuthController(t)

			body, _ := json.Marshal(tc.requestBody)
			c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBuffer(body))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Request.Header.Set("Authorization", "Bearer token")

			tc.setupMocks(mockAuthService)
			authController.Register(c)

			if w.Code != tc.expectedStatusCode {
				t.Errorf("Expected status code %d, got %d. Body: %s", tc.expectedStatusCode, w.Code, w.Body.String())
			}
		})
	}
}

func TestAuthController_RegisterExternally(t *testing.T) {
	testCases := []struct {
		name               string
		requestBody        requests.Empty
		setupMocks         func(m *mocks.MockAuthService)
		expectedStatusCode int
		expectedResponse   string
	}{}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c, w, _, mockAuthService, _, authController := setupAuthController(t)

			body, _ := json.Marshal(tc.requestBody)
			c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/register/externally", bytes.NewBuffer(body))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Request.Header.Set("Authorization", "Bearer token")

			tc.setupMocks(mockAuthService)
			authController.RegisterExternally(c)

			if w.Code != tc.expectedStatusCode {
				t.Errorf("Expected status code %d, got %d. Body: %s", tc.expectedStatusCode, w.Code, w.Body.String())
			}
		})
	}
}

func TestAuthController_Login(t *testing.T) {
	testCases := []struct {
		name               string
		requestBody        requests.Login
		setupMocks         func(m *mocks.MockAuthService)
		expectedStatusCode int
		expectedResponse   string
	}{
		{
			name: "Success",
			requestBody: requests.Login{
				Mail:     "random@mail.com",
				Password: "StrongPassword!3@3",
				Device: requests.Device{
					DeviceToken:    "token123",
					IPAddress:      "192.168.1.1",
					UserAgent:      "Mozilla/5.0",
					OSName:         "Windows",
					OSVersion:      "10",
					BrowserName:    "Chrome",
					BrowserVersion: "120.0",
				},
			},
			setupMocks: func(m *mocks.MockAuthService) {
				m.On("Login", mock.Anything, &requests.Login{
					Mail:     "random@mail.com",
					Password: "StrongPassword!3@3",
					Device: requests.Device{
						DeviceToken:    "token123",
						IPAddress:      "192.168.1.1",
						UserAgent:      "Mozilla/5.0",
						OSName:         "Windows",
						OSVersion:      "10",
						BrowserName:    "Chrome",
						BrowserVersion: "120.0",
					},
				}).Return(&ent.User{Mail: "random@mail.com"}, "token", nil)
			},
			expectedStatusCode: http.StatusOK,

			expectedResponse: `{"status":"success", "data":{"user":{"active":false,"blacklisted":false,"created_at":"0001-01-01T00:00:00Z","edges":{"user_details":null,"user_devices":null,"user_roles":null,"user_settings":null},"id":0,"mail":"random@mail.com","organization_ids":null,"phone":"","removed":false,"subscription_ids":null,"team_ids":null,"updated_at":"0001-01-01T00:00:00Z","verified":false},"token":"token"}}`,
		},
		{
			name: "Invalid Credentials 1.1",
			requestBody: requests.Login{
				Mail:     "random@mail.com",
				Password: "easypassword",
				Device: requests.Device{
					DeviceToken:    "token123",
					IPAddress:      "192.168.1.1",
					UserAgent:      "Mozilla/5.0",
					OSName:         "Windows",
					OSVersion:      "10",
					BrowserName:    "Chrome",
					BrowserVersion: "120.0",
				},
			},
			setupMocks:         func(m *mocks.MockAuthService) {},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   `{"status":"failure","error":"invalid request data: password does not meet complexity requirements"}`,
		},
		{
			name: "Invalid Credentials 1.2",
			requestBody: requests.Login{
				Mail:     "invalid-mail.com",
				Password: "easypassword",
				Device: requests.Device{
					DeviceToken:    "token123",
					IPAddress:      "192.168.1.1",
					UserAgent:      "Mozilla/5.0",
					OSName:         "Windows",
					OSVersion:      "10",
					BrowserName:    "Chrome",
					BrowserVersion: "120.0",
				},
			},
			setupMocks:         func(m *mocks.MockAuthService) {},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   `{"status":"failure","error":"invalid request data: invalid email format provided"}`,
		},
		{
			name: "Invalid Input Data",
			requestBody: requests.Login{
				Mail:     "random@mail.com",
				Password: "StrongPassword!3@3",
				Device: requests.Device{
					DeviceToken:    "token123",
					IPAddress:      "192.168.1.1",
					UserAgent:      "Mozilla/5.0",
					OSName:         "Windows",
					OSVersion:      "10",
					BrowserName:    "Chrome",
					BrowserVersion: "120.0",
				},
			},
			setupMocks: func(m *mocks.MockAuthService) {
				m.On("Login", mock.Anything, &requests.Login{
					Mail:     "random@mail.com",
					Password: "StrongPassword!3@3",
					Device: requests.Device{
						DeviceToken:    "token123",
						IPAddress:      "192.168.1.1",
						UserAgent:      "Mozilla/5.0",
						OSName:         "Windows",
						OSVersion:      "10",
						BrowserName:    "Chrome",
						BrowserVersion: "120.0",
					},
				}).Return(nil, "", errors.New("failed to login"))
			},
			expectedStatusCode: http.StatusUnprocessableEntity,
			expectedResponse:   `{"error":"failed to login","status":"failure"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c, w, _, mockAuthService, _, authController := setupAuthController(t)

			body, _ := json.Marshal(tc.requestBody)
			c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(body))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Request.Header.Set("Authorization", "Bearer token")

			tc.setupMocks(mockAuthService)
			authController.Login(c)

			if w.Code != tc.expectedStatusCode {
				t.Errorf("Expected status code %d, got %d. Body: %s", tc.expectedStatusCode, w.Code, w.Body.String())
			}
		})
	}
}

func TestAuthController_LoginExternally(t *testing.T) {
	testCases := []struct {
		name               string
		requestBody        requests.Empty
		setupMocks         func(m *mocks.MockAuthService)
		expectedStatusCode int
		expectedResponse   string
	}{}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c, w, _, mockAuthService, _, authController := setupAuthController(t)

			body, _ := json.Marshal(tc.requestBody)
			c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/login/externally", bytes.NewBuffer(body))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Request.Header.Set("Authorization", "Bearer token")

			tc.setupMocks(mockAuthService)
			authController.LoginExternally(c)

			if w.Code != tc.expectedStatusCode {
				t.Errorf("Expected status code %d, got %d. Body: %s", tc.expectedStatusCode, w.Code, w.Body.String())
			}
		})
	}
}

func TestAuthController_Logout(t *testing.T) {
	testCases := []struct {
		name               string
		requestBody        requests.Empty
		setupMocks         func(m *mocks.MockAuthService)
		expectedStatusCode int
		expectedResponse   string
	}{}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c, w, _, mockAuthService, _, authController := setupAuthController(t)

			body, _ := json.Marshal(tc.requestBody)
			c.Request = httptest.NewRequest(http.MethodDelete, "/api/v1/auth/logout", bytes.NewBuffer(body))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Request.Header.Set("Authorization", "Bearer token")

			tc.setupMocks(mockAuthService)
			authController.Logout(c)

			if w.Code != tc.expectedStatusCode {
				t.Errorf("Expected status code %d, got %d. Body: %s", tc.expectedStatusCode, w.Code, w.Body.String())
			}
		})
	}
}

func TestAuthController_ActivateAccount(t *testing.T) {
	testCases := []struct {
		name               string
		requestBody        requests.Code
		setupMocks         func(m *mocks.MockAuthService)
		expectedStatusCode int
		expectedResponse   string
	}{
		{
			name: "Success",
			requestBody: requests.Code{
				Code: "123123",
				Device: requests.Device{
					DeviceToken:    "token123",
					IPAddress:      "192.168.1.1",
					UserAgent:      "Mozilla/5.0",
					OSName:         "Windows",
					OSVersion:      "10",
					BrowserName:    "Chrome",
					BrowserVersion: "120.0",
				},
			},
			setupMocks: func(m *mocks.MockAuthService) {
				m.On("ActivateAccount", mock.Anything, &requests.Code{
					Code: "123123",
					Device: requests.Device{
						DeviceToken:    "token123",
						IPAddress:      "192.168.1.1",
						UserAgent:      "Mozilla/5.0",
						OSName:         "Windows",
						OSVersion:      "10",
						BrowserName:    "Chrome",
						BrowserVersion: "120.0",
					},
				}).Return(&ent.User{Mail: "random@mail.com"}, nil)
			},
			expectedStatusCode: http.StatusOK,

			expectedResponse: `{"status":"success", "data":{"user":{"active":false,"blacklisted":false,"created_at":"0001-01-01T00:00:00Z","edges":{"user_details":null,"user_devices":null,"user_roles":null,"user_settings":null},"id":0,"mail":"random@mail.com","organization_ids":null,"phone":"","removed":false,"subscription_ids":null,"team_ids":null,"updated_at":"0001-01-01T00:00:00Z","verified":false}}}`,
		},
		{
			name: "Invalid Input",
			requestBody: requests.Code{
				Code: "123123",
				Device: requests.Device{
					DeviceToken:    "token123",
					IPAddress:      "192.168.1.1",
					UserAgent:      "Mozilla/5.0",
					OSName:         "Windows",
					OSVersion:      "10",
					BrowserName:    "Chrome",
					BrowserVersion: "120.0",
				},
			},
			setupMocks: func(m *mocks.MockAuthService) {
				m.On("ActivateAccount", mock.Anything, &requests.Code{
					Code: "123123",
					Device: requests.Device{
						DeviceToken:    "token123",
						IPAddress:      "192.168.1.1",
						UserAgent:      "Mozilla/5.0",
						OSName:         "Windows",
						OSVersion:      "10",
						BrowserName:    "Chrome",
						BrowserVersion: "120.0",
					},
				}).Return(nil, errors.New("failed to update"))
			},
			expectedStatusCode: http.StatusUnprocessableEntity,
			expectedResponse:   `{"error":"failed to update","status":"failure"}`,
		},
		{
			name: "Invalid Credentials",
			requestBody: requests.Code{
				Code: "123-123",
				Device: requests.Device{
					DeviceToken:    "token123",
					IPAddress:      "192.168.1.1",
					UserAgent:      "Mozilla/5.0",
					OSName:         "Windows",
					OSVersion:      "10",
					BrowserName:    "Chrome",
					BrowserVersion: "120.0",
				},
			},
			setupMocks:         func(m *mocks.MockAuthService) {},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   `{"status":"failure","error":"invalid request data: invalid code format provided"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c, w, _, mockAuthService, _, authController := setupAuthController(t)

			body, _ := json.Marshal(tc.requestBody)
			c.Request = httptest.NewRequest(http.MethodPatch, "/api/v1/auth/activation/account", bytes.NewBuffer(body))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Request.Header.Set("Authorization", "Bearer token")

			tc.setupMocks(mockAuthService)
			authController.ActivateAccount(c)

			if w.Code != tc.expectedStatusCode {
				t.Errorf("Expected status code %d, got %d. Body: %s", tc.expectedStatusCode, w.Code, w.Body.String())
			}
		})
	}
}

func TestAuthController_VerifySecondFactorCode(t *testing.T) {
	testCases := []struct {
		name               string
		requestBody        requests.Code
		setupMocks         func(m *mocks.MockAuthService)
		expectedStatusCode int
		expectedResponse   string
	}{
		{
			name: "Success",
			requestBody: requests.Code{
				Code: "123123",
				Device: requests.Device{
					DeviceToken:    "token123",
					IPAddress:      "192.168.1.1",
					UserAgent:      "Mozilla/5.0",
					OSName:         "Windows",
					OSVersion:      "10",
					BrowserName:    "Chrome",
					BrowserVersion: "120.0",
				},
			},
			setupMocks: func(m *mocks.MockAuthService) {
				m.On("VerifySecondFactor", mock.Anything, &requests.Code{
					Code: "123123",
					Device: requests.Device{
						DeviceToken:    "token123",
						IPAddress:      "192.168.1.1",
						UserAgent:      "Mozilla/5.0",
						OSName:         "Windows",
						OSVersion:      "10",
						BrowserName:    "Chrome",
						BrowserVersion: "120.0",
					},
				}).Return(&ent.User{Mail: "random@mail.com"}, "token", nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse:   `{"status":"success", "data":{"user":{"active":false,"blacklisted":false,"created_at":"0001-01-01T00:00:00Z","edges":{"user_details":null,"user_devices":null,"user_roles":null,"user_settings":null},"id":0,"mail":"random@mail.com","organization_ids":null,"phone":"","removed":false,"subscription_ids":null,"team_ids":null,"updated_at":"0001-01-01T00:00:00Z","verified":false}, "token": "token"}}`,
		},
		{
			name: "Invalid Input",
			requestBody: requests.Code{
				Code: "123123",
				Device: requests.Device{
					DeviceToken:    "token123",
					IPAddress:      "192.168.1.1",
					UserAgent:      "Mozilla/5.0",
					OSName:         "Windows",
					OSVersion:      "10",
					BrowserName:    "Chrome",
					BrowserVersion: "120.0",
				},
			},
			setupMocks: func(m *mocks.MockAuthService) {
				m.On("VerifySecondFactor", mock.Anything, &requests.Code{
					Code: "123123",
					Device: requests.Device{
						DeviceToken:    "token123",
						IPAddress:      "192.168.1.1",
						UserAgent:      "Mozilla/5.0",
						OSName:         "Windows",
						OSVersion:      "10",
						BrowserName:    "Chrome",
						BrowserVersion: "120.0",
					},
				}).Return(nil, "", errors.New("failed to update"))
			},
			expectedStatusCode: http.StatusUnprocessableEntity,
			expectedResponse:   `{"error":"failed to update","status":"failure"}`,
		},
		{
			name: "Invalid Credentials",
			requestBody: requests.Code{
				Code: "123-123",
				Device: requests.Device{
					DeviceToken:    "token123",
					IPAddress:      "192.168.1.1",
					UserAgent:      "Mozilla/5.0",
					OSName:         "Windows",
					OSVersion:      "10",
					BrowserName:    "Chrome",
					BrowserVersion: "120.0",
				},
			},
			setupMocks:         func(m *mocks.MockAuthService) {},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   `{"status":"failure","error":"invalid request data: invalid code format provided"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c, w, _, mockAuthService, _, authController := setupAuthController(t)

			body, _ := json.Marshal(tc.requestBody)
			c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/second-factor/verification/code", bytes.NewBuffer(body))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Request.Header.Set("Authorization", "Bearer token")

			tc.setupMocks(mockAuthService)
			authController.VerifySecondFactorCode(c)

			if w.Code != tc.expectedStatusCode {
				t.Errorf("Expected status code %d, got %d. Body: %s", tc.expectedStatusCode, w.Code, w.Body.String())
			}
		})
	}
}

func TestAuthController_VerifyAccount(t *testing.T) {
	testCases := []struct {
		name               string
		requestBody        requests.Code
		setupMocks         func(m *mocks.MockAuthService)
		expectedStatusCode int
		expectedResponse   string
	}{
		{
			name: "Success",
			requestBody: requests.Code{
				Code: "123123",
				Device: requests.Device{
					DeviceToken:    "token123",
					IPAddress:      "192.168.1.1",
					UserAgent:      "Mozilla/5.0",
					OSName:         "Windows",
					OSVersion:      "10",
					BrowserName:    "Chrome",
					BrowserVersion: "120.0",
				},
			},
			setupMocks: func(m *mocks.MockAuthService) {
				m.On("VerifyAccount", mock.Anything, &requests.Code{
					Code: "123123",
					Device: requests.Device{
						DeviceToken:    "token123",
						IPAddress:      "192.168.1.1",
						UserAgent:      "Mozilla/5.0",
						OSName:         "Windows",
						OSVersion:      "10",
						BrowserName:    "Chrome",
						BrowserVersion: "120.0",
					},
				}).Return(&ent.User{Mail: "random@mail.com"}, "token", nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse:   `{"status":"success", "data":{"user":{"active":false,"blacklisted":false,"created_at":"0001-01-01T00:00:00Z","edges":{"user_details":null,"user_devices":null,"user_roles":null,"user_settings":null},"id":0,"mail":"random@mail.com","organization_ids":null,"phone":"","removed":false,"subscription_ids":null,"team_ids":null,"updated_at":"0001-01-01T00:00:00Z","verified":false},"token":"token"}}`,
		},
		{
			name: "Invalid Input",
			requestBody: requests.Code{
				Code: "123123",
				Device: requests.Device{
					DeviceToken:    "token123",
					IPAddress:      "192.168.1.1",
					UserAgent:      "Mozilla/5.0",
					OSName:         "Windows",
					OSVersion:      "10",
					BrowserName:    "Chrome",
					BrowserVersion: "120.0",
				},
			},
			setupMocks: func(m *mocks.MockAuthService) {
				m.On("VerifyAccount", mock.Anything, &requests.Code{
					Code: "123123",
					Device: requests.Device{
						DeviceToken:    "token123",
						IPAddress:      "192.168.1.1",
						UserAgent:      "Mozilla/5.0",
						OSName:         "Windows",
						OSVersion:      "10",
						BrowserName:    "Chrome",
						BrowserVersion: "120.0",
					},
				}).Return(nil, "", errors.New("failed to update"))
			},
			expectedStatusCode: http.StatusUnprocessableEntity,
			expectedResponse:   `{"error":"failed to update","status":"failure"}`,
		},
		{
			name: "Invalid Credentials",
			requestBody: requests.Code{
				Code: "123-123",
				Device: requests.Device{
					DeviceToken:    "token123",
					IPAddress:      "192.168.1.1",
					UserAgent:      "Mozilla/5.0",
					OSName:         "Windows",
					OSVersion:      "10",
					BrowserName:    "Chrome",
					BrowserVersion: "120.0",
				},
			},
			setupMocks:         func(m *mocks.MockAuthService) {},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   `{"status":"failure","error":"invalid request data: invalid code format provided"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c, w, _, mockAuthService, _, authController := setupAuthController(t)

			body, _ := json.Marshal(tc.requestBody)
			c.Request = httptest.NewRequest(http.MethodPatch, "/api/v1/auth/verification/account", bytes.NewBuffer(body))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Request.Header.Set("Authorization", "Bearer token")

			tc.setupMocks(mockAuthService)
			authController.VerifyAccount(c)

			if w.Code != tc.expectedStatusCode {
				t.Errorf("Expected status code %d, got %d. Body: %s", tc.expectedStatusCode, w.Code, w.Body.String())
			}
		})
	}
}

func TestAuthController_ResendActivationCode(t *testing.T) {
	testCases := []struct {
		name               string
		requestBody        requests.Mail
		setupMocks         func(m *mocks.MockAuthService)
		expectedStatusCode int
		expectedResponse   string
	}{
		{
			name: "Success",
			requestBody: requests.Mail{
				Mail: "random@mail.com",
				Device: requests.Device{
					DeviceToken:    "token123",
					IPAddress:      "192.168.1.1",
					UserAgent:      "Mozilla/5.0",
					OSName:         "Windows",
					OSVersion:      "10",
					BrowserName:    "Chrome",
					BrowserVersion: "120.0",
				},
			},
			setupMocks: func(m *mocks.MockAuthService) {
				m.On("ResendActivationCode", mock.Anything, &requests.Mail{
					Mail: "random@mail.com",
					Device: requests.Device{
						DeviceToken:    "token123",
						IPAddress:      "192.168.1.1",
						UserAgent:      "Mozilla/5.0",
						OSName:         "Windows",
						OSVersion:      "10",
						BrowserName:    "Chrome",
						BrowserVersion: "120.0",
					},
				}).Return(nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse:   `{"status":"success","data":"requested activation code resend"}`,
		},
		{
			name: "Invalid Input",
			requestBody: requests.Mail{
				Mail: "random@mail.com",
				Device: requests.Device{
					DeviceToken:    "token123",
					IPAddress:      "192.168.1.1",
					UserAgent:      "Mozilla/5.0",
					OSName:         "Windows",
					OSVersion:      "10",
					BrowserName:    "Chrome",
					BrowserVersion: "120.0",
				},
			},
			setupMocks: func(m *mocks.MockAuthService) {
				m.On("ResendActivationCode", mock.Anything, &requests.Mail{
					Mail: "random@mail.com",
					Device: requests.Device{
						DeviceToken:    "token123",
						IPAddress:      "192.168.1.1",
						UserAgent:      "Mozilla/5.0",
						OSName:         "Windows",
						OSVersion:      "10",
						BrowserName:    "Chrome",
						BrowserVersion: "120.0",
					},
				}).Return(errors.New("failed to update"))
			},
			expectedStatusCode: http.StatusUnprocessableEntity,
			expectedResponse:   `{"error":"failed to update","status":"failure"}`,
		},
		{
			name: "Invalid Credentials",
			requestBody: requests.Mail{
				Mail: "random-mail.com",
				Device: requests.Device{
					DeviceToken:    "token123",
					IPAddress:      "192.168.1.1",
					UserAgent:      "Mozilla/5.0",
					OSName:         "Windows",
					OSVersion:      "10",
					BrowserName:    "Chrome",
					BrowserVersion: "120.0",
				},
			},
			setupMocks:         func(m *mocks.MockAuthService) {},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   `{"status":"failure","error":"invalid request data: wrong mail provided"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c, w, _, mockAuthService, _, authController := setupAuthController(t)

			body, _ := json.Marshal(tc.requestBody)
			c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/activation/resend/code", bytes.NewBuffer(body))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Request.Header.Set("Authorization", "Bearer token")

			tc.setupMocks(mockAuthService)
			authController.ResendActivationCode(c)

			if w.Code != tc.expectedStatusCode {
				t.Errorf("Expected status code %d, got %d. Body: %s", tc.expectedStatusCode, w.Code, w.Body.String())
			}
		})
	}
}

func TestAuthController_ResendVerificationCode(t *testing.T) {
	testCases := []struct {
		name               string
		requestBody        requests.Mail
		setupMocks         func(m *mocks.MockAuthService)
		expectedStatusCode int
		expectedResponse   string
	}{
		{
			name: "Success",
			requestBody: requests.Mail{
				Mail: "random@mail.com",
				Device: requests.Device{
					DeviceToken:    "token123",
					IPAddress:      "192.168.1.1",
					UserAgent:      "Mozilla/5.0",
					OSName:         "Windows",
					OSVersion:      "10",
					BrowserName:    "Chrome",
					BrowserVersion: "120.0",
				},
			},
			setupMocks: func(m *mocks.MockAuthService) {
				m.On("ResendVerificationCode", mock.Anything, &requests.Mail{
					Mail: "random@mail.com",
					Device: requests.Device{
						DeviceToken:    "token123",
						IPAddress:      "192.168.1.1",
						UserAgent:      "Mozilla/5.0",
						OSName:         "Windows",
						OSVersion:      "10",
						BrowserName:    "Chrome",
						BrowserVersion: "120.0",
					},
				}).Return(nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse:   `{"status":"success","data":"requested verification code resend"}`,
		},
		{
			name: "Invalid Input",
			requestBody: requests.Mail{
				Mail: "random@mail.com",
				Device: requests.Device{
					DeviceToken:    "token123",
					IPAddress:      "192.168.1.1",
					UserAgent:      "Mozilla/5.0",
					OSName:         "Windows",
					OSVersion:      "10",
					BrowserName:    "Chrome",
					BrowserVersion: "120.0",
				},
			},
			setupMocks: func(m *mocks.MockAuthService) {
				m.On("ResendVerificationCode", mock.Anything, &requests.Mail{
					Mail: "random@mail.com",
					Device: requests.Device{
						DeviceToken:    "token123",
						IPAddress:      "192.168.1.1",
						UserAgent:      "Mozilla/5.0",
						OSName:         "Windows",
						OSVersion:      "10",
						BrowserName:    "Chrome",
						BrowserVersion: "120.0",
					},
				}).Return(errors.New("failed to update"))
			},
			expectedStatusCode: http.StatusUnprocessableEntity,
			expectedResponse:   `{"error":"failed to update","status":"failure"}`,
		},
		{
			name: "Invalid Credentials",
			requestBody: requests.Mail{
				Mail: "random-mail.com",
				Device: requests.Device{
					DeviceToken:    "token123",
					IPAddress:      "192.168.1.1",
					UserAgent:      "Mozilla/5.0",
					OSName:         "Windows",
					OSVersion:      "10",
					BrowserName:    "Chrome",
					BrowserVersion: "120.0",
				},
			},
			setupMocks:         func(m *mocks.MockAuthService) {},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   `{"status":"failure","error":"invalid request data: wrong mail provided"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c, w, _, mockAuthService, _, authController := setupAuthController(t)

			body, _ := json.Marshal(tc.requestBody)
			c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/verification/resend/code", bytes.NewBuffer(body))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Request.Header.Set("Authorization", "Bearer token")

			tc.setupMocks(mockAuthService)
			authController.ResendVerificationCode(c)

			if w.Code != tc.expectedStatusCode {
				t.Errorf("Expected status code %d, got %d. Body: %s", tc.expectedStatusCode, w.Code, w.Body.String())
			}
		})
	}
}

func TestAuthController_ResendSecondFactorCode(t *testing.T) {
	testCases := []struct {
		name               string
		requestBody        requests.Mail
		setupMocks         func(m *mocks.MockAuthService)
		expectedStatusCode int
		expectedResponse   string
	}{
		{
			name: "Success",
			requestBody: requests.Mail{
				Mail: "random@mail.com",
				Device: requests.Device{
					DeviceToken:    "token123",
					IPAddress:      "192.168.1.1",
					UserAgent:      "Mozilla/5.0",
					OSName:         "Windows",
					OSVersion:      "10",
					BrowserName:    "Chrome",
					BrowserVersion: "120.0",
				},
			},
			setupMocks: func(m *mocks.MockAuthService) {
				m.On("ResendSecondFactorCode", mock.Anything, &requests.Mail{
					Mail: "random@mail.com",
					Device: requests.Device{
						DeviceToken:    "token123",
						IPAddress:      "192.168.1.1",
						UserAgent:      "Mozilla/5.0",
						OSName:         "Windows",
						OSVersion:      "10",
						BrowserName:    "Chrome",
						BrowserVersion: "120.0",
					},
				}).Return(nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse:   `{"status":"success","data":"requested verification code resend"}`,
		},
		{
			name: "Invalid Input",
			requestBody: requests.Mail{
				Mail: "random@mail.com",
				Device: requests.Device{
					DeviceToken:    "token123",
					IPAddress:      "192.168.1.1",
					UserAgent:      "Mozilla/5.0",
					OSName:         "Windows",
					OSVersion:      "10",
					BrowserName:    "Chrome",
					BrowserVersion: "120.0",
				},
			},
			setupMocks: func(m *mocks.MockAuthService) {
				m.On("ResendSecondFactorCode", mock.Anything, &requests.Mail{
					Mail: "random@mail.com",
					Device: requests.Device{
						DeviceToken:    "token123",
						IPAddress:      "192.168.1.1",
						UserAgent:      "Mozilla/5.0",
						OSName:         "Windows",
						OSVersion:      "10",
						BrowserName:    "Chrome",
						BrowserVersion: "120.0",
					},
				}).Return(errors.New("failed to update"))
			},
			expectedStatusCode: http.StatusUnprocessableEntity,
			expectedResponse:   `{"error":"failed to update","status":"failure"}`,
		},
		{
			name: "Invalid Credentials",
			requestBody: requests.Mail{
				Mail: "random-mail.com",
				Device: requests.Device{
					DeviceToken:    "token123",
					IPAddress:      "192.168.1.1",
					UserAgent:      "Mozilla/5.0",
					OSName:         "Windows",
					OSVersion:      "10",
					BrowserName:    "Chrome",
					BrowserVersion: "120.0",
				},
			},
			setupMocks:         func(m *mocks.MockAuthService) {},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   `{"status":"failure","error":"invalid request data: wrong mail provided"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c, w, _, mockAuthService, _, authController := setupAuthController(t)

			body, _ := json.Marshal(tc.requestBody)
			c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/second-factor/resend/code", bytes.NewBuffer(body))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Request.Header.Set("Authorization", "Bearer token")

			tc.setupMocks(mockAuthService)
			authController.ResendSecondFactorCode(c)

			if w.Code != tc.expectedStatusCode {
				t.Errorf("Expected status code %d, got %d. Body: %s", tc.expectedStatusCode, w.Code, w.Body.String())
			}
		})
	}
}

func TestAuthController_RequestPasswordReset(t *testing.T) {
	testCases := []struct {
		name               string
		requestBody        requests.Mail
		setupMocks         func(m *mocks.MockAuthService)
		expectedStatusCode int
		expectedResponse   string
	}{
		{
			name: "Success",
			requestBody: requests.Mail{
				Mail: "random@mail.com",
				Device: requests.Device{
					DeviceToken:    "token123",
					IPAddress:      "192.168.1.1",
					UserAgent:      "Mozilla/5.0",
					OSName:         "Windows",
					OSVersion:      "10",
					BrowserName:    "Chrome",
					BrowserVersion: "120.0",
				},
			},
			setupMocks: func(m *mocks.MockAuthService) {
				m.On("RequestPasswordReset", mock.Anything, &requests.Mail{
					Mail: "random@mail.com",
					Device: requests.Device{
						DeviceToken:    "token123",
						IPAddress:      "192.168.1.1",
						UserAgent:      "Mozilla/5.0",
						OSName:         "Windows",
						OSVersion:      "10",
						BrowserName:    "Chrome",
						BrowserVersion: "120.0",
					},
				}).Return(nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse:   `{"status":"success","data":"requested reset code resend"}`,
		},
		{
			name: "Invalid Input",
			requestBody: requests.Mail{
				Mail: "random@mail.com",
				Device: requests.Device{
					DeviceToken:    "token123",
					IPAddress:      "192.168.1.1",
					UserAgent:      "Mozilla/5.0",
					OSName:         "Windows",
					OSVersion:      "10",
					BrowserName:    "Chrome",
					BrowserVersion: "120.0",
				},
			},
			setupMocks: func(m *mocks.MockAuthService) {
				m.On("RequestPasswordReset", mock.Anything, &requests.Mail{
					Mail: "random@mail.com",
					Device: requests.Device{
						DeviceToken:    "token123",
						IPAddress:      "192.168.1.1",
						UserAgent:      "Mozilla/5.0",
						OSName:         "Windows",
						OSVersion:      "10",
						BrowserName:    "Chrome",
						BrowserVersion: "120.0",
					},
				}).Return(errors.New("failed to update"))
			},
			expectedStatusCode: http.StatusUnprocessableEntity,
			expectedResponse:   `{"error":"failed to update","status":"failure"}`,
		},
		{
			name: "Invalid Credentials",
			requestBody: requests.Mail{
				Mail: "random-mail.com",
				Device: requests.Device{
					DeviceToken:    "token123",
					IPAddress:      "192.168.1.1",
					UserAgent:      "Mozilla/5.0",
					OSName:         "Windows",
					OSVersion:      "10",
					BrowserName:    "Chrome",
					BrowserVersion: "120.0",
				},
			},
			setupMocks:         func(m *mocks.MockAuthService) {},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   `{"status":"failure","error":"invalid request data: wrong mail provided"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c, w, _, mockAuthService, _, authController := setupAuthController(t)

			body, _ := json.Marshal(tc.requestBody)
			c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/password/reset/request", bytes.NewBuffer(body))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Request.Header.Set("Authorization", "Bearer token")

			tc.setupMocks(mockAuthService)
			authController.RequestPasswordReset(c)

			if w.Code != tc.expectedStatusCode {
				t.Errorf("Expected status code %d, got %d. Body: %s", tc.expectedStatusCode, w.Code, w.Body.String())
			}
		})
	}
}

func TestAuthController_CancelPasswordReset(t *testing.T) {
	testCases := []struct {
		name               string
		requestBody        requests.Mail
		setupMocks         func(m *mocks.MockAuthService)
		expectedStatusCode int
		expectedResponse   string
	}{
		{
			name: "Success",
			requestBody: requests.Mail{
				Mail: "random@mail.com",
				Device: requests.Device{
					DeviceToken:    "token123",
					IPAddress:      "192.168.1.1",
					UserAgent:      "Mozilla/5.0",
					OSName:         "Windows",
					OSVersion:      "10",
					BrowserName:    "Chrome",
					BrowserVersion: "120.0",
				},
			},
			setupMocks: func(m *mocks.MockAuthService) {
				m.On("CancelPasswordReset", mock.Anything, &requests.Mail{
					Mail: "random@mail.com",
					Device: requests.Device{
						DeviceToken:    "token123",
						IPAddress:      "192.168.1.1",
						UserAgent:      "Mozilla/5.0",
						OSName:         "Windows",
						OSVersion:      "10",
						BrowserName:    "Chrome",
						BrowserVersion: "120.0",
					},
				}).Return(nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse:   `{"status":"success","data":"cancelled password resetting procedure"}`,
		},
		{
			name: "Invalid Input",
			requestBody: requests.Mail{
				Mail: "random@mail.com",
				Device: requests.Device{
					DeviceToken:    "token123",
					IPAddress:      "192.168.1.1",
					UserAgent:      "Mozilla/5.0",
					OSName:         "Windows",
					OSVersion:      "10",
					BrowserName:    "Chrome",
					BrowserVersion: "120.0",
				},
			},
			setupMocks: func(m *mocks.MockAuthService) {
				m.On("CancelPasswordReset", mock.Anything, &requests.Mail{
					Mail: "random@mail.com",
					Device: requests.Device{
						DeviceToken:    "token123",
						IPAddress:      "192.168.1.1",
						UserAgent:      "Mozilla/5.0",
						OSName:         "Windows",
						OSVersion:      "10",
						BrowserName:    "Chrome",
						BrowserVersion: "120.0",
					},
				}).Return(errors.New("failed to update"))
			},
			expectedStatusCode: http.StatusUnprocessableEntity,
			expectedResponse:   `{"error":"failed to update","status":"failure"}`,
		},
		{
			name: "Invalid Credentials",
			requestBody: requests.Mail{
				Mail: "random-mail.com",
				Device: requests.Device{
					DeviceToken:    "token123",
					IPAddress:      "192.168.1.1",
					UserAgent:      "Mozilla/5.0",
					OSName:         "Windows",
					OSVersion:      "10",
					BrowserName:    "Chrome",
					BrowserVersion: "120.0",
				},
			},
			setupMocks:         func(m *mocks.MockAuthService) {},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   `{"status":"failure","error":"invalid request data: wrong mail provided"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c, w, _, mockAuthService, _, authController := setupAuthController(t)

			body, _ := json.Marshal(tc.requestBody)
			c.Request = httptest.NewRequest(http.MethodDelete, "/api/v1/auth/password/reset/request/cancel", bytes.NewBuffer(body))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Request.Header.Set("Authorization", "Bearer token")

			tc.setupMocks(mockAuthService)
			authController.CancelPasswordReset(c)

			if w.Code != tc.expectedStatusCode {
				t.Errorf("Expected status code %d, got %d. Body: %s", tc.expectedStatusCode, w.Code, w.Body.String())
			}
		})
	}
}

func TestAuthController_ConfirmPasswordReset(t *testing.T) {
	testCases := []struct {
		name               string
		requestBody        requests.ConfirmPasswordReset
		setupMocks         func(m *mocks.MockAuthService)
		expectedStatusCode int
		expectedResponse   string
	}{
		{
			name: "Success",
			requestBody: requests.ConfirmPasswordReset{
				Code:     "123123",
				Password: "StrongPassword!@33",
				Device: requests.Device{
					DeviceToken:    "token123",
					IPAddress:      "192.168.1.1",
					UserAgent:      "Mozilla/5.0",
					OSName:         "Windows",
					OSVersion:      "10",
					BrowserName:    "Chrome",
					BrowserVersion: "120.0",
				},
			},
			setupMocks: func(m *mocks.MockAuthService) {
				m.On("ConfirmPasswordReset", mock.Anything, &requests.ConfirmPasswordReset{
					Code:     "123123",
					Password: "StrongPassword!@33",
					Device: requests.Device{
						DeviceToken:    "token123",
						IPAddress:      "192.168.1.1",
						UserAgent:      "Mozilla/5.0",
						OSName:         "Windows",
						OSVersion:      "10",
						BrowserName:    "Chrome",
						BrowserVersion: "120.0",
					},
				}).Return(nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse:   `{"status":"success","data":"confirmed password reset"}`,
		},
		{
			name: "Invalid Input",
			requestBody: requests.ConfirmPasswordReset{
				Code:     "123123",
				Password: "StrongPassword!@33",
				Device: requests.Device{
					DeviceToken:    "token123",
					IPAddress:      "192.168.1.1",
					UserAgent:      "Mozilla/5.0",
					OSName:         "Windows",
					OSVersion:      "10",
					BrowserName:    "Chrome",
					BrowserVersion: "120.0",
				},
			},
			setupMocks: func(m *mocks.MockAuthService) {
				m.On("ConfirmPasswordReset", mock.Anything, &requests.ConfirmPasswordReset{
					Code:     "123123",
					Password: "StrongPassword!@33",
					Device: requests.Device{
						DeviceToken:    "token123",
						IPAddress:      "192.168.1.1",
						UserAgent:      "Mozilla/5.0",
						OSName:         "Windows",
						OSVersion:      "10",
						BrowserName:    "Chrome",
						BrowserVersion: "120.0",
					},
				}).Return(errors.New("failed to update"))
			},
			expectedStatusCode: http.StatusUnprocessableEntity,
			expectedResponse:   `{"error":"failed to update","status":"failure"}`,
		},
		{
			name: "Invalid Credentials 1.1",
			requestBody: requests.ConfirmPasswordReset{
				Code:     "123-123",
				Password: "StrongPassword!@33",
				Device: requests.Device{
					DeviceToken:    "token123",
					IPAddress:      "192.168.1.1",
					UserAgent:      "Mozilla/5.0",
					OSName:         "Windows",
					OSVersion:      "10",
					BrowserName:    "Chrome",
					BrowserVersion: "120.0",
				},
			},
			setupMocks:         func(m *mocks.MockAuthService) {},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   `{"status":"failure","error":"invalid request data: wrong code provided"}`,
		},
		{
			name: "Invalid Credentials 1.2",
			requestBody: requests.ConfirmPasswordReset{
				Code:     "123123",
				Password: "easypassword",
				Device: requests.Device{
					DeviceToken:    "token123",
					IPAddress:      "192.168.1.1",
					UserAgent:      "Mozilla/5.0",
					OSName:         "Windows",
					OSVersion:      "10",
					BrowserName:    "Chrome",
					BrowserVersion: "120.0",
				},
			},
			setupMocks:         func(m *mocks.MockAuthService) {},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   `{"status":"failure","error":"invalid request data: wrong password provided"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c, w, _, mockAuthService, _, authController := setupAuthController(t)

			body, _ := json.Marshal(tc.requestBody)
			c.Request = httptest.NewRequest(http.MethodPatch, "/api/v1/auth/password/reset/request/confirm", bytes.NewBuffer(body))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Request.Header.Set("Authorization", "Bearer token")

			tc.setupMocks(mockAuthService)
			authController.ConfirmPasswordReset(c)

			if w.Code != tc.expectedStatusCode {
				t.Errorf("Expected status code %d, got %d. Body: %s", tc.expectedStatusCode, w.Code, w.Body.String())
			}
		})
	}
}

func TestAuthController_ResendResetCode(t *testing.T) {
	testCases := []struct {
		name               string
		requestBody        requests.Mail
		setupMocks         func(m *mocks.MockAuthService)
		expectedStatusCode int
		expectedResponse   string
	}{
		{
			name: "Success",
			requestBody: requests.Mail{
				Mail: "random@mail.com",
				Device: requests.Device{
					DeviceToken:    "token123",
					IPAddress:      "192.168.1.1",
					UserAgent:      "Mozilla/5.0",
					OSName:         "Windows",
					OSVersion:      "10",
					BrowserName:    "Chrome",
					BrowserVersion: "120.0",
				},
			},
			setupMocks: func(m *mocks.MockAuthService) {
				m.On("ResendResetCode", mock.Anything, &requests.Mail{
					Mail: "random@mail.com",
					Device: requests.Device{
						DeviceToken:    "token123",
						IPAddress:      "192.168.1.1",
						UserAgent:      "Mozilla/5.0",
						OSName:         "Windows",
						OSVersion:      "10",
						BrowserName:    "Chrome",
						BrowserVersion: "120.0",
					},
				}).Return(nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse:   `{"status":"success","data":"requested reset code resend"}`,
		},
		{
			name: "Invalid Input",
			requestBody: requests.Mail{
				Mail: "valid@mail.com",
				Device: requests.Device{
					DeviceToken:    "token123",
					IPAddress:      "192.168.1.1",
					UserAgent:      "Mozilla/5.0",
					OSName:         "Windows",
					OSVersion:      "10",
					BrowserName:    "Chrome",
					BrowserVersion: "120.0",
				},
			},
			setupMocks: func(m *mocks.MockAuthService) {
				m.On("ResendResetCode", mock.Anything, &requests.Mail{
					Mail: "valid@mail.com",
					Device: requests.Device{
						DeviceToken:    "token123",
						IPAddress:      "192.168.1.1",
						UserAgent:      "Mozilla/5.0",
						OSName:         "Windows",
						OSVersion:      "10",
						BrowserName:    "Chrome",
						BrowserVersion: "120.0",
					},
				}).Return(errors.New("failed to update"))
			},
			expectedStatusCode: http.StatusUnprocessableEntity,
			expectedResponse:   `{"error":"failed to update","status":"failure"}`,
		},
		{
			name: "Invalid Credentials",
			requestBody: requests.Mail{
				Mail: "invalid-mail.com",
				Device: requests.Device{
					DeviceToken:    "token123",
					IPAddress:      "192.168.1.1",
					UserAgent:      "Mozilla/5.0",
					OSName:         "Windows",
					OSVersion:      "10",
					BrowserName:    "Chrome",
					BrowserVersion: "120.0",
				},
			},
			setupMocks:         func(m *mocks.MockAuthService) {},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   `{"status":"failure","error":"invalid request data: wrong mail provided"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c, w, _, mockAuthService, _, authController := setupAuthController(t)

			body, _ := json.Marshal(tc.requestBody)
			c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/password/reset/request/resend", bytes.NewBuffer(body))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Request.Header.Set("Authorization", "Bearer token")

			tc.setupMocks(mockAuthService)
			authController.ResendResetCode(c)

			if w.Code != tc.expectedStatusCode {
				t.Errorf("Expected status code %d, got %d. Body: %s", tc.expectedStatusCode, w.Code, w.Body.String())
			}
		})
	}
}
