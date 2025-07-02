package mocks

import (
	"context"
	"time"
	"users-service/app/models/ent"
	"users-service/app/models/utils"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/mock"
)

type MockRedisClient struct {
	mock.Mock
}

func (m *MockRedisClient) Get(ctx context.Context, key string) *redis.StringCmd {
	args := m.Called(ctx, key)
	if ret := args.Get(0); ret != nil {
		return ret.(*redis.StringCmd)
	}
	return nil
}

func (m *MockRedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	args := m.Called(ctx, key, value, expiration)
	if ret := args.Get(0); ret != nil {
		return ret.(*redis.StatusCmd)
	}
	return nil
}

func (m *MockRedisClient) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	args := m.Called(ctx, keys)
	if ret := args.Get(0); ret != nil {
		return ret.(*redis.IntCmd)
	}
	return nil
}

type MockCacheStore struct {
	mock.Mock
}

func (m *MockCacheStore) Get(ctx context.Context, key string) ([]byte, error) {
	args := m.Called(ctx, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockCacheStore) Set(ctx context.Context, key string, payload []byte, ttl time.Duration) error {
	args := m.Called(ctx, key, payload, ttl)
	return args.Error(0)
}

func (m *MockCacheStore) Del(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockCacheStore) CacheClaims(payload []byte) ([]byte, error) {
	args := m.Called(payload)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockCacheStore) CacheUser(user *ent.User) ([]byte, error) {
	args := m.Called(user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockCacheStore) DecacheClaims(payloadBytes []byte) (*utils.Claims, error) {
	args := m.Called(payloadBytes)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*utils.Claims), args.Error(1)
}

func (m *MockCacheStore) DecacheUser(payloadBytes []byte) (*ent.User, error) {
	args := m.Called(payloadBytes)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ent.User), args.Error(1)
}

func (m *MockCacheStore) Encrypt(phrase []byte) (string, string, error) {
	args := m.Called(phrase)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *MockCacheStore) Decrypt(cipherPhrase string, nonceBase64 string) ([]byte, error) {
	args := m.Called(cipherPhrase, nonceBase64)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}
