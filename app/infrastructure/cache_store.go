package infrastructure

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"
	"users-service/app/models"
	"users-service/app/models/ent"
	"users-service/app/models/utils"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"golang.org/x/crypto/chacha20poly1305"
)

type CacheStore struct {
	CipherKey   []byte
	RedisClient CacheClient
	Logger      zerolog.Logger
}

type CacheClient interface {
	Get(ctx context.Context, key string) *redis.StringCmd
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
	Ping(ctx context.Context) *redis.StatusCmd
}

type CacheStoreInterface interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, payload []byte, ttl time.Duration) error
	Del(ctx context.Context, key string) error

	CacheClaims(payload []byte) ([]byte, error)
	CacheUser(user *ent.User) ([]byte, error)

	DecacheClaims(payloadBytes []byte) (*utils.Claims, error)
	DecacheUser(payloadBytes []byte) (*ent.User, error)

	Encrypt(phrase []byte) (string, string, error)
	Decrypt(cipherPhrase string, nonceBase64 string) ([]byte, error)
	Ping(ctx context.Context) *redis.StatusCmd
}

var _ CacheStoreInterface = (*CacheStore)(nil)

func NewCacheStore(redisClient CacheClient, logger zerolog.Logger, encryptionSecretKey string) *CacheStore {
	key := []byte(encryptionSecretKey)
	if len(key) == 0 {
		logger.Fatal().Msg("ENCRYPTION_SECRET_KEY in settings is not set")
	}
	if len(key) != chacha20poly1305.KeySize {
		logger.Fatal().
			Int("expectedLength", chacha20poly1305.KeySize).
			Int("actualLength", len(key)).
			Msg("Key length does not match chacha20poly1305 requirements")
	}
	logger.Info().
		Int("keyLength", len(key)).
		Msg("Successfully initialized CacheStore with encryption key")
	return &CacheStore{
		CipherKey:   key,
		RedisClient: redisClient,
		Logger:      logger,
	}
}

func (s *CacheStore) Get(ctx context.Context, key string) ([]byte, error) {
	s.Logger.Debug().Str("key", key).Msg("Attempting to get and decrypt value from cache")

	encryptedPayloadBytes, err := s.RedisClient.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			s.Logger.Warn().Str("key", key).Msg("Key not found in cache")
			return nil, err
		}
		s.Logger.Error().Err(err).Str("key", key).Msg("Failed to get value from Redis")
		return nil, err
	}
	s.Logger.Debug().Str("key", key).Int("payloadSize", len(encryptedPayloadBytes)).Msg("Retrieved encrypted payload from Redis")

	var cachePayload models.CachePayload
	if err := json.Unmarshal(encryptedPayloadBytes, &cachePayload); err != nil {
		s.Logger.Error().Err(err).Str("key", key).Msg("Failed to unmarshal cache payload from JSON")
		return nil, errors.New("invalid cache format")
	}
	s.Logger.Debug().Str("key", key).Msg("Successfully unmarshaled cache payload")

	plaintext, err := s.Decrypt(cachePayload.CipherText, cachePayload.Nonce)
	if err != nil {
		s.Logger.Error().Err(err).Str("key", key).Msg("Failed to decrypt cache payload")
		return nil, err
	}

	s.Logger.Debug().Str("key", key).Int("plaintextLength", len(plaintext)).Msg("Successfully decrypted value from cache")
	return plaintext, nil
}

func (s *CacheStore) Set(ctx context.Context, key string, payload []byte, ttl time.Duration) error {
	s.Logger.Debug().Str("key", key).Dur("ttl", ttl).Int("payloadLength", len(payload)).Msg("Attempting to encrypt and set value in cache")

	encryptedData, nonce, err := s.Encrypt(payload)
	if err != nil {
		s.Logger.Error().Err(err).Str("key", key).Msg("Failed to encrypt payload for cache")
		return err
	}
	s.Logger.Debug().Str("key", key).Msg("Payload encrypted successfully")

	cachePayload := &models.CachePayload{
		CipherText: encryptedData,
		Nonce:      nonce,
	}

	payloadBytes, err := json.Marshal(cachePayload)
	if err != nil {
		s.Logger.Error().Err(err).Str("key", key).Msg("Failed to marshal cache payload to JSON")
		return err
	}
	s.Logger.Debug().Str("key", key).Int("marshaledSize", len(payloadBytes)).Msg("Cache payload marshaled to JSON")

	if err := s.RedisClient.Set(ctx, key, payloadBytes, ttl).Err(); err != nil {
		s.Logger.Error().Err(err).Str("key", key).Msg("Failed to set value in Redis")
		return err
	}

	s.Logger.Info().Str("key", key).Dur("ttl", ttl).Msg("Successfully set encrypted value in cache")
	return nil
}

