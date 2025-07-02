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
		Int("KeyLength", len(key)).
		Msg("Successfully initialized Utils with encryption key")
	return &CacheStore{
		CipherKey:   key,
		RedisClient: redisClient,
		Logger:      logger,
	}
}

func (s *CacheStore) Get(ctx context.Context, key string) ([]byte, error) {
	s.Logger.Debug().Str("key", key).Msg("Getting and decrypting value from cache")

	encryptedPayloadBytes, err := s.RedisClient.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			s.Logger.Warn().Str("key", key).Msg("Key not found in cache")
			return nil, err
		}
		s.Logger.Error().Err(err).Str("key", key).Msg("Failed to get value from cache")
		return nil, err
	}

	var cachePayload models.CachePayload
	if err := json.Unmarshal(encryptedPayloadBytes, &cachePayload); err != nil {
		s.Logger.Error().Err(err).Str("key", key).Msg("Failed to unmarshal cache payload")
		return nil, errors.New("invalid cache format")
	}

	plaintext, err := s.Decrypt(cachePayload.CipherText, cachePayload.Nonce)
	if err != nil {
		s.Logger.Error().Err(err).Str("key", key).Msg("Failed to decrypt cache payload")
		return nil, err
	}

	s.Logger.Debug().Str("key", key).Msg("Successfully decrypted value from cache")
	return plaintext, nil
}

func (s *CacheStore) Set(ctx context.Context, key string, payload []byte, ttl time.Duration) error {
	s.Logger.Debug().Str("key", key).Dur("ttl", ttl).Msg("Encrypting and setting value in cache")

	encryptedData, nonce, err := s.Encrypt(payload)
	if err != nil {
		s.Logger.Error().Err(err).Str("key", key).Msg("Failed to encrypt payload for cache")
		return err
	}

	cachePayload := &models.CachePayload{
		CipherText: encryptedData,
		Nonce:      nonce,
	}

	payloadBytes, err := json.Marshal(cachePayload)
	if err != nil {
		s.Logger.Error().Err(err).Str("key", key).Msg("Failed to marshal cache payload")
		return err
	}

	if err := s.RedisClient.Set(ctx, key, payloadBytes, ttl).Err(); err != nil {
		s.Logger.Error().Err(err).Str("key", key).Msg("Failed to set value in cache")
		return err
	}

	s.Logger.Debug().Str("key", key).Msg("Successfully set encrypted value in cache")
	return nil
}

func (s *CacheStore) Del(ctx context.Context, key string) error {
	s.Logger.Debug().Str("key", key).Msg("Deleting value from cache")
	err := s.RedisClient.Del(ctx, key).Err()
	if err != nil {
		s.Logger.Error().Err(err).Str("key", key).Msg("Failed to delete value from cache")
		return err
	}
	return nil
}

// CacheClaims encrypts and caches JWT claims from a byte payload.
// It encrypts the provided payload using the Utils.Encrypt method and wraps the encrypted data
// along with the nonce in a CachePayload structure. The final payload is serialized to JSON bytes
// for storage.
//
// Parameters:
//   - payload: The raw byte data representing the claims to be encrypted.
//
// Returns:
//   - []byte: The encrypted and serialized payload ready for caching.
//   - error: An error if encryption or marshaling fails.
func (s *CacheStore) CacheClaims(payload []byte) ([]byte, error) {
	encryptedData, nonce, err := s.Encrypt(payload)
	if err != nil {
		return []byte{}, err
	}

	cachePayload := &models.CachePayload{
		CipherText: encryptedData,
		Nonce:      nonce,
	}

	payloadBytes, err := json.Marshal(cachePayload)
	if err != nil {
		return []byte{}, err
	}

	return payloadBytes, nil
}

// DecacheClaims decrypts and retrieves JWT claims from a cached payload.
// It deserializes the cached payload into a CachePayload structure, decrypts the ciphertext
// using the Utils.Decrypt method with the provided nonce, and unmarshals the decrypted data
// into the provided claims structure.
//
// Parameters:
//   - payloadBytes: The encrypted payload bytes retrieved from cache (e.g., Redis).
//   - claims: A pointer to the claims structure where decrypted data will be unmarshaled (e.g., *JWTClaims).
//
// Returns:
//   - error: An error if deserialization, decryption, or unmarshaling fails.
func (s *CacheStore) DecacheClaims(payloadBytes []byte) (*utils.Claims, error) {
	var cachePayload models.CachePayload
	if err := json.Unmarshal(payloadBytes, &cachePayload); err != nil {
		return nil, err
	}

	plaintext, err := s.Decrypt(cachePayload.CipherText, cachePayload.Nonce)
	if err != nil {
		return nil, err
	}

	claims := &utils.Claims{}
	if err := json.Unmarshal(plaintext, claims); err != nil {
		return nil, err
	}

	return claims, nil
}

