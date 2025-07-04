package services

import (
	"context"
	"crypto"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"
	"users-service/app/infrastructure"
	"users-service/app/models/ent"
	requests "users-service/app/models/requests"
	models "users-service/app/models/utils"
	"users-service/app/repositories"
	"users-service/app/utils"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog"
	"golang.org/x/crypto/argon2"
)

type AuthService struct {
	UsersRepository repositories.UsersRepositoryInterface
	EventNotifier   infrastructure.EventNotifierInterface
	Logger          zerolog.Logger
	PublicKey       crypto.PublicKey
	PrivateKey      crypto.PrivateKey
	HashingPepper   []byte
	TokenDuration   time.Duration
	HashPool        chan func()
}

type AuthServiceInterface interface {
	GenerateJWT(ctx context.Context, userID int, deviceID int, userRoles []models.UserRoleInfo) (string, error)
	ValidateJWT(ctx context.Context, tokenString string) (*models.Claims, error)
	HashPassword(password string) (string, error)
	CompareHashes(storedHash, password string) error

	Register(ctx context.Context, req *requests.Register) (*ent.User, error)
	Login(ctx context.Context, req *requests.Login) (*ent.User, string, error)

	ActivateAccount(ctx context.Context, req *requests.Code) (*ent.User, error)
	VerifyAccount(ctx context.Context, req *requests.Code) (*ent.User, string, error)
	VerifySecondFactor(ctx context.Context, req *requests.Code) (*ent.User, string, error)

	RequestPasswordReset(ctx context.Context, req *requests.Mail) error
	CancelPasswordReset(ctx context.Context, req *requests.Mail) error
	ConfirmPasswordReset(ctx context.Context, req *requests.ConfirmPasswordReset) error

	ResendActivationCode(ctx context.Context, req *requests.Mail) error
	ResendVerificationCode(ctx context.Context, req *requests.Mail) error
	ResendResetCode(ctx context.Context, req *requests.Mail) error
	ResendSecondFactorCode(ctx context.Context, req *requests.Mail) error
}

var _ AuthServiceInterface = (*AuthService)(nil)

func NewAuthService(
	usersRepository repositories.UsersRepositoryInterface,
	eventNotifier infrastructure.EventNotifierInterface,
	logger zerolog.Logger,
	publicKey crypto.PublicKey,
	privateKey crypto.PrivateKey,
	poolSize int) *AuthService {

	authService := &AuthService{
		UsersRepository: usersRepository,
		EventNotifier:   eventNotifier,
		Logger:          logger,
		PublicKey:       publicKey,
		PrivateKey:      privateKey,
		HashPool:        make(chan func(), poolSize),
	}

	for i := 0; i < poolSize; i++ {
		go func() {
			for task := range authService.HashPool {
				task()
			}
		}()
	}

	return authService
}

// GenerateJWT creates a signed JWT token for a user with specified roles.
// It sets the user ID, roles, issuance time, and expiration based on TokenDuration.
//
// Parameters:
//   - userID: The user’s ID.
//   - deviceID: The device’s ID.
//   - roles: A slice of role names for the user.
//
// Returns:
//   - A string containing the signed JWT token, or empty if an error occurs.
//   - An error if signing fails, or nil on success.
func (s *AuthService) GenerateJWT(ctx context.Context, userID int, deviceID int, userRoles []models.UserRoleInfo) (string, error) {
	s.Logger.Info().Int("userID", userID).Int("deviceID", deviceID).Msg("Generating JWT for device")

	claims := models.Claims{
		UserID:    userID,
		DeviceID:  deviceID,
		UserRoles: userRoles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(s.TokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		},
	}

	token := jwt.NewWithClaims(&jwt.SigningMethodEd25519{}, claims)
	resultChan := make(chan models.JWTResult, 1)

	go func() {
		signedToken, err := token.SignedString(s.PrivateKey)
		if err != nil {
			s.Logger.Error().Int("userID", userID).Err(err).Msg("Failed to sign JWT")
			resultChan <- models.JWTResult{Token: "", Err: err}
			return
		}
		s.Logger.Info().Int("userID", userID).Msg("JWT successfully generated")
		resultChan <- models.JWTResult{Token: signedToken, Err: nil}
	}()

	select {
	case result := <-resultChan:
		if result.Err != nil {
			return "", result.Err
		}
		return result.Token, nil
	case <-ctx.Done():
		s.Logger.Warn().Int("userID", userID).Err(ctx.Err()).Msg("JWT generation timed out")
		return "", errors.New("JWT generation timed out")
	}
}

