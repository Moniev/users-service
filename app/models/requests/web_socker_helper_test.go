//go:build unit

package requests

import (
	"encoding/json"
	"errors"
	"testing"
	"users-service/app/models/responses"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockConn struct {
	mock.Mock
}

func (m *MockConn) WriteJSON(v interface{}) error {
	args := m.Called(v)
	return args.Error(0)
}

func (m *MockConn) ReadMessage() (int, []byte, error) {
	args := m.Called()
	var p []byte
	if args.Get(1) != nil {
		p = args.Get(1).([]byte)
	}
	return args.Int(0), p, args.Error(2)
}

func (m *MockConn) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestSendMessage_Success(t *testing.T) {
	mockConn := new(MockConn)
	helper := &WebSocketHelper{Conn: mockConn, Logger: zerolog.Nop()}
	testMsg := responses.Message{Status: "status", Message: "data"}

	mockConn.On("WriteJSON", testMsg).Return(nil)

	helper.SendMessage(testMsg)

	mockConn.AssertExpectations(t)
}

func TestReadRequest_Success(t *testing.T) {
	mockConn := new(MockConn)
	helper := &WebSocketHelper{Conn: mockConn, Logger: zerolog.Nop()}

	type TestRequest struct {
		Name string `json:"name"`
	}
	expectedRequest := TestRequest{Name: "test"}
	jsonData, err := json.Marshal(expectedRequest)
	require.NoError(t, err)

	mockConn.On("ReadMessage").Return(websocket.TextMessage, jsonData, nil)

	var actualRequest TestRequest
	err = helper.ReadRequest(&actualRequest)

	require.NoError(t, err)
	assert.Equal(t, expectedRequest, actualRequest)
	mockConn.AssertExpectations(t)
}

func TestReadRequest_ReadError(t *testing.T) {
	mockConn := new(MockConn)
	helper := &WebSocketHelper{Conn: mockConn, Logger: zerolog.Nop()}
	expectedError := errors.New("connection closed")

	mockConn.On("ReadMessage").Return(0, nil, expectedError)

	var dummyRequest struct{}
	err := helper.ReadRequest(&dummyRequest)

	require.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockConn.AssertExpectations(t)
}
