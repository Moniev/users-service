//go:build unit

package controllers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"users-service/tests/mocks"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupDiagnosticsController(t *testing.T) (*gin.Context, *httptest.ResponseRecorder, *mocks.MockDiagnosticsService, *DiagnosticsController) {
	gin.SetMode(gin.TestMode)

	mockDiagnosticsService := new(mocks.MockDiagnosticsService)
	logger := zerolog.Nop()

	diagnosticsController := NewDiagnosticsController(mockDiagnosticsService, logger)

	router := gin.New()
	router.GET("/api/v1/diagnostics/readiness", diagnosticsController.ReadinessProbe)
	router.GET("/api/v1/diagnostics/health", diagnosticsController.HealthProbe)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	return c, w, mockDiagnosticsService, diagnosticsController
}

func TestDiagnosticsController_ReadinessProbe(t *testing.T) {
	testCases := []struct {
		name               string
		setupMocks         func(m *mocks.MockDiagnosticsService)
		expectedStatusCode int
		expectedResponse   string
	}{
		{
			name: "Success",
			setupMocks: func(m *mocks.MockDiagnosticsService) {
				m.On("CheckReadiness", mock.Anything).Return(nil).Once()
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse:   `{"status":"success","data":{"message":"Service is ready"}}`,
		},
		{
			name: "Service Not Ready",
			setupMocks: func(m *mocks.MockDiagnosticsService) {
				m.On("CheckReadiness", mock.Anything).Return(errors.New("redis connection failed")).Once()
			},
			expectedStatusCode: http.StatusServiceUnavailable,
			expectedResponse:   `{"status":"failure","error":"Service not ready: redis connection failed"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c, w, mockDiagnosticsService, diagnosticsController := setupDiagnosticsController(t)

			c.Set("requestID", "")
			c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/diagnostics/readiness", nil)
			c.Request.Header.Set("Content-Type", "application/json")

			tc.setupMocks(mockDiagnosticsService)
			diagnosticsController.ReadinessProbe(c)

			if w.Code != tc.expectedStatusCode {
				t.Errorf("Expected status code %d, got %d. Body: %s", tc.expectedStatusCode, w.Code, w.Body.String())
			}
			if !assert.JSONEq(t, tc.expectedResponse, w.Body.String()) {
				t.Errorf("Expected response %s, got %s", tc.expectedResponse, w.Body.String())
			}

			mockDiagnosticsService.AssertExpectations(t)
		})
	}
}

func TestDiagnosticsController_HealthProbe(t *testing.T) {
	testCases := []struct {
		name               string
		setupMocks         func(m *mocks.MockDiagnosticsService)
		expectedStatusCode int
		expectedResponse   string
	}{
		{
			name: "Success",
			setupMocks: func(m *mocks.MockDiagnosticsService) {
				m.On("CheckHealth", mock.Anything).Return(nil).Once()
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse:   `{"status":"success","data":{"message":"Service is healthy"}}`,
		},
		{
			name: "Service Unhealthy",
			setupMocks: func(m *mocks.MockDiagnosticsService) {
				m.On("CheckHealth", mock.Anything).Return(errors.New("database connection failed")).Once()
			},
			expectedStatusCode: http.StatusServiceUnavailable,
			expectedResponse:   `{"status":"failure","error":"Service unhealthy: database connection failed"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c, w, mockDiagnosticsService, diagnosticsController := setupDiagnosticsController(t)

			c.Set("requestID", "")
			c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/diagnostics/health", nil)
			c.Request.Header.Set("Content-Type", "application/json")

			tc.setupMocks(mockDiagnosticsService)
			diagnosticsController.HealthProbe(c)

			if w.Code != tc.expectedStatusCode {
				t.Errorf("Expected status code %d, got %d. Body: %s", tc.expectedStatusCode, w.Code, w.Body.String())
			}
			if !assert.JSONEq(t, tc.expectedResponse, w.Body.String()) {
				t.Errorf("Expected response %s, got %s", tc.expectedResponse, w.Body.String())
			}

			mockDiagnosticsService.AssertExpectations(t)
		})
	}
}