func (s *AuthService) ValidateJWT(ctx context.Context, tokenString string) (*models.Claims, error) {
	s.Logger.Info().Msg("Validating JWT")

	resultChan := make(chan models.JWTValidationResult, 1)

	go func() {
		token, err := jwt.ParseWithClaims(tokenString, &models.Claims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodEd25519); !ok {
				s.Logger.Warn().Str("alg", fmt.Sprintf("%v", token.Header["alg"])).Msg("Unexpected signing method")
				return nil, errors.New("unexpected signing method")
			}
			return s.PublicKey, nil
		})

		if err != nil {
			s.Logger.Error().Err(err).Msg("JWT parsing failed")
			resultChan <- models.JWTValidationResult{Claims: nil, Err: err}
			return
		}

		claims, ok := token.Claims.(*models.Claims)
		if !ok || !token.Valid {
			s.Logger.Warn().Msg("Invalid token or claims mismatch")
			resultChan <- models.JWTValidationResult{Claims: nil, Err: errors.New("invalid token or claims")}
			return
		}

		if claims.DeviceID == 0 {
			s.Logger.Warn().Int("userID", claims.UserID).Msg("Token is missing device ID")
			resultChan <- models.JWTValidationResult{Claims: nil, Err: errors.New("token is not device-specific")}
			return
		}

		if claims.ExpiresAt != nil && claims.ExpiresAt.Unix() <= time.Now().UTC().Unix() {
			s.Logger.Warn().
				Int("userID", claims.UserID).
				Int("deviceID", claims.DeviceID).
				Time("exp", claims.ExpiresAt.Time).
				Msg("Token expired")
			resultChan <- models.JWTValidationResult{Claims: nil, Err: errors.New("token has expired")}
			return
		}

		s.Logger.Info().
			Int("userID", claims.UserID).
			Int("deviceID", claims.DeviceID).
			Interface("userRoles", claims.UserRoles).
			Msg("Token validated successfully")
		resultChan <- models.JWTValidationResult{Claims: claims, Err: nil}
	}()

	select {
	case result := <-resultChan:
		if result.Err != nil {
			return nil, result.Err
		}
		return result.Claims, nil
	case <-ctx.Done():
		s.Logger.Warn().Err(ctx.Err()).Msg("JWT validation timed out")
		return nil, errors.New("JWT validation timed out")
	}
}

func (s *AuthService) Register(
	ctx context.Context,
	req *requests.Register) (*ent.User, error) {

	if user, _ := s.UsersRepository.GetUserByMail(ctx, req.Mail); user != nil {
		return nil, errors.New("this email is already registered")
	}

	hashedPassword, err := s.HashPassword(req.Password)
	if err != nil {
		return nil, errors.New("failed to hash user's password")
	}

	newUser, activationCode, err := s.UsersRepository.CreateUser(ctx, req, hashedPassword)
	if err != nil {
		return nil, errors.New("failed to create new user")
	}

	if err := s.EventNotifier.CreateRegistrationEvent(newUser, activationCode); err != nil {
		s.Logger.Error().Err(err).Int("userID", newUser.ID).Msg("Failed to produce registration event")
	}

	return newUser, nil
}

func (s *AuthService) ResendActivationCode(ctx context.Context, req *requests.Mail) error {
	user, err := s.UsersRepository.GetUserByMail(ctx, req.Mail)
	if err != nil {
		return errors.New("failed to find user with the provided email")
	}

	if err := s.EventNotifier.CreateRegistrationEvent(user, user.Edges.ActivationCode); err != nil {
		s.Logger.Error().Err(err).Int("userID", user.ID).Msg("Failed to produce resend-activation event")
		return errors.New("failed to send activation code")
	}

	s.Logger.Info().Int("userID", user.ID).Msg("Resent activation code event produced")
	return nil
}

