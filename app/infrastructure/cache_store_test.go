//go:build unit

package infrastructure

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"
	"users-service/app/models"
	"users-service/app/models/ent"
	"users-service/tests/mocks"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var _ CacheClient = (*mocks.MockRedisClient)(nil)

func newTestCacheStore(t *testing.T) (*CacheStore, *mocks.MockRedisClient) {
	key := "a-valid-32-byte-long-secret-key!"
	require.Len(t, []byte(key), 32)

	mockClient := new(mocks.MockRedisClient)

	store := NewCacheStore(mockClient, zerolog.Nop(), key)
	require.NotNil(t, store)

	return store, mockClient
}

func createMockStatusCmd(val string, err error) *redis.StatusCmd {
	cmd := redis.NewStatusCmd(context.TODO())
	cmd.SetVal(val)
	cmd.SetErr(err)
	return cmd
}

func TestCacheStore_Get(t *testing.T) {
	store, mockClient := newTestCacheStore(t)
	ctx := context.Background()
	key := "test:get:key"
	originalPayload := []byte("my-secret-data")

	encryptedData, nonce, err := store.Encrypt(originalPayload)
	require.NoError(t, err)
	cachePayload := models.CachePayload{CipherText: encryptedData, Nonce: nonce}
	encryptedBytes, err := json.Marshal(cachePayload)
	require.NoError(t, err)

	testCases := []struct {
		name            string
		setupMock       func()
		expectedPayload []byte
		expectErr       bool
		expectedErrIs   error
		expectedErrMsg  string
	}{
		{
			name: "Success - Key Found",
			setupMock: func() {
				mockClient.On("Get", ctx, key).Return(redis.NewStringResult(string(encryptedBytes), nil)).Once()
			},
			expectedPayload: originalPayload,
			expectErr:       false,
		},
		{
			name: "Failure - Key Not Found",
			setupMock: func() {
				mockClient.On("Get", ctx, key).Return(redis.NewStringResult("", redis.Nil)).Once()
			},
			expectedPayload: nil,
			expectErr:       true,
			expectedErrIs:   redis.Nil,
		},
		{
			name: "Failure - Redis Error",
			setupMock: func() {
				mockClient.On("Get", ctx, key).Return(redis.NewStringResult("", errors.New("connection error"))).Once()
			},
			expectedPayload: nil,
			expectErr:       true,
			expectedErrMsg:  "connection error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock()

			decryptedPayload, err := store.Get(ctx, key)

			if tc.expectErr {
				require.Error(t, err)
				if tc.expectedErrIs != nil {
					assert.ErrorIs(t, err, tc.expectedErrIs)
				}
				if tc.expectedErrMsg != "" {
					assert.EqualError(t, err, tc.expectedErrMsg)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expectedPayload, decryptedPayload)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestCacheStore_Set(t *testing.T) {
	ctx := context.Background()
	key := "test:set:key"
	payload := []byte("test-payload")
	ttl := 5 * time.Minute

	testCases := []struct {
		name      string
		setupMock func(m *mocks.MockRedisClient)
		expectErr bool
	}{
		{
			name: "Success",
			setupMock: func(m *mocks.MockRedisClient) {
				m.On("Set", ctx, key, mock.Anything, ttl).Return(redis.NewStatusResult("OK", nil)).Once()
			},
			expectErr: false,
		},
		{
			name: "Failure - Redis Error",
			setupMock: func(m *mocks.MockRedisClient) {
				m.On("Set", ctx, key, mock.Anything, ttl).Return(redis.NewStatusResult("", errors.New("connection error"))).Once()
			},
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			store, mockClient := newTestCacheStore(t)
			tc.setupMock(mockClient)

			err := store.Set(ctx, key, payload, ttl)

			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestCacheStore_Del(t *testing.T) {
	ctx := context.Background()
	key := "test:del:key"

	testCases := []struct {
		name      string
		setupMock func(m *mocks.MockRedisClient)
		expectErr bool
	}{
		{
			name: "Success",
			setupMock: func(m *mocks.MockRedisClient) {
				m.On("Del", ctx, []string{key}).Return(redis.NewIntResult(1, nil)).Once()
			},
			expectErr: false,
		},
		{
			name: "Failure - Redis Error",
			setupMock: func(m *mocks.MockRedisClient) {
				m.On("Del", ctx, []string{key}).Return(redis.NewIntResult(0, errors.New("connection error"))).Once()
			},
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			store, mockClient := newTestCacheStore(t)
			tc.setupMock(mockClient)

			err := store.Del(ctx, key)

			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestEncryptDecrypt(t *testing.T) {
	store, _ := newTestCacheStore(t)
	originalText := "this is a secret message"

	cipherText, nonce, err := store.Encrypt([]byte(originalText))
	require.NoError(t, err)
	decryptedText, err := store.Decrypt(cipherText, nonce)
	require.NoError(t, err)

	assert.Equal(t, originalText, string(decryptedText))
}

func TestCacheStore_DecacheUser(t *testing.T) {
	store, _ := newTestCacheStore(t)

	originalUser := &ent.User{ID: 1, Mail: "test@example.com", Active: true}

	cachedBytes, err := store.CacheUser(originalUser)
	require.NoError(t, err)
	decachedUser, err := store.DecacheUser(cachedBytes)
	require.NoError(t, err)

	assert.Equal(t, originalUser.ID, decachedUser.ID)
	assert.Equal(t, originalUser.Mail, decachedUser.Mail)
}

func TestCacheStore_Ping(t *testing.T) {
	testCases := []struct {
		name        string
		setupMock   func(m *mocks.MockRedisClient)
		expectErr   bool
		errContains string
	}{
		{
			name: "Success",
			setupMock: func(m *mocks.MockRedisClient) {
				m.On("Ping", mock.Anything).Return(createMockStatusCmd("PONG", nil)).Once()
			},
			expectErr: false,
		},
		{
			name: "Connection Error",
			setupMock: func(m *mocks.MockRedisClient) {
				m.On("Ping", mock.Anything).Return(createMockStatusCmd("", errors.New("connection refused"))).Once()
			},
			expectErr:   true,
			errContains: "connection refused",
		},
		{
			name: "Timeout Error",
			setupMock: func(m *mocks.MockRedisClient) {
				m.On("Ping", mock.Anything).Return(createMockStatusCmd("", context.DeadlineExceeded)).Once()
			},
			expectErr:   true,
			errContains: "context deadline exceeded",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			store, mockRedisClient := newTestCacheStore(t)
			tc.setupMock(mockRedisClient)

			ctx := context.Background()

			cmd := store.Ping(ctx)
			_, err := cmd.Result()

			if tc.expectErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errContains)
			} else {
				require.NoError(t, err)
			}

			mockRedisClient.AssertExpectations(t)
		})
	}
}