func (s *CacheStore) Del(ctx context.Context, key string) error {
	s.Logger.Debug().Str("key", key).Msg("Attempting to delete value from cache")
	err := s.RedisClient.Del(ctx, key).Err()
	if err != nil {
		s.Logger.Error().Err(err).Str("key", key).Msg("Failed to delete value from Redis")
		return err
	}
	s.Logger.Info().Str("key", key).Msg("Successfully deleted value from cache")
	return nil
}

func (s *CacheStore) CacheClaims(payload []byte) ([]byte, error) {
	s.Logger.Debug().Int("payloadLength", len(payload)).Msg("Attempting to cache claims")
	encryptedData, nonce, err := s.Encrypt(payload)
	if err != nil {
		s.Logger.Error().Err(err).Msg("Failed to encrypt claims payload")
		return []byte{}, err
	}
	s.Logger.Debug().Msg("Claims payload encrypted")

	cachePayload := &models.CachePayload{
		CipherText: encryptedData,
		Nonce:      nonce,
	}

	payloadBytes, err := json.Marshal(cachePayload)
	if err != nil {
		s.Logger.Error().Err(err).Msg("Failed to marshal CachePayload for claims")
		return []byte{}, err
	}
	s.Logger.Debug().Int("marshaledSize", len(payloadBytes)).Msg("Claims CachePayload marshaled to JSON")

	return payloadBytes, nil
}

func (s *CacheStore) DecacheClaims(payloadBytes []byte) (*utils.Claims, error) {
	s.Logger.Debug().Int("payloadLength", len(payloadBytes)).Msg("Attempting to decache claims")
	var cachePayload models.CachePayload
	if err := json.Unmarshal(payloadBytes, &cachePayload); err != nil {
		s.Logger.Error().Err(err).Msg("Failed to unmarshal cached claims payload")
		return nil, err
	}
	s.Logger.Debug().Msg("Cached claims payload unmarshaled")

	plaintext, err := s.Decrypt(cachePayload.CipherText, cachePayload.Nonce)
	if err != nil {
		s.Logger.Error().Err(err).Msg("Failed to decrypt cached claims ciphertext")
		return nil, err
	}
	s.Logger.Debug().Int("plaintextLength", len(plaintext)).Msg("Claims plaintext decrypted")

	claims := &utils.Claims{}
	if err := json.Unmarshal(plaintext, claims); err != nil {
		s.Logger.Error().Err(err).Msg("Failed to unmarshal plaintext into Claims struct")
		return nil, err
	}
	s.Logger.Debug().Msg("Claims decached successfully")

	return claims, nil
}

func (s *CacheStore) CacheUser(user *ent.User) ([]byte, error) {
	s.Logger.Debug().Int("userID", user.ID).Msg("Attempting to cache user data")
	userData, err := json.Marshal(user)
	if err != nil {
		s.Logger.Error().Err(err).Int("userID", user.ID).Msg("Failed to marshal user data to JSON for caching")
		return []byte{}, err
	}
	s.Logger.Debug().Int("userID", user.ID).Int("userDataLength", len(userData)).Msg("User data marshaled to JSON")

	encryptedData, nonce, err := s.Encrypt(userData)
	if err != nil {
		s.Logger.Error().Err(err).Int("userID", user.ID).Msg("Failed to encrypt user data")
		return []byte{}, err
	}
	s.Logger.Debug().Int("userID", user.ID).Msg("User data encrypted")

	cachePayload := &models.CachePayload{
		CipherText: encryptedData,
		Nonce:      nonce,
	}

	payloadBytes, err := json.Marshal(cachePayload)
	if err != nil {
		s.Logger.Error().Err(err).Int("userID", user.ID).Msg("Failed to marshal CachePayload for user data")
		return []byte{}, err
	}
	s.Logger.Debug().Int("userID", user.ID).Int("marshaledSize", len(payloadBytes)).Msg("User CachePayload marshaled to JSON")

	return payloadBytes, err
}