func (s *AuthService) ActivateAccount(
	ctx context.Context,
	req *requests.Code) (*ent.User, error) {

	user, err := s.UsersRepository.ActivateAccount(ctx, req.Code)
	if err != nil {
		return nil, errors.New("provided code doesn't exist")
	}

	action := &ent.UserAction{Action: "account.activated"}
	if err := s.EventNotifier.CreateUserActionEvent(user.Edges.UserSettings, action); err != nil {
		s.Logger.Error().Err(err).Int("userID", user.ID).Msg("Failed to produce account activation event")
	}

	return user, nil
}

func (s *AuthService) HashPassword(password string) (string, error) {
	resultChan := make(chan struct {
		Hash string
		Err  error
	}, 1)

	s.HashPool <- func() {
		salt := make([]byte, 16)
		if _, err := rand.Read(salt); err != nil {
			s.Logger.Error().Err(err).Msg("Failed to generate salt for password hashing")
			resultChan <- struct {
				Hash string
				Err  error
			}{"", err}
			return
		}

		h := hmac.New(sha256.New, s.HashingPepper)
		h.Write([]byte(password))
		hmacValue := h.Sum(nil)

		hashed := argon2.IDKey(hmacValue, salt, 3, 128*1024, 4, 32)
		hashedStr := fmt.Sprintf("$argon2id$v=19$m=131072,t=3,p=4$%s$%s",
			base64.RawStdEncoding.EncodeToString(salt),
			base64.RawStdEncoding.EncodeToString(hashed))

		resultChan <- struct {
			Hash string
			Err  error
		}{hashedStr, nil}
	}

	result := <-resultChan
	return result.Hash, result.Err
}

func (s *AuthService) CompareHashes(storedHash, password string) error {
	parts := strings.Split(storedHash, "$")
	if len(parts) != 6 {
		return errors.New("invalid stored hash format")
	}

	var memory, hashTime uint32
	var threads uint8
	_, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &hashTime, &threads)
	if err != nil {
		return fmt.Errorf("failed to parse hash parameters: %w", err)
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return fmt.Errorf("failed to decode salt from hash: %w", err)
	}

	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return fmt.Errorf("failed to decode hash part: %w", err)
	}

	resultChan := make(chan error, 1)

	s.HashPool <- func() {
		h := hmac.New(sha256.New, s.HashingPepper)
		h.Write([]byte(password))
		hmacValue := h.Sum(nil)

		comparisonHash := argon2.IDKey(hmacValue, salt, hashTime, memory, threads, uint32(len(hash)))

		if subtle.ConstantTimeCompare(hash, comparisonHash) == 1 {
			resultChan <- nil
		} else {
			resultChan <- errors.New("invalid credentials")
		}
	}

	return <-resultChan
}

func (s *AuthService) Login(ctx context.Context, req *requests.Login) (*ent.User, string, error) {
	user, err := s.UsersRepository.GetUserByMail(ctx, req.Mail)
	if err != nil {
		return nil, "", err
	}

	if user.Blacklisted {
		return nil, "", errors.New("account is blacklisted")
	}

	if !user.Active {
		return nil, "", errors.New("account is not activated")
	}

	if err := s.CompareHashes(user.Password, req.Password); err != nil {
		return nil, "", errors.New("provided wrong password")
	}

	device, err := s.UsersRepository.FindOrCreateDevice(ctx, user.ID, &req.Device)
	if err != nil {
		s.Logger.Error().Err(err).Int("userID", user.ID).Msg("Failed to find or create device")
		return nil, "", errors.New("failed to handle user device")
	}

	if user.Edges.UserSettings.TwoFactor {
		userWithCode, secondFactor, err := s.UsersRepository.CreateSecondFactorCode(ctx, user, device)
		if err != nil {
			return nil, "", errors.New("failed to create second factor code")
		}

		if err := s.EventNotifier.CreateSecondFactorEvent(userWithCode.Edges.UserSettings, secondFactor); err != nil {
			s.Logger.Error().Err(err).Int("userID", user.ID).Msg("Failed to produce second factor event")
			return nil, "", errors.New("failed to send second factor code")
		}

		return nil, "2fa_required", nil
	}

	userRoles := utils.MarshalUserRoles(user.Edges.UserRoles)
	token, err := s.GenerateJWT(ctx, user.ID, device.ID, userRoles)
	if err != nil {
		return nil, "", err
	}

	if err := s.EventNotifier.CreateLoginEvent(user.Edges.UserSettings, "password"); err != nil {
		s.Logger.Error().Err(err).Int("userID", user.ID).Msg("Failed to produce login event")
	}

	return user, token, nil
}