// CacheUser encrypts and caches user data in a JSON-encoded payload.
// It marshals the user data to JSON, encrypts it using the utils' Encrypt method,
// and creates a UserCachePayload containing the encrypted data and nonce.
//
// Parameters:
//   - user *ent.User: The user entity to be cached
//
// Returns:
//   - []byte: JSON-encoded UserCachePayload containing encrypted user data and nonce
//   - error: Any error that occurred during marshalling or encryption
func (s *CacheStore) CacheUser(user *ent.User) ([]byte, error) {
	userData, err := json.Marshal(user)
	if err != nil {
		return []byte{}, err
	}

	encryptedData, nonce, err := s.Encrypt(userData)
	if err != nil {
		return []byte{}, err
	}

	cachePayload := &models.CachePayload{
		CipherText: encryptedData,
		Nonce:      nonce,
	}

	payloadBytes, err := json.Marshal(cachePayload)
	if err != nil {
		return []byte{}, err
	}

	return payloadBytes, err
}

// DecacheUser decrypts and reconstructs a user entity from a cached payload.
// It unmarshals the payload, decrypts the contained ciphertext using the utils' Decrypt method,
// and reconstructs the original user entity from the decrypted data.
//
// Parameters:
//   - payloadBytes []byte: JSON-encoded UserCachePayload containing encrypted user data
//
// Returns:
//   - *ent.User: The reconstructed user entity
//   - error: Any error that occurred during unmarshalling or decryption
func (s *CacheStore) DecacheUser(payloadBytes []byte) (*ent.User, error) {
	var payload models.CachePayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return nil, err
	}

	plaintext, err := s.Decrypt(payload.CipherText, payload.Nonce)
	if err != nil {
		return nil, err
	}

	var user ent.User
	if err := json.Unmarshal(plaintext, &user); err != nil {
		return nil, err
	}

	return &user, nil
}

// Encrypt encrypts a phrase using ChaCha20-Poly1305 with the utils' cipher key.
// The encrypted data and nonce are returned as base64-encoded strings.
// It uses the X variant of ChaCha20-Poly1305 which uses a 24-byte nonce.
//
// Parameters:
//   - phrase: []byte The plaintext data to encrypt
//
// Returns:
//   - string: Base64-encoded ciphertext
//   - string: Base64-encoded nonce
//   - error: Any error that occurred during encryption
func (s *CacheStore) Encrypt(phrase []byte) (string, string, error) {
	aead, err := chacha20poly1305.NewX(s.CipherKey)
	if err != nil {
		return "", "", err
	}

	nonce, err := s.GenerateNonce()
	if err != nil {
		return "", "", err
	}

	ciphertext := aead.Seal(nil, nonce, phrase, nil)
	return base64.StdEncoding.EncodeToString(ciphertext), base64.StdEncoding.EncodeToString(nonce), nil
}

// Decrypt decrypts a base64-encoded ciphertext using ChaCha20-Poly1305 with the utils' cipher key.
// It uses the X variant of ChaCha20-Poly1305 and requires the corresponding nonce.
// The function returns the decrypted plaintext or an error if decryption fails.
//
// Parameters:
//   - cipherPhrase: string Base64-encoded encrypted data
//   - nonceBase64: string Base64-encoded nonce used during encryption
//
// Returns:
//   - []byte: The decrypted plaintext
//   - error: - Any error that occurred during decryption (invalid key, malformed input, or authentication failure)
func (s *CacheStore) Decrypt(cipherPhrase string, nonceBase64 string) ([]byte, error) {
	aead, err := chacha20poly1305.NewX(s.CipherKey)
	if err != nil {
		return nil, err
	}

	nonce, err := base64.StdEncoding.DecodeString(nonceBase64)
	if err != nil {
		return nil, err
	}

	ciphertext, err := base64.StdEncoding.DecodeString(cipherPhrase)
	if err != nil {
		return nil, err
	}

	plaintext, err := aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// GenerateNonce creates a random nonce for use in ChaCha20-Poly1305 encryption.
// The nonce is generated with the size required for the X variant of the algorithm.
//
// Returns:
//   - []byte: The generated nonce
//   - error: Any error that occurred during random number generation
func (s *CacheStore) GenerateNonce() ([]byte, error) {
	nonce := make([]byte, chacha20poly1305.NonceSizeX)

	_, err := rand.Read(nonce)
	if err != nil {
		return nil, err
	}

	return nonce, nil
}
