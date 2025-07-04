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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupUserController(t *testing.T) (*gin.Context, *httptest.ResponseRecorder, *mocks.MockUsersService, *UsersController) {
	gin.SetMode(gin.TestMode)

	mockUsersService := new(mocks.MockUsersService)
	logger := zerolog.Nop()

	usersController := NewUsersController(mockUsersService, logger)

	router := gin.New()
	router.PATCH("/api/v1/users/update", usersController.UpdateUser)
	router.PATCH("/api/v1/users/details/update", usersController.UpdateDetails)
	router.PATCH("/api/v1/users/settings/update", usersController.UpdateSettings)
	router.DELETE("/api/v1/users/remove", usersController.RemoveAccount)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	return c, w, mockUsersService, usersController
}

func TestUsersController_UpdateUser(t *testing.T) {
	testCases := []struct {
		name               string
		userID             interface{}
		requestBody        requests.User
		setupMocks         func(m *mocks.MockUsersService)
		expectedStatusCode int
		expectedResponse   string
	}{
		{
			name:   "Success",
			userID: 1,
			requestBody: requests.User{
				Mail:  "test@example.com",
				Phone: "+1234567890",
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
			setupMocks: func(m *mocks.MockUsersService) {
				m.On("UpdateUser", mock.Anything, 1, &requests.User{
					Mail:  "test@example.com",
					Phone: "+1234567890",
					Device: requests.Device{
						DeviceToken:    "token123",
						IPAddress:      "192.168.1.1",
						UserAgent:      "Mozilla/5.0",
						OSName:         "Windows",
						OSVersion:      "10",
						BrowserName:    "Chrome",
						BrowserVersion: "120.0",
					},
				}).Return(&ent.User{Mail: "test@example.com"}, nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse:   `{"status":"success","data":{"user":{"active":false,"blacklisted":false,"created_at":"0001-01-01T00:00:00Z","edges":{"user_details":null,"user_devices":null,"user_roles":null,"user_settings":null},"id":0,"mail":"test@example.com","organization_ids":null,"phone":"","removed":false,"subscription_ids":null,"team_ids":null,"updated_at":"0001-01-01T00:00:00Z","verified":false}}}`,
		},
		{
			name:   "Invalid Input",
			userID: 1,
			requestBody: requests.User{
				Mail:  "",
				Phone: "",
				Device: requests.Device{
					DeviceToken: "",
				},
			},
			setupMocks:         func(m *mocks.MockUsersService) {},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   `{"status":"failure","error":"Bad Request - Invalid input data"}`,
		},
		{
			name:   "Unauthorized",
			userID: nil,
			requestBody: requests.User{
				Mail:  "test@example.com",
				Phone: "+1234567890",
				Device: requests.Device{
					DeviceToken: "token123",
					IPAddress:   "192.168.1.1",
				},
			},
			setupMocks:         func(m *mocks.MockUsersService) {},
			expectedStatusCode: http.StatusUnauthorized,
			expectedResponse:   `{"status":"failure","error":"Unauthorized - Invalid or missing token"}`,
		},
		{
			name:   "Service Error",
			userID: 1,
			requestBody: requests.User{
				Mail:  "test@example.com",
				Phone: "+1234567890",
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
			setupMocks: func(m *mocks.MockUsersService) {
				m.On("UpdateUser", mock.Anything, 1, &requests.User{
					Mail:  "test@example.com",
					Phone: "+1234567890",
					Device: requests.Device{
						DeviceToken:    "token123",
						IPAddress:      "192.168.1.1",
						UserAgent:      "Mozilla/5.0",
						OSName:         "Windows",
						OSVersion:      "10",
						BrowserName:    "Chrome",
						BrowserVersion: "120.0",
					},
				}).Return(nil, errors.New("update failed"))
			},
			expectedStatusCode: http.StatusUnprocessableEntity,
			expectedResponse:   `{"status":"failure","error":"Unprocessable Entity - Update failed"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c, w, mockUsersService, usersController := setupUserController(t)

			if tc.userID != nil {
				c.Set("UserID", tc.userID)
			}
			c.Set("requestID", "")

			body, _ := json.Marshal(tc.requestBody)
			c.Request = httptest.NewRequest(http.MethodPatch, "/api/v1/users/update", bytes.NewBuffer(body))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Request.Header.Set("Authorization", "Bearer token")
			c.Params = []gin.Param{{Key: "id", Value: "1"}}

			tc.setupMocks(mockUsersService)
			usersController.UpdateUser(c)

			if w.Code != tc.expectedStatusCode {
				t.Errorf("Expected status code %d, got %d. Body: %s, Headers: %v, Request: %v", tc.expectedStatusCode, w.Code, w.Body.String(), w.Header(), c.Request)
			}
			if !assert.JSONEq(t, tc.expectedResponse, w.Body.String()) {
				t.Errorf("Expected response %s, got %s, Headers: %v", tc.expectedResponse, w.Body.String(), w.Header())
			}
		})
	}
}

func TestUsersController_UpdateDetails(t *testing.T) {
	testCases := []struct {
		name               string
		userID             interface{}
		requestBody        requests.Details
		setupMocks         func(m *mocks.MockUsersService)
		expectedStatusCode int
		expectedResponse   string
	}{
		{
			name:   "Success",
			userID: 1,
			requestBody: requests.Details{
				FirstName: "John",
				LastName:  "Doe",
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
			setupMocks: func(m *mocks.MockUsersService) {
				m.On("UpdateDetails", mock.Anything, 1, &requests.Details{
					FirstName: "John",
					LastName:  "Doe",
					Device: requests.Device{
						DeviceToken:    "token123",
						IPAddress:      "192.168.1.1",
						UserAgent:      "Mozilla/5.0",
						OSName:         "Windows",
						OSVersion:      "10",
						BrowserName:    "Chrome",
						BrowserVersion: "120.0",
					},
				}).
					Return(&ent.User{Mail: "test@example.com"}, nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse:   `{"status":"success","data":{"user":{"active":false,"blacklisted":false,"created_at":"0001-01-01T00:00:00Z","edges":{"user_details":null,"user_devices":null,"user_roles":null,"user_settings":null},"id":0,"mail":"test@example.com","organization_ids":null,"phone":"","removed":false,"subscription_ids":null,"team_ids":null,"updated_at":"0001-01-01T00:00:00Z","verified":false}}}`,
		},
		{
			name:   "Invalid Credentials",
			userID: 1,
			requestBody: requests.Details{
				FirstName: "",
				LastName:  "",
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
			setupMocks:         func(m *mocks.MockUsersService) {},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   `{"status":"failure","error":"Bad Request - Invalid input data"}`,
		},
		{
			name:   "Unauthorized",
			userID: nil,
			requestBody: requests.Details{
				FirstName: "John",
				LastName:  "Doe",
				Device: requests.Device{
					DeviceToken:    "token123",
					IPAddress:      "192.168.1.1",
					UserAgent:      "Mozilla/5.0",
					OSName:         "Windows",
					OSVersion:      "10",
					BrowserName:    "Chrome",
					BrowserVersion: "120.0",
				}},
			setupMocks:         func(m *mocks.MockUsersService) {},
			expectedStatusCode: http.StatusUnauthorized,
			expectedResponse:   `{"status":"failure","error":"Unauthorized - Invalid or missing token"}`,
		},
		{
			name:   "Service Error",
			userID: 1,
			requestBody: requests.Details{
				FirstName: "John",
				LastName:  "Doe",
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
			setupMocks: func(m *mocks.MockUsersService) {
				m.On("UpdateDetails", mock.Anything, 1, &requests.Details{
					FirstName: "John",
					LastName:  "Doe",
					Device: requests.Device{
						DeviceToken:    "token123",
						IPAddress:      "192.168.1.1",
						UserAgent:      "Mozilla/5.0",
						OSName:         "Windows",
						OSVersion:      "10",
						BrowserName:    "Chrome",
						BrowserVersion: "120.0",
					},
				}).
					Return(nil, errors.New("update failed"))
			},
			expectedStatusCode: http.StatusUnprocessableEntity,
			expectedResponse:   `{"status":"failure","error":"Unprocessable Entity - Update failed"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c, w, mockUsersService, usersController := setupUserController(t)

			if tc.userID != nil {
				c.Set("UserID", tc.userID)
			} else {
				c.Set("UserID", "")
			}
			c.Set("RequestID", "")

			body, _ := json.Marshal(tc.requestBody)
			c.Request = httptest.NewRequest(http.MethodPatch, "/api/v1/users/details/update", bytes.NewBuffer(body))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Request.Header.Set("Authorization", "Bearer token")

			tc.setupMocks(mockUsersService)
			usersController.UpdateDetails(c)

			if w.Code != tc.expectedStatusCode {
				t.Errorf("Expected status code %d, got %d. Body: %s", tc.expectedStatusCode, w.Code, w.Body.String())
			}
			if !assert.JSONEq(t, tc.expectedResponse, w.Body.String()) {
				t.Errorf("Expected response %s, got %s", tc.expectedResponse, w.Body.String())
			}
		})
	}
}

func TestUsersController_UpdateSettings(t *testing.T) {
	testCases := []struct {
		name               string
		userID             interface{}
		requestBody        requests.Settings
		setupMocks         func(m *mocks.MockUsersService)
		expectedStatusCode int
		expectedResponse   string
	}{
		{
			name:   "Success",
			userID: 1,
			requestBody: requests.Settings{
				TwoFactor:                    true,
				NightMode:                    true,
				SecondFactorTargetID:         1,
				NotificationsTargetDeviceIDs: []int{},
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

			setupMocks: func(m *mocks.MockUsersService) {
				m.On("UpdateSettings", mock.Anything, 1, &requests.Settings{
					TwoFactor:                    true,
					NightMode:                    true,
					SecondFactorTargetID:         1,
					NotificationsTargetDeviceIDs: []int{},
					Device: requests.Device{
						DeviceToken:    "token123",
						IPAddress:      "192.168.1.1",
						UserAgent:      "Mozilla/5.0",
						OSName:         "Windows",
						OSVersion:      "10",
						BrowserName:    "Chrome",
						BrowserVersion: "120.0",
					}}).
					Return(&ent.User{Mail: "test@example.com"}, nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse:   `{"status":"success","data":{"user":{"active":false,"blacklisted":false,"created_at":"0001-01-01T00:00:00Z","edges":{"user_details":null,"user_devices":null,"user_roles":null,"user_settings":null},"id":0,"mail":"test@example.com","organization_ids":null,"phone":"","removed":false,"subscription_ids":null,"team_ids":null,"updated_at":"0001-01-01T00:00:00Z","verified":false}}}`,
		},
		{
			name:   "Invalid Input",
			userID: 1,
			requestBody: requests.Settings{
				TwoFactor:                    true,
				NightMode:                    true,
				SecondFactorTargetID:         -1,
				NotificationsTargetDeviceIDs: []int{},
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
			setupMocks:         func(m *mocks.MockUsersService) {},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   `{"status":"failure","error":"Bad Request - Invalid input data"}`,
		},
		{
			name:   "Unauthorized",
			userID: nil,
			requestBody: requests.Settings{
				TwoFactor:                    true,
				NightMode:                    true,
				SecondFactorTargetID:         1,
				NotificationsTargetDeviceIDs: []int{},
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
			setupMocks:         func(m *mocks.MockUsersService) {},
			expectedStatusCode: http.StatusUnauthorized,
			expectedResponse:   `{"status":"failure","error":"Unauthorized - Invalid or missing token"}`,
		},
		{
			name:   "Service Error",
			userID: 1,
			requestBody: requests.Settings{
				TwoFactor:                    true,
				NightMode:                    true,
				SecondFactorTargetID:         1,
				NotificationsTargetDeviceIDs: []int{},
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
			setupMocks: func(m *mocks.MockUsersService) {
				m.On("UpdateSettings", mock.Anything, 1, &requests.Settings{
					TwoFactor:                    true,
					NightMode:                    true,
					SecondFactorTargetID:         1,
					NotificationsTargetDeviceIDs: []int{},
					Device: requests.Device{
						DeviceToken:    "token123",
						IPAddress:      "192.168.1.1",
						UserAgent:      "Mozilla/5.0",
						OSName:         "Windows",
						OSVersion:      "10",
						BrowserName:    "Chrome",
						BrowserVersion: "120.0",
					}}).
					Return(nil, errors.New("update failed"))
			},
			expectedStatusCode: http.StatusUnprocessableEntity,
			expectedResponse:   `{"status":"failure","error":"Unprocessable Entity - Update failed"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c, w, mockUsersService, usersController := setupUserController(t)

			if tc.userID != nil {
				c.Set("UserID", tc.userID)
			}
			c.Set("RequestID", "")

			body, _ := json.Marshal(tc.requestBody)
			c.Request = httptest.NewRequest(http.MethodPatch, "/api/v1/users/settings/update", bytes.NewBuffer(body))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Request.Header.Set("Authorization", "Bearer token")

			tc.setupMocks(mockUsersService)
			usersController.UpdateSettings(c)

			if w.Code != tc.expectedStatusCode {
				t.Errorf("Expected status code %d, got %d. Body: %s", tc.expectedStatusCode, w.Code, w.Body.String())
			}
			if !assert.JSONEq(t, tc.expectedResponse, w.Body.String()) {
				t.Errorf("Expected response %s, got %s", tc.expectedResponse, w.Body.String())
			}
		})
	}
}

func TestUsersController_RemoveAccount(t *testing.T) {
	testCases := []struct {
		name               string
		userID             interface{}
		requestBody        requests.Empty
		setupMocks         func(m *mocks.MockUsersService)
		expectedStatusCode int
		expectedResponse   string
	}{
		{
			name:   "Success",
			userID: 1,
			requestBody: requests.Empty{
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
			setupMocks: func(m *mocks.MockUsersService) {
				m.On("RemoveAccount", mock.Anything, 1).Return(nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse:   `{"status":"success","data":"successfully removed user"}`,
		},
		{
			name:   "Unauthorized",
			userID: nil,
			requestBody: requests.Empty{
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
			setupMocks:         func(m *mocks.MockUsersService) {},
			expectedStatusCode: http.StatusUnauthorized,
			expectedResponse:   `{"status":"failure","error":"Unauthorized - Invalid or missing token"}`,
		},
		{
			name:   "Service Error",
			userID: 1,
			requestBody: requests.Empty{
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
			setupMocks: func(m *mocks.MockUsersService) {
				m.On("RemoveAccount", mock.Anything, 1).Return(errors.New("failed to update"))
			},
			expectedStatusCode: http.StatusUnprocessableEntity,
			expectedResponse:   `{"status":"failure","error":"Unprocessable Entity - Update failed"}`,
		},
		{
			name:   "Invalid Credentials",
			userID: 1,
			requestBody: requests.Empty{
				Device: requests.Device{
					DeviceToken:    "token123",
					IPAddress:      "",
					UserAgent:      "Mozilla/5.0",
					OSName:         "Windows",
					OSVersion:      "10",
					BrowserName:    "Chrome",
					BrowserVersion: "120.0",
				},
			},
			setupMocks:         func(m *mocks.MockUsersService) {},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   `{"status":"failure","error":"Bad Request - Invalid input data"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c, w, mockUsersService, usersController := setupUserController(t)

			if tc.userID != nil {
				c.Set("UserID", tc.userID)
			} else {
				c.Set("UserID", "")
			}

			c.Set("RequestID", "")

			body, _ := json.Marshal(tc.requestBody)
			c.Request = httptest.NewRequest(http.MethodDelete, "/api/v1/users/remove", bytes.NewBuffer(body))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Request.Header.Set("Authorization", "Bearer token")

			tc.setupMocks(mockUsersService)
			usersController.RemoveAccount(c)

			if w.Code != tc.expectedStatusCode {
				t.Errorf("Expected status code %d, got %d. Body: %s", tc.expectedStatusCode, w.Code, w.Body.String())
			}
			if !assert.JSONEq(t, tc.expectedResponse, w.Body.String()) {
				t.Errorf("Expected response %s, got %s", tc.expectedResponse, w.Body.String())
			}
		})
	}
}