func (s *AuthService) ResendSecondFactorCode(ctx context.Context, req *requests.Mail) error {
	user, err := s.UsersRepository.GetUserByMail(ctx, req.Mail)
	if err != nil {
		return errors.New("failed to find user with the provided email")
	}

	if !user.Edges.UserSettings.TwoFactor {
		return errors.New("second factor is not enabled for this account")
	}

	if user.Edges.SecondFactorCode == nil {
		return errors.New("user has no login attempts created")
	}

	if err := s.EventNotifier.CreateSecondFactorEvent(user.Edges.UserSettings, user.Edges.SecondFactorCode); err != nil {
		s.Logger.Error().Err(err).Int("userID", user.ID).Msg("Failed to produce resend-second-factor event")
		return errors.New("failed to send second factor code")
	}

	s.Logger.Info().Int("userID", user.ID).Msg("Resent second factor event produced")
	return nil
}

func (s *AuthService) VerifySecondFactor(ctx context.Context, req *requests.Code) (*ent.User, string, error) {
	user, err := s.UsersRepository.GetUserBySecondFactor(ctx, req.Code)
	if err != nil {
		return nil, "", errors.New("failed to fetch user by second factor code")
	}

	if time.Now().UTC().After(user.Edges.SecondFactorCode.ExpiresAt) {
		go s.UsersRepository.RemoveSecondFactorCode(context.Background(), user)
		return nil, "", errors.New("second factor token has expired")
	}

	device := user.Edges.SecondFactorCode.Edges.TargetUserDevice
	if device == nil {
		s.Logger.Error().Int("userID", user.ID).Msg("Device context not found for second factor verification")
		return nil, "", errors.New("could not determine the device for this session")
	}

	userRoles := utils.MarshalUserRoles(user.Edges.UserRoles)
	token, err := s.GenerateJWT(ctx, user.ID, device.ID, userRoles)
	if err != nil {
		return nil, "", errors.New("failed to create Bearer for user")
	}

	user, err = s.UsersRepository.RemoveSecondFactorCode(ctx, user)
	if err != nil {
		s.Logger.Error().Err(err).Int("userID", user.ID).Msg("Failed to remove second factor code after use")
	}

	if err := s.EventNotifier.CreateLoginEvent(user.Edges.UserSettings, "second-factor"); err != nil {
		s.Logger.Error().Err(err).Int("userID", user.ID).Msg("Failed to produce login event after 2FA")
	}

	return user, token, nil
}

func (s *AuthService) ResendVerificationCode(ctx context.Context, req *requests.Mail) error {
	user, err := s.UsersRepository.GetUserByMail(ctx, req.Mail)
	if err != nil {
		return errors.New("failed to find user with the provided email")
	}

	if err := s.EventNotifier.CreateVerificationEvent(user.Edges.UserSettings, user.Phone, user.Edges.VerificationCode); err != nil {
		s.Logger.Error().Err(err).Int("userID", user.ID).Msg("Failed to produce resend-verification event")
		return errors.New("failed to send verification code")
	}

	s.Logger.Info().Int("userID", user.ID).Msg("Resent verification code event produced")
	return nil
}