func (s *CacheStore) DecacheUser(payloadBytes []byte) (*ent.User, error) {
	s.Logger.Debug().Int("payloadLength", len(payloadBytes)).Msg("Attempting to decache user data")
	var payload models.CachePayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		s.Logger.Error().Err(err).Msg("Failed to unmarshal cached user payload")
		return nil, err
	}
	s.Logger.Debug().Msg("Cached user payload unmarshaled")

	plaintext, err := s.Decrypt(payload.CipherText, payload.Nonce)
	if err != nil {
		s.Logger.Error().Err(err).Msg("Failed to decrypt cached user ciphertext")
		return nil, err
	}
	s.Logger.Debug().Int("plaintextLength", len(plaintext)).Msg("User plaintext decrypted")

	var user ent.User
	if err := json.Unmarshal(plaintext, &user); err != nil {
		s.Logger.Error().Err(err).Msg("Failed to unmarshal plaintext into User struct")
		return nil, err
	}
	s.Logger.Debug().Int("userID", user.ID).Msg("User decached successfully")

	return &user, nil
}

func (s *CacheStore) Encrypt(phrase []byte) (string, string, error) {
	s.Logger.Debug().Int("phraseLength", len(phrase)).Msg("Attempting to encrypt phrase")
	aead, err := chacha20poly1305.NewX(s.CipherKey)
	if err != nil {
		s.Logger.Error().Err(err).Msg("Failed to create AEAD cipher for encryption")
		return "", "", err
	}
	s.Logger.Debug().Msg("AEAD cipher created for encryption")

	nonce, err := s.GenerateNonce()
	if err != nil {
		s.Logger.Error().Err(err).Msg("Failed to generate nonce for encryption")
		return "", "", err
	}
	s.Logger.Debug().Int("nonceLength", len(nonce)).Msg("Nonce generated for encryption")

	ciphertext := aead.Seal(nil, nonce, phrase, nil)
	s.Logger.Debug().Int("ciphertextLength", len(ciphertext)).Msg("Phrase encrypted successfully")
	return base64.StdEncoding.EncodeToString(ciphertext), base64.StdEncoding.EncodeToString(nonce), nil
}

func (s *CacheStore) Decrypt(cipherPhrase string, nonceBase64 string) ([]byte, error) {
	s.Logger.Debug().Int("cipherPhraseLength", len(cipherPhrase)).Msg("Attempting to decrypt phrase")
	aead, err := chacha20poly1305.NewX(s.CipherKey)
	if err != nil {
		s.Logger.Error().Err(err).Msg("Failed to create AEAD cipher for decryption")
		return nil, err
	}
	s.Logger.Debug().Msg("AEAD cipher created for decryption")

	nonce, err := base64.StdEncoding.DecodeString(nonceBase64)
	if err != nil {
		s.Logger.Error().Err(err).Msg("Failed to decode nonce from base64")
		return nil, err
	}
	s.Logger.Debug().Int("nonceLength", len(nonce)).Msg("Nonce decoded")

	ciphertext, err := base64.StdEncoding.DecodeString(cipherPhrase)
	if err != nil {
		s.Logger.Error().Err(err).Msg("Failed to decode ciphertext from base64")
		return nil, err
	}
	s.Logger.Debug().Int("ciphertextLength", len(ciphertext)).Msg("Ciphertext decoded")

	plaintext, err := aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		s.Logger.Error().Err(err).Msg("Failed to open (decrypt) ciphertext. Possible tampering or incorrect key/nonce.")
		return nil, err
	}
	s.Logger.Debug().Int("plaintextLength", len(plaintext)).Msg("Phrase decrypted successfully")

	return plaintext, nil
}

func (s *CacheStore) GenerateNonce() ([]byte, error) {
	s.Logger.Debug().Msg("Attempting to generate nonce")
	nonce := make([]byte, chacha20poly1305.NonceSizeX)

	_, err := rand.Read(nonce)
	if err != nil {
		s.Logger.Error().Err(err).Msg("Failed to read random bytes for nonce generation")
		return nil, err
	}
	s.Logger.Debug().Int("nonceLength", len(nonce)).Msg("Nonce generated successfully")

	return nonce, nil
}

func (s *CacheStore) Ping(ctx context.Context) *redis.StatusCmd {
	s.Logger.Info().Msg("Attempting to ping Redis client")
	cmd := s.RedisClient.Ping(ctx)
	if cmd.Err() != nil {
		s.Logger.Error().Err(cmd.Err()).Msg("Failed to ping Redis client")
	} else {
		s.Logger.Info().Str("status", cmd.Val()).Msg("Redis client ping successful")
	}
	return cmd
}
