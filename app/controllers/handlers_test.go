package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	requests "users-service/app/models/requests"
	"users-service/app/models/responses"
	"users-service/tests/mocks"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHandleStandardRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testCases := []struct {
		name               string
		requestBody        interface{}
		serviceCall        func(reqCtx context.Context, req *mocks.TestRequest) (string, error)
		setupMocks         func(m *mocks.MockStandardController)
		expectedStatusCode int
	}{
		{
			name:        "Success",
			requestBody: mocks.TestRequest{Name: "test"},
			serviceCall: func(reqCtx context.Context, req *mocks.TestRequest) (string, error) {
				assert.Equal(t, "test", req.Name)
				return "success", nil
			},
			setupMocks: func(m *mocks.MockStandardController) {
				m.On("RespondSuccess", mock.Anything, http.StatusOK, "success").Once()
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:        "Failure - Invalid JSON",
			requestBody: "{invalid_json",
			setupMocks: func(m *mocks.MockStandardController) {
				m.On("Log").Return(&zerolog.Logger{}).Once()
				m.On("RespondFailure", mock.Anything, http.StatusBadRequest, "invalid body request").Once()
			},
			expectedStatusCode: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockController := new(mocks.MockStandardController)
			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)

			var reqBodyBytes []byte
			if bodyStr, ok := tc.requestBody.(string); ok {
				reqBodyBytes = []byte(bodyStr)
			} else {
				reqBodyBytes, _ = json.Marshal(tc.requestBody)
			}

			ctx.Request, _ = http.NewRequest(http.MethodPost, "/", bytes.NewBuffer(reqBodyBytes))
			ctx.Request.Header.Set("Content-Type", "application/json")

			tc.setupMocks(mockController)

			handleStandardRequest(ctx, mockController, tc.serviceCall)

			assert.Equal(t, tc.expectedStatusCode, w.Code)
			mockController.AssertExpectations(t)
		})
	}
}

func TestHandleAuthenticatedRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testCases := []struct {
		name               string
		setUserID          bool
		userID             int
		serviceCall        func(reqCtx context.Context, userID int, req *mocks.TestRequest) (string, error)
		setupMocks         func(m *mocks.MockStandardController)
		expectedStatusCode int
	}{
		{
			name:      "Success",
			setUserID: true,
			userID:    123,
			serviceCall: func(reqCtx context.Context, userID int, req *mocks.TestRequest) (string, error) {
				assert.Equal(t, 123, userID)
				assert.Equal(t, "auth_test", req.Name)
				return "authenticated success", nil
			},
			setupMocks: func(m *mocks.MockStandardController) {
				m.On("RespondSuccess", mock.Anything, http.StatusOK, "authenticated success").Once()
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:      "Bad Request",
			setUserID: true,
			userID:    123,
			serviceCall: func(reqCtx context.Context, userID int, req *mocks.TestRequest) (string, error) {
				assert.Equal(t, 123, userID)
				assert.Equal(t, "auth_test", req.Name)
				return "authenticated success", nil
			},
			setupMocks: func(m *mocks.MockStandardController) {
				m.On("RespondSuccess", mock.Anything, http.StatusOK, "authenticated success").Once()
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:      "Failure - Missing UserID in Context",
			setUserID: false,
			setupMocks: func(m *mocks.MockStandardController) {
				m.On("Log").Return(&zerolog.Logger{}).Once()
				m.On("RespondFailure", mock.Anything, http.StatusUnauthorized, "Unauthorized - Invalid or missing token").Once()
			},
			expectedStatusCode: http.StatusUnauthorized,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockController := new(mocks.MockStandardController)
			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)

			if tc.setUserID {
				ctx.Set("UserID", tc.userID)
			}

			reqBody := mocks.TestRequest{Name: "auth_test"}
			jsonBody, _ := json.Marshal(reqBody)
			ctx.Request, _ = http.NewRequest(http.MethodPost, "/", bytes.NewBuffer(jsonBody))
			ctx.Request.Header.Set("Content-Type", "application/json")

			tc.setupMocks(mockController)

			handleAuthenticatedRequest(ctx, mockController, tc.serviceCall)

			assert.Equal(t, tc.expectedStatusCode, w.Code)
			mockController.AssertExpectations(t)
		})
	}
}

func TestHandleWebSocketRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	type TestWSRequest struct {
		Action string `json:"action"`
	}

	testCases := []struct {
		name        string
		setupMocks  func(m *mocks.MockWebSocketController, conn *mocks.MockConn)
		serviceCall func(reqCtx context.Context, req *TestWSRequest, reporter responses.Reporter) (string, error)
	}{
		{
			setupMocks: func(m *mocks.MockWebSocketController, conn *mocks.MockConn) {
				wsHelper := &requests.WebSocketHelper{Conn: conn, Logger: zerolog.Nop()}
				m.On("InitWebSocketHelper", mock.Anything).Return(wsHelper, nil).Once()

				jsonData, _ := json.Marshal(TestWSRequest{Action: "start"})
				conn.On("ReadMessage").Return(websocket.TextMessage, jsonData, nil).Once()

				conn.On("WriteJSON", mock.MatchedBy(func(msg responses.Message) bool {
					t.Logf("WriteJSON called with: %+v", msg)
					return msg.Status == "success" && msg.Message == "success" && msg.Data == "ok" && msg.RequestID == ""
				})).Return(nil).Once()

				conn.On("Close").Return(nil).Once()
				m.On("Log").Return(&zerolog.Logger{}).Maybe()
			},
			serviceCall: func(reqCtx context.Context, req *TestWSRequest, reporter responses.Reporter) (string, error) {
				assert.Equal(t, "start", req.Action)
				reporter.Success("success", "ok")
				return "ok", nil
			},
		},
		{
			name: "Failure - Upgrade Fails",
			setupMocks: func(m *mocks.MockWebSocketController, conn *mocks.MockConn) {
				m.On("InitWebSocketHelper", mock.Anything).Return(nil, errors.New("upgrade failed")).Once()
				m.On("Log").Return(&zerolog.Logger{}).Once()
			},
			serviceCall: func(reqCtx context.Context, req *TestWSRequest, reporter responses.Reporter) (string, error) {
				t.Fail()
				return "", nil
			},
		},
		{
			name: "Failure - Read Request Fails",
			setupMocks: func(m *mocks.MockWebSocketController, conn *mocks.MockConn) {
				wsHelper := &requests.WebSocketHelper{Conn: conn, Logger: zerolog.Nop()}
				m.On("InitWebSocketHelper", mock.Anything).Return(wsHelper, nil).Once()
				conn.On("ReadMessage").Return(0, nil, errors.New("read error")).Once()
				conn.On("WriteJSON", mock.MatchedBy(func(msg responses.Message) bool {
					return msg.Status == "failure" && msg.Message == "Invalid credentials provided" && msg.Error == "read error"
				})).Return(nil).Once()
				conn.On("Close").Return(nil).Once()
				m.On("Log").Return(&zerolog.Logger{}).Maybe()
			},
			serviceCall: func(reqCtx context.Context, req *TestWSRequest, reporter responses.Reporter) (string, error) {
				t.Fail()
				return "", nil
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockController := new(mocks.MockWebSocketController)
			mockConn := new(mocks.MockConn)

			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request, _ = http.NewRequest(http.MethodGet, "/", nil)

			tc.setupMocks(mockController, mockConn)

			handleWebSocketRequest(ctx, mockController, tc.serviceCall)

			time.Sleep(10 * time.Millisecond)

			mockController.AssertExpectations(t)
			mockConn.AssertExpectations(t)
		})
	}
}