func (s *AuthService) VerifyAccount(ctx context.Context, req *requests.Code) (*ent.User, string, error) {
	user, err := s.UsersRepository.VerifyAccount(ctx, req.Code)
	if err != nil {
		return nil, "", errors.New("failed to fetch user by verification code")
	}

	device, err := s.UsersRepository.FindOrCreateDevice(ctx, user.ID, &req.Device)
	if err != nil {
		s.Logger.Error().Err(err).Int("userID", user.ID).Msg("Failed to find or create device during account verification")
		return nil, "", errors.New("failed to handle user device")
	}

	userRoles := utils.MarshalUserRoles(user.Edges.UserRoles)
	token, err := s.GenerateJWT(ctx, user.ID, device.ID, userRoles)
	if err != nil {
		return nil, "", errors.New("failed to create Bearer for user")
	}

	return user, token, nil
}

func (s *AuthService) RequestPasswordReset(ctx context.Context, req *requests.Mail) error {
	user, err := s.UsersRepository.GetUserByMail(ctx, req.Mail)
	if err != nil {
		return errors.New("failed to fetch user")
	}

	resetCode, err := s.UsersRepository.CreateResetCode(ctx, user)
	if err != nil {
		return errors.New("failed to create reset token")
	}

	if err := s.EventNotifier.CreateResetPasswordEvent(user.Edges.UserSettings, resetCode); err != nil {
		s.Logger.Error().Err(err).Int("userID", user.ID).Msg("Failed to produce reset password event")
		return errors.New("failed to send reset password code")
	}

	return nil
}

func (s *AuthService) CancelPasswordReset(ctx context.Context, req *requests.Mail) error {
	user, err := s.UsersRepository.GetUserByMail(ctx, req.Mail)
	if err != nil {
		return errors.New("failed to fetch user")
	}

	if _, err = s.UsersRepository.RemoveResetCode(ctx, user); err != nil {
		return errors.New("failed to cancel password reset process")
	}

	action := &ent.UserAction{Action: "password.reset.cancelled"}
	if err := s.EventNotifier.CreateUserActionEvent(user.Edges.UserSettings, action); err != nil {
		s.Logger.Error().Err(err).Int("userID", user.ID).Msg("Failed to produce password reset cancellation event")
	}

	return nil
}

func (s *AuthService) ConfirmPasswordReset(ctx context.Context, req *requests.ConfirmPasswordReset) error {
	user, err := s.UsersRepository.GetUserByResetCode(ctx, req.Code)
	if err != nil {
		return errors.New("failed to fetch user by reset code")
	}

	hashedPassword, err := s.HashPassword(req.Password)
	if err != nil {
		return errors.New("failed to hash user's password")
	}

	if _, err = s.UsersRepository.UpdateUsersPassword(ctx, user, hashedPassword); err != nil {
		return errors.New("failed to update users password")
	}

	if _, err = s.UsersRepository.RemoveResetCode(ctx, user); err != nil {
		return errors.New("failed to remove reset code")
	}

	action := &ent.UserAction{Action: "password.reset.confirmed"}
	if err := s.EventNotifier.CreateUserActionEvent(user.Edges.UserSettings, action); err != nil {
		s.Logger.Error().Err(err).Int("userID", user.ID).Msg("Failed to produce password reset confirmation event")
	}

	return nil
}

func (s *AuthService) ResendResetCode(ctx context.Context, req *requests.Mail) error {
	user, err := s.UsersRepository.GetUserByMail(ctx, req.Mail)
	if err != nil {
		return errors.New("failed to find user with the provided email")
	}

	resetCode, err := s.UsersRepository.CreateResetCode(ctx, user)
	if err != nil {
		return errors.New("failed to create a new reset code")
	}

	if err := s.EventNotifier.CreateResetPasswordEvent(user.Edges.UserSettings, resetCode); err != nil {
		s.Logger.Error().Err(err).Int("userID", user.ID).Msg("Failed to produce resend-reset-code event")
		return errors.New("failed to send reset code")
	}

	s.Logger.Info().Int("userID", user.ID).Msg("Resent reset code event produced")
	return nil
}
