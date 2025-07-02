package mocks

import "github.com/stretchr/testify/mock"

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
