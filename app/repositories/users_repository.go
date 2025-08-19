package repositories

import (
	"context"
	"errors"
	"strconv"
	"time"
	"users-service/app/infrastructure"
	"users-service/app/models/ent"
	"users-service/app/models/ent/activationcode"
	"users-service/app/models/ent/entrepreneurdetails"
	"users-service/app/models/ent/location"
	_ "users-service/app/models/ent/runtime"
	"users-service/app/models/ent/secondfactorcode"
	"users-service/app/models/ent/user"
	"users-service/app/models/ent/userdetails"
	"users-service/app/models/ent/userdevice"
	"users-service/app/models/ent/verificationcode"
	"users-service/app/models/requests"
	"users-service/app/utils"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type UsersRepository struct {
	CacheStore infrastructure.CacheStoreInterface
	DB         *ent.Client
	Driver     *sql.Driver
	Logger     zerolog.Logger
}

type UsersRepositoryInterface interface {
	CreateUser(ctx context.Context, req *requests.Register, hashedPassword string) (*ent.User, *ent.ActivationCode, error)
	CreateResetCode(ctx context.Context, user *ent.User) (*ent.ResetCode, error)
	CreateSecondFactorCode(ctx context.Context, user *ent.User, device *ent.UserDevice) (*ent.User, *ent.SecondFactorCode, error)
	CreateUserAction(ctx context.Context, user *ent.User, action, originDevice, details string) error

	GetUserByID(ctx context.Context, ID int) (*ent.User, error)
	GetUserPublicByID(ctx context.Context, ID int) (*ent.User, error)
	GetUserByMail(ctx context.Context, mail string) (*ent.User, error)
	GetUserByPhone(ctx context.Context, phone string) (*ent.User, error)
	GetUserBySecondFactor(ctx context.Context, code string) (*ent.User, error)
	GetUserByVerificationCode(ctx context.Context, code string) (*ent.User, error)
	GetUserByResetCode(ctx context.Context, code string) (*ent.User, error)
	GetUsersPublic(ctx context.Context, page, pageSize int) ([]*ent.User, error)

	FindOrCreateDevice(ctx context.Context, userID int, req *requests.Device) (*ent.UserDevice, error)

	ActivateAccount(ctx context.Context, code string) (*ent.User, error)
	AddSubscriptions(ctx context.Context, user *ent.User, subIDs []int) error
	VerifyAccount(ctx context.Context, code string) (*ent.User, error)
	UpdateUser(ctx context.Context, user *ent.User, req *requests.User) (*ent.User, error)
	UpdateUsersPassword(ctx context.Context, user *ent.User, hashedPassword string) (*ent.User, error)
	UpdateUsersDetails(ctx context.Context, user *ent.User, req *requests.Details) (*ent.User, error)
	UpdateUsersSettings(ctx context.Context, user *ent.User, req *requests.Settings) (*ent.User, error)
	UpdateEntrepreneurDetails(ctx context.Context, user *ent.User, req *requests.EntrepreneurDetails) (*ent.User, error)
	UpdateLocation(ctx context.Context, user *ent.User, req *requests.Location) (*ent.User, error)

	RemoveResetCode(ctx context.Context, user *ent.User) (*ent.User, error)
	RemoveSecondFactorCode(ctx context.Context, user *ent.User) (*ent.User, error)
	RemoveAccount(ctx context.Context, user *ent.User) error
	RemoveSubscriptions(ctx context.Context, user *ent.User, subIDs []int) error

	Ping() error
}

var _ UsersRepositoryInterface = (*UsersRepository)(nil)

func NewUsersRepository(
	cacheStore infrastructure.CacheStoreInterface,
	entClient *ent.Client,
	driver *sql.Driver,
	logger zerolog.Logger) *UsersRepository {

	return &UsersRepository{
		CacheStore: cacheStore,
		DB:         entClient,
		Driver:     driver,
		Logger:     logger,
	}
}

func (r *UsersRepository) GetUserByID(ctx context.Context, ID int) (*ent.User, error) {
	r.Logger.Info().Int("userID", ID).Msg("Attempting to get user by ID")
	var foundUser *ent.User
	var err error

	payload, _ := r.CacheStore.Get(ctx, "user:"+strconv.Itoa(ID))
	if payload != nil {
		return r.CacheStore.DecacheUser(payload)
	}

	if err = WithTransaction(ctx, r.DB, func(tx *ent.Tx) error {
		foundUser, err = GetUserByID(ctx, tx, ID)
		if err != nil {
			r.Logger.Error().Err(err).Int("userID", ID).Msg("Failed to get user by ID within transaction")
			return err
		}

		r.Logger.Debug().Int("userID", ID).Msg("User found by ID within transaction")
		return nil
	}); err != nil {
		r.Logger.Error().Err(err).Int("userID", ID).Msg("Transaction failed for GetUserByID")
		return nil, err
	}

	if err := UpdateCache(ctx, foundUser, r); err != nil {
		return nil, errors.New("failed to update user")
	}

	r.Logger.Info().Int("userID", foundUser.ID).Msg("Successfully retrieved user by ID")
	return foundUser, nil
}

func (r *UsersRepository) GetUserPublicByID(ctx context.Context, ID int) (*ent.User, error) {
	r.Logger.Info().Int("userID", ID).Msg("Attempting to get user by ID")
	var foundUser *ent.User
	var err error

	payload, _ := r.CacheStore.Get(ctx, "user-public:"+strconv.Itoa(ID))
	if payload != nil {
		return r.CacheStore.DecacheUser(payload)
	}

	if err = WithTransaction(ctx, r.DB, func(tx *ent.Tx) error {
		foundUser, err = GetUserPublicByID(ctx, tx, ID)
		if err != nil {
			r.Logger.Error().Err(err).Int("userID", ID).Msg("Failed to get user by ID within transaction")
			return err
		}

		r.Logger.Debug().Int("userID", ID).Msg("User found by ID within transaction")
		return nil
	}); err != nil {
		r.Logger.Error().Err(err).Int("userID", ID).Msg("Transaction failed for GetUserByID")
		return nil, err
	}

	if err := UpdateCache(ctx, foundUser, r); err != nil {
		return nil, errors.New("failed to update user")
	}

	r.Logger.Info().Int("userID", foundUser.ID).Msg("Successfully retrieved user by ID")
	return foundUser, nil
}

func (r *UsersRepository) GetUserByMail(ctx context.Context, mail string) (*ent.User, error) {
	r.Logger.Info().Str("mail", mail).Msg("Attempting to get user by mail")
	var foundUser *ent.User
	var err error

	payload, _ := r.CacheStore.Get(ctx, "user:"+mail)
	if payload != nil {
		return r.CacheStore.DecacheUser(payload)
	}

	if err := WithTransaction(ctx, r.DB, func(tx *ent.Tx) error {
		foundUser, err = GetUserByMail(ctx, tx, mail)
		if err != nil {
			r.Logger.Error().Err(err).Str("mail", mail).Msg("Failed to get user by mail within transaction")
			return err
		}
		r.Logger.Debug().Str("mail", mail).Msg("User found by mail within transaction")
		return nil
	}); err != nil {
		r.Logger.Error().Err(err).Str("mail", mail).Msg("Transaction failed for GetUserByMail")
		return nil, err
	}

	if err := UpdateCache(ctx, foundUser, r); err != nil {
		return nil, errors.New("failed to update user")
	}

	r.Logger.Info().Str("mail", foundUser.Mail).Msg("Successfully retrieved user by mail")
	return foundUser, nil
}

func (r *UsersRepository) ActivateAccount(ctx context.Context, code string) (*ent.User, error) {
	r.Logger.Info().Str("activationCode", code).Msg("Attempting to activate account")
	var foundUser *ent.User
	var err error

	if err := WithTransaction(ctx, r.DB, func(tx *ent.Tx) error {
		foundUser, err = GetUserByActivationCode(ctx, tx, code)
		if err != nil {
			r.Logger.Error().Err(err).Msg("Failed to get user by activation code within transaction")
			return err
		}
		r.Logger.Debug().Int("userID", foundUser.ID).Msg("User found for activation")

		if _, err := foundUser.
			Update().
			SetActive(true).
			Save(ctx); err != nil {
			r.Logger.Error().Err(err).Int("userID", foundUser.ID).Msg("Failed to set user active")
			return err
		}
		r.Logger.Debug().Int("userID", foundUser.ID).Msg("User account set to active")

		if _, err := tx.ActivationCode.
			Delete().
			Where(activationcode.CodeEQ(code)).
			Exec(ctx); err != nil {
			r.Logger.Error().Msg("Failed to delete activation code")
			return err
		}
		r.Logger.Debug().Str("activationCode", code).Msg("Activation code deleted")

		foundUser, err = GetUserByID(ctx, tx, foundUser.ID)
		if err != nil {
			r.Logger.Error().Err(err).Int("userID", foundUser.ID).Msg("Failed to re-fetch user after activation")
			return err
		}
		r.Logger.Debug().Int("userID", foundUser.ID).Msg("User re-fetched after activation")

		return nil
	}); err != nil {
		r.Logger.Error().Err(err).Str("activationCode", code).Msg("Transaction failed for ActivateAccount")
		return nil, err
	}

	if err := UpdateCache(ctx, foundUser, r); err != nil {

	}

	r.Logger.Info().Int("userID", foundUser.ID).Msg("Account successfully activated")
	return foundUser, nil
}

func (r *UsersRepository) VerifyAccount(ctx context.Context, code string) (*ent.User, error) {
	r.Logger.Info().Str("verificationCode", code).Msg("Attempting to verify account")
	var foundUser *ent.User
	var err error

	if err := WithTransaction(ctx, r.DB, func(tx *ent.Tx) error {
		foundUser, err = GetUserByVerificationCode(ctx, tx, code)
		if err != nil {
			r.Logger.Error().Err(err).Str("verificationCode", code).Msg("Failed to get user by verification code within transaction")
			return err
		}
		r.Logger.Debug().Int("userID", foundUser.ID).Msg("User found for verification")

		if _, err := tx.User.
			UpdateOneID(foundUser.ID).
			SetVerified(true).
			Save(ctx); err != nil {
			r.Logger.Error().Err(err).Int("userID", foundUser.ID).Msg("Failed to set user verified")
			return err
		}
		r.Logger.Debug().Int("userID", foundUser.ID).Msg("User account set to verified")

		if _, err := tx.VerificationCode.
			Delete().
			Where(verificationcode.CodeEQ(code)).
			Exec(ctx); err != nil {
			r.Logger.Error().Err(err).Str("verificationCode", code).Msg("Failed to delete verification code")
			return err
		}
		r.Logger.Debug().Str("verificationCode", code).Msg("Verification code deleted")

		foundUser, err = GetUserByID(ctx, tx, foundUser.ID)
		if err != nil {
			r.Logger.Error().Err(err).Int("userID", foundUser.ID).Msg("Failed to re-fetch user after verification")
			return err
		}
		r.Logger.Debug().Int("userID", foundUser.ID).Msg("User re-fetched after verification")

		return nil
	}); err != nil {
		r.Logger.Error().Err(err).Str("verificationCode", code).Msg("Transaction failed for VerifyAccount")
		return nil, err
	}

	if err := UpdateCache(ctx, foundUser, r); err != nil {
		return nil, err
	}

	r.Logger.Info().Int("userID", foundUser.ID).Msg("Account successfully verified")
	return foundUser, nil
}

func (r *UsersRepository) GetUserBySecondFactor(ctx context.Context, code string) (*ent.User, error) {
	r.Logger.Info().Str("secondFactorCode", code).Msg("Attempting to get user by second factor code")
	var foundUser *ent.User
	var err error

	if err := WithTransaction(ctx, r.DB, func(tx *ent.Tx) error {
		foundUser, err = GetUserBySecondFactor(ctx, tx, code)
		if err != nil {
			r.Logger.Error().Err(err).Str("secondFactorCode", code).Msg("Failed to get user by second factor code within transaction")
			return err
		}
		r.Logger.Debug().Int("userID", foundUser.ID).Msg("User found by second factor code within transaction")
		return nil
	}); err != nil {
		r.Logger.Error().Err(err).Str("secondFactorCode", code).Msg("Transaction failed for GetUserBySecondFactor")
		return nil, err
	}

	if err := UpdateCache(ctx, foundUser, r); err != nil {
		return nil, err
	}

	r.Logger.Info().Int("userID", foundUser.ID).Msg("Successfully retrieved user by second factor code")
	return foundUser, nil
}

func (r *UsersRepository) GetUserByVerificationCode(ctx context.Context, code string) (*ent.User, error) {
	r.Logger.Info().Str("verificationCode", code).Msg("Attempting to get user by verification code")
	var foundUser *ent.User
	var err error

	if err := WithTransaction(ctx, r.DB, func(tx *ent.Tx) error {
		foundUser, err = GetUserByVerificationCode(ctx, tx, code)
		if err != nil {
			r.Logger.Error().Err(err).Str("verificationCode", code).Msg("Failed to get user by verification code within transaction")
			return err
		}
		r.Logger.Debug().Int("userID", foundUser.ID).Msg("User found by verification code within transaction")
		return nil
	}); err != nil {
		r.Logger.Error().Err(err).Str("verificationCode", code).Msg("Transaction failed for GetUserByVerificationCode")
		return nil, err
	}

	if err := UpdateCache(ctx, foundUser, r); err != nil {
		return nil, err
	}

	r.Logger.Info().Int("userID", foundUser.ID).Msg("Successfully retrieved user by verification code")
	return foundUser, nil
}

func (r *UsersRepository) GetUserByResetCode(ctx context.Context, code string) (*ent.User, error) {
	r.Logger.Info().Str("resetCode", code).Msg("Attempting to get user by reset code")
	var foundUser *ent.User
	var err error

	if err := WithTransaction(ctx, r.DB, func(tx *ent.Tx) error {
		foundUser, err = GetUserByResetCode(ctx, tx, code)
		if err != nil {
			r.Logger.Error().Err(err).Str("resetCode", code).Msg("Failed to get user by reset code within transaction")
			return err
		}
		r.Logger.Debug().Int("userID", foundUser.ID).Msg("User found by reset code within transaction")
		return nil
	}); err != nil {
		r.Logger.Error().Err(err).Str("resetCode", code).Msg("Transaction failed for GetUserByResetCode")
		return nil, err
	}

	if err := UpdateCache(ctx, foundUser, r); err != nil {
		return nil, err
	}

	r.Logger.Info().Int("userID", foundUser.ID).Msg("Successfully retrieved user by reset code")
	return foundUser, nil
}

func (r *UsersRepository) GetUserByPhone(ctx context.Context, phone string) (*ent.User, error) {
	r.Logger.Info().Str("phone", phone).Msg("Attempting to get user by phone")
	var foundUser *ent.User
	var err error

	if err := WithTransaction(ctx, r.DB, func(tx *ent.Tx) error {
		foundUser, err = GetUserByPhone(ctx, tx, phone)
		if err != nil {
			r.Logger.Error().Err(err).Str("phone", phone).Msg("Failed to get user by phone within transaction")
			return err
		}
		r.Logger.Debug().Str("phone", phone).Msg("User found by phone within transaction")
		return nil
	}); err != nil {
		r.Logger.Error().Err(err).Str("phone", phone).Msg("Transaction failed for GetUserByPhone")
		return nil, err
	}

	if err := UpdateCache(ctx, foundUser, r); err != nil {
		return nil, err
	}

	r.Logger.Info().Str("phone", foundUser.Phone).Msg("Successfully retrieved user by phone")
	return foundUser, nil
}

func (r *UsersRepository) CreateUser(ctx context.Context, req *requests.Register, hashedPassword string) (*ent.User, *ent.ActivationCode, error) {
	r.Logger.Info().Str("mail", req.Mail).Msg("Attempting to create new user")
	var newUser *ent.User
	var activationCode *ent.ActivationCode
	var err error

	if err := WithTransaction(ctx, r.DB, func(tx *ent.Tx) error {
		tempUser, err := tx.User.
			Create().
			SetMail(req.Mail).
			SetPassword(hashedPassword).
			Save(ctx)
		if err != nil {
			r.Logger.Error().Err(err).Str("mail", req.Mail).Msg("Failed to create user")
			return err
		}
		r.Logger.Debug().Int("userID", tempUser.ID).Msg("User created successfully")

		if _, err := tx.UserSettings.
			Create().
			SetUUID(uuid.New().String()).
			SetOwner(tempUser).
			Save(ctx); err != nil {
			r.Logger.Error().Err(err).Int("userID", tempUser.ID).Msg("Failed to create user settings")
			return errors.New("failed to create user settings")
		}
		r.Logger.Debug().Int("userID", tempUser.ID).Msg("User settings created")

		if _, err := tx.UserDetails.
			Create().
			SetName(req.Name).
			SetOwner(tempUser).
			Save(ctx); err != nil {
			r.Logger.Error().Err(err).Int("userID", tempUser.ID).Msg("Failed to create user details")
			return errors.New("failed to create user details")
		}
		r.Logger.Debug().Int("userID", tempUser.ID).Msg("User details created")

		if _, err := tx.UserDevice.
			Create().
			SetOwner(tempUser).
			SetToken(req.DeviceToken).
			SetIPAddress(req.IPAddress).
			SetUserAgent(req.UserAgent).
			SetOsName(req.OSName).
			SetOsVersion(req.OSVersion).
			SetBrowserName(req.BrowserName).
			SetBrowserVersion(req.BrowserVersion).
			Save(ctx); err != nil {
			r.Logger.Error().Err(err).Int("userID", tempUser.ID).Msg("Failed to create user device")
			return errors.New("failed to create user device")
		}
		r.Logger.Debug().Int("userID", tempUser.ID).Str("deviceToken", req.DeviceToken).Msg("User device created")

		var codeStr string
		for {
			code := utils.GenerateRandomCode()
			codeStr = strconv.Itoa(code)

			if exists, _ := tx.ActivationCode.
				Query().
				Where(activationcode.CodeEQ(codeStr)).
				Exist(ctx); !exists {
				r.Logger.Debug().Str("activationCode", codeStr).Msg("Generated unique activation code")
				break
			}
			r.Logger.Debug().Str("activationCode", codeStr).Msg("Generated activation code already exists, retrying")
		}

		activationCode, err = tx.ActivationCode.
			Create().
			SetCode(codeStr).
			SetOwner(tempUser).
			Save(ctx)
		if err != nil {
			r.Logger.Error().Err(err).Int("userID", tempUser.ID).Msg("Failed to create activation code")
			return err
		}
		r.Logger.Debug().Int("activationCodeID", activationCode.ID).Msg("Activation code created")

		newUser, err = GetUserByID(ctx, tx, tempUser.ID)
		if err != nil {
			return errors.New("failed to fetch user")
		}

		return nil
	}); err != nil {
		r.Logger.Error().Err(err).Str("mail", req.Mail).Msg("Transaction failed for CreateUser")
		return nil, nil, err
	}

	if err = UpdateCache(ctx, newUser, r); err != nil {
		return nil, nil, errors.New("failed to update user")
	}

	r.Logger.Info().Int("userID", newUser.ID).Msg("User and associated entities created successfully")
	return newUser, activationCode, nil
}

func (r *UsersRepository) CreateSecondFactorCode(ctx context.Context, user *ent.User, device *ent.UserDevice) (*ent.User, *ent.SecondFactorCode, error) {
	r.Logger.Info().Int("userID", user.ID).Int("deviceID", device.ID).Msg("Attempting to create second factor code")
	var updatedUser *ent.User
	var secondFactor *ent.SecondFactorCode
	var err error

	if err := WithTransaction(ctx, r.DB, func(tx *ent.Tx) error {
		var codeStr string
		for {
			code := utils.GenerateRandomCode()
			codeStr = strconv.Itoa(code)

			exists, err := tx.SecondFactorCode.
				Query().
				Where(secondfactorcode.CodeEQ(codeStr)).
				Exist(ctx)
			if err != nil {
				r.Logger.Error().Err(err).Msg("Failed to check existence of second factor code")
				return err
			}
			if !exists {
				r.Logger.Debug().Str("secondFactorCode", codeStr).Msg("Generated unique second factor code")
				break
			}
			r.Logger.Debug().Str("secondFactorCode", codeStr).Msg("Generated second factor code already exists, retrying")
		}

		secondFactor, err = tx.SecondFactorCode.
			Create().
			SetCode(codeStr).
			SetOwner(user).
			SetTargetUserDeviceID(device.ID).
			Save(ctx)
		if err != nil {
			r.Logger.Error().Err(err).Int("userID", user.ID).Msg("Failed to create second factor code")
			return err
		}
		r.Logger.Debug().Int("secondFactorID", secondFactor.ID).Msg("Second factor code created")

		updatedUser, err = GetUserByID(ctx, tx, user.ID)
		if err != nil {
			r.Logger.Error().Err(err).Int("userID", user.ID).Msg("Failed to re-fetch user after second factor code creation")
			return err
		}
		r.Logger.Debug().Int("userID", updatedUser.ID).Msg("User re-fetched after second factor code creation")

		return nil
	}); err != nil {
		r.Logger.Error().Err(err).Int("userID", user.ID).Msg("Transaction failed for CreateSecondFactorCode")
		return nil, nil, err
	}

	if err := UpdateCache(ctx, updatedUser, r); err != nil {
		return nil, nil, errors.New("failed to update user")
	}

	r.Logger.Info().Int("userID", updatedUser.ID).Msg("Second factor code created successfully")
	return updatedUser, secondFactor, nil
}

func (r *UsersRepository) CreateResetCode(ctx context.Context, user *ent.User) (*ent.ResetCode, error) {
	r.Logger.Info().Int("userID", user.ID).Msg("Attempting to create reset code")
	var resetCode *ent.ResetCode
	var updatedUser *ent.User
	var err error

	if err := WithTransaction(ctx, r.DB, func(tx *ent.Tx) error {
		var codeStr string

		for {
			code := utils.GenerateRandomCode()
			codeStr = strconv.Itoa(code)

			if exists, _ := tx.SecondFactorCode.
				Query().
				Where(secondfactorcode.CodeEQ(codeStr)).
				Exist(ctx); !exists {
				r.Logger.Debug().Str("resetCode", codeStr).Msg("Generated unique reset code")
				break
			}
			r.Logger.Debug().Str("resetCode", codeStr).Msg("Generated reset code already exists, retrying")
		}

		resetCode, err = tx.ResetCode.
			Create().
			SetCode(codeStr).
			SetOwner(user).
			Save(ctx)
		if err != nil {
			r.Logger.Error().Err(err).Int("userID", user.ID).Msg("Failed to create reset code")
			return err
		}
		r.Logger.Debug().Int("resetCodeID", resetCode.ID).Msg("Reset code created")

		updatedUser, err = GetUserByID(ctx, tx, user.ID)
		if err != nil {
			r.Logger.Error().Err(err).Int("userID", user.ID).Msg("Failed to re-fetch user after reset code creation")
			return err
		}

		r.Logger.Debug().Int("userID", user.ID).Msg("User re-fetched after reset code creation (for consistency check)")

		return nil
	}); err != nil {
		r.Logger.Error().Err(err).Int("userID", user.ID).Msg("Transaction failed for CreateResetCode")
		return nil, err
	}

	if err := UpdateCache(ctx, updatedUser, r); err != nil {
		return nil, errors.New("failed to update user")
	}

	r.Logger.Info().Int("userID", user.ID).Msg("Reset code created successfully")
	return resetCode, nil
}

func (r *UsersRepository) RemoveResetCode(ctx context.Context, user *ent.User) (*ent.User, error) {
	r.Logger.Info().Int("userID", user.ID).Msg("Attempting to remove reset code")
	var updatedUser *ent.User
	var err error

	if err := WithTransaction(ctx, r.DB, func(tx *ent.Tx) error {
		if user.Edges.ResetCode == nil {
			r.Logger.Warn().Int("userID", user.ID).Msg("User has no reset code to remove")
			return errors.New("user has no reset code")
		}
		r.Logger.Debug().Int("resetCodeID", user.Edges.ResetCode.ID).Msg("Reset code found for deletion")

		if err := tx.ResetCode.
			DeleteOneID(user.Edges.ResetCode.ID).
			Exec(ctx); err != nil {
			r.Logger.Error().Err(err).Int("userID", user.ID).Int("resetCodeID", user.Edges.ResetCode.ID).Msg("Failed to delete reset code")
			return err
		}
		r.Logger.Debug().Int("resetCodeID", user.Edges.ResetCode.ID).Msg("Reset code deleted")

		updatedUser, err = GetUserByID(ctx, tx, user.ID)
		if err != nil {
			r.Logger.Error().Err(err).Int("userID", user.ID).Msg("Failed to re-fetch user after reset code removal")
			return err
		}
		r.Logger.Debug().Int("userID", updatedUser.ID).Msg("User re-fetched after reset code removal")

		return nil
	}); err != nil {
		r.Logger.Error().Err(err).Int("userID", user.ID).Msg("Transaction failed for RemoveResetCode")
		return nil, err
	}

	if err := UpdateCache(ctx, updatedUser, r); err != nil {
		return nil, errors.New("failed to update user")
	}

	r.Logger.Info().Int("userID", updatedUser.ID).Msg("Reset code removed successfully")
	return updatedUser, nil
}

func (r *UsersRepository) RemoveSecondFactorCode(ctx context.Context, user *ent.User) (*ent.User, error) {
	r.Logger.Info().Int("userID", user.ID).Msg("Attempting to remove second factor code")
	var updatedUser *ent.User
	var err error

	if err := WithTransaction(ctx, r.DB, func(tx *ent.Tx) error {
		if user.Edges.SecondFactorCode == nil {
			r.Logger.Warn().Int("userID", user.ID).Msg("User has no second factor code to remove")
			return errors.New("user has no reset code")
		}

		r.Logger.Debug().Int("secondFactorCodeID", user.Edges.SecondFactorCode.ID).Msg("Second factor code found for deletion")

		if err := tx.SecondFactorCode.
			DeleteOneID(user.Edges.SecondFactorCode.ID).
			Exec(ctx); err != nil {
			r.Logger.Error().Err(err).Int("userID", user.ID).Int("secondFactorCodeID", user.Edges.SecondFactorCode.ID).Msg("Failed to remove second factor code")
			return errors.New("failed to remove second factor code")
		}
		r.Logger.Debug().Int("secondFactorCodeID", user.Edges.SecondFactorCode.ID).Msg("Second factor code deleted")

		updatedUser, err = GetUserByID(ctx, tx, user.ID)
		if err != nil {
			r.Logger.Error().Err(err).Int("userID", user.ID).Msg("Failed to re-fetch user after second factor code removal")
			return err
		}
		r.Logger.Debug().Int("userID", updatedUser.ID).Msg("User re-fetched after second factor code removal")

		return nil
	}); err != nil {
		r.Logger.Error().Err(err).Int("userID", user.ID).Msg("Transaction failed for RemoveSecondFactorCode")
		return nil, err
	}

	r.Logger.Info().Int("userID", updatedUser.ID).Msg("Second factor code removed successfully. Caching user.")
	cachedUser, err := r.CacheStore.CacheUser(updatedUser)
	if err != nil {
		r.Logger.Error().Err(err).Int("userID", updatedUser.ID).Msg("Failed to cache user after second factor code removal")
	} else {
		cacheKey := "user:" + strconv.Itoa(updatedUser.ID)
		r.CacheStore.Set(ctx, cacheKey, cachedUser, time.Minute*5)
		r.Logger.Debug().Str("cacheKey", cacheKey).Msg("User cached successfully")
	}

	if err := UpdateCache(ctx, updatedUser, r); err != nil {
		return nil, errors.New("failed to update user")
	}

	return updatedUser, nil
}

func (r *UsersRepository) UpdateUsersPassword(ctx context.Context, user *ent.User, hashedPassword string) (*ent.User, error) {
	r.Logger.Info().Int("userID", user.ID).Msg("Attempting to update user password")
	var updatedUser *ent.User
	var err error

	if err := WithTransaction(ctx, r.DB, func(tx *ent.Tx) error {
		if _, err = tx.User.
			UpdateOneID(user.ID).
			SetPassword(hashedPassword).
			Save(ctx); err != nil {
			r.Logger.Error().Err(err).Int("userID", user.ID).Msg("Failed to update user password in DB")
			return err
		}
		r.Logger.Debug().Int("userID", user.ID).Msg("User password updated in DB")

		updatedUser, err = GetUserByID(ctx, tx, user.ID)
		if err != nil {
			r.Logger.Error().Err(err).Int("userID", user.ID).Msg("Failed to re-fetch user after password update")
			return err
		}
		r.Logger.Debug().Int("userID", updatedUser.ID).Msg("User re-fetched after password update")

		return nil
	}); err != nil {
		r.Logger.Error().Err(err).Int("userID", user.ID).Msg("Transaction failed for UpdateUsersPassword")
		return nil, err
	}

	if err := UpdateCache(ctx, updatedUser, r); err != nil {
		return nil, errors.New("failed to update user")
	}

	r.Logger.Info().Int("userID", updatedUser.ID).Msg("User password updated successfully")
	return updatedUser, nil
}

func (r *UsersRepository) UpdateUser(ctx context.Context, user *ent.User, req *requests.User) (*ent.User, error) {
	r.Logger.Info().Int("userID", user.ID).Msg("Attempting to update user mail/phone")
	var updatedUser *ent.User
	var err error

	if err := WithTransaction(ctx, r.DB, func(tx *ent.Tx) error {
		userUpdater := tx.User.UpdateOneID(user.ID)
		updated := false

		if req.Phone != user.Phone {
			userUpdater.SetPhone(req.Phone)
			updated = true
			r.Logger.Debug().Int("userID", user.ID).Str("oldPhone", user.Phone).Str("newPhone", req.Phone).Msg("Updating user phone")
		}

		if req.Mail != user.Mail {
			userUpdater.SetMail(req.Mail)
			updated = true
			r.Logger.Debug().Int("userID", user.ID).Str("oldMail", user.Mail).Str("newMail", req.Mail).Msg("Updating user mail")
		}

		if updated {
			if _, err := userUpdater.Save(ctx); err != nil {
				r.Logger.Error().Err(err).Int("userID", user.ID).Msg("Failed to update user mail/phone in DB")
				return errors.New("failed to update user's mail/phone")
			}
			r.Logger.Debug().Int("userID", user.ID).Msg("User mail/phone updated in DB")
		} else {
			r.Logger.Debug().Int("userID", user.ID).Msg("No changes detected for user mail/phone, skipping update")
		}

		updatedUser, err = GetUserByID(ctx, tx, user.ID)
		if err != nil {
			r.Logger.Error().Err(err).Int("userID", user.ID).Msg("Failed to re-fetch user after mail/phone update")
			return err
		}
		r.Logger.Debug().Int("userID", updatedUser.ID).Msg("User re-fetched after mail/phone update")

		return nil
	}); err != nil {
		r.Logger.Error().Err(err).Int("userID", user.ID).Msg("Transaction failed for UpdateUser")
		return nil, errors.New("failed to update user")
	}

	if err := UpdateCache(ctx, updatedUser, r); err != nil {
		return nil, errors.New("failed to update user")
	}

	r.Logger.Info().Int("userID", updatedUser.ID).Msg("User mail/phone updated successfully")
	return updatedUser, nil
}

func (r *UsersRepository) UpdateUsersDetails(ctx context.Context, user *ent.User, req *requests.Details) (*ent.User, error) {
	r.Logger.Debug().Int("userID", user.ID).Msg("Attempting to update user details")
	var updatedUser *ent.User
	var err error

	if err := WithTransaction(ctx, r.DB, func(tx *ent.Tx) error {
		detailsUpdated := tx.UserDetails.
			UpdateOneID(user.Edges.UserDetails.ID).
			SetFirstName(req.FirstName).
			SetLastName(req.LastName)

		if _, err = detailsUpdated.Save(ctx); err != nil {
			r.Logger.Error().Err(err).Int("userID", user.ID).Msg("Failed to update user's details in DB")
			return errors.New("failed to update user's details")
		}
		r.Logger.Debug().Int("userID", user.ID).Msg("User details updated in DB")

		updatedUser, err = GetUserByID(ctx, tx, user.ID)
		if err != nil {
			r.Logger.Error().Err(err).Int("userID", user.ID).Msg("Failed to re-fetch user after details update")
			return err
		}

		r.Logger.Debug().Int("userID", updatedUser.ID).Msg("User re-fetched after details update")

		return nil
	}); err != nil {
		r.Logger.Error().Err(err).Int("userID", user.ID).Msg("Transaction failed for UpdateUsersDetails")
		return nil, errors.New("failed to update user")
	}

	if err := UpdateCache(ctx, updatedUser, r); err != nil {
		return nil, errors.New("failed to update user")
	}

	r.Logger.Debug().Int("userID", updatedUser.ID).Msg("User details updated successfully")
	return updatedUser, nil
}

func (r *UsersRepository) UpdateUsersSettings(ctx context.Context, user *ent.User, req *requests.Settings) (*ent.User, error) {
	r.Logger.Debug().Int("userID", user.ID).Msg("Attempting to update user settings")
	var updatedUser *ent.User
	var err error

	if err := WithTransaction(ctx, r.DB, func(tx *ent.Tx) error {
		settingsUpdater := tx.UserSettings.
			UpdateOneID(user.Edges.UserSettings.ID).
			SetNightMode(req.NightMode).
			SetTwoFactor(req.TwoFactor).
			SetSecondFactorTargetID(req.SecondFactorTargetID)

		r.Logger.Debug().Int("userID", user.ID).
			Bool("nightMode", req.NightMode).
			Bool("twoFactor", req.TwoFactor).
			Int("secondFactorTargetID", req.SecondFactorTargetID).
			Msg("Updating user settings fields")

		settingsUpdater.ClearNotificationTargetDevices().
			AddNotificationTargetDeviceIDs(req.NotificationsTargetDeviceIDs...)
		r.Logger.Debug().Int("userID", user.ID).
			Ints("notificationTargetDeviceIDs", req.NotificationsTargetDeviceIDs).
			Msg("Updating notification target devices")

		if _, err = settingsUpdater.Save(ctx); err != nil {
			r.Logger.Error().Err(err).Int("userID", user.ID).Msg("Failed to update user's settings in DB")
			return errors.New("failed to update user's settings")
		}
		r.Logger.Debug().Int("userID", user.ID).Msg("User settings updated in DB")

		updatedUser, err = GetUserByID(ctx, tx, user.ID)
		if err != nil {
			r.Logger.Error().Err(err).Int("userID", user.ID).Msg("Failed to re-fetch user after settings update")
			return err
		}
		r.Logger.Debug().Int("userID", updatedUser.ID).Msg("User re-fetched after settings update")

		return nil
	}); err != nil {
		r.Logger.Error().Err(err).Int("userID", user.ID).Msg("Transaction failed for UpdateUsersSettings")
		return nil, errors.New("failed to update user")
	}

	if err := UpdateCache(ctx, updatedUser, r); err != nil {
		return nil, errors.New("failed to update user")
	}

	r.Logger.Debug().Int("userID", updatedUser.ID).Msg("User settings updated successfully")
	return updatedUser, nil
}

func (r *UsersRepository) RemoveAccount(
	ctx context.Context,
	user *ent.User,
) error {
	r.Logger.Info().Int("userID", user.ID).Msg("Attempting to soft delete user account")
	var err error

	if err = WithTransaction(ctx, r.DB, func(tx *ent.Tx) error {
		if _, err := tx.User.
			UpdateOneID(user.ID).
			SetRemoved(true).
			Save(ctx); err != nil {
			r.Logger.Error().Err(err).Int("userID", user.ID).Msg("Failed to mark user as removed in DB")
			err = errors.New("failed to remove user")
			return err
		}
		r.Logger.Debug().Int("userID", user.ID).Msg("User account marked as removed in DB")

		return nil
	}); err != nil {
		r.Logger.Error().Err(err).Int("userID", user.ID).Msg("Transaction failed for RemoveAccount")
		return errors.New("failed to remove user")
	}

	keys := utils.GetUserKeys(user)
	for _, key := range keys {
		if err := r.CacheStore.Del(ctx, key); err != nil {
			return errors.New("failed to revoke cache")
		}
	}

	r.Logger.Debug().Int("userID", user.ID).Msg("User account successfully marked as removed")
	return err
}

func (r *UsersRepository) FindOrCreateDevice(
	ctx context.Context,
	userID int,
	req *requests.Device,
) (*ent.UserDevice, error) {

	r.Logger.Debug().Int("userID", userID).Str("deviceToken", req.DeviceToken).Msg("Attempting to find or create device")
	var device *ent.UserDevice

	if err := WithTransaction(ctx, r.DB, func(tx *ent.Tx) error {
		d, err := tx.UserDevice.Query().
			Where(
				userdevice.HasOwnerWith(user.ID(userID)),
				userdevice.Token(req.DeviceToken),
			).
			Only(ctx)

		if err != nil {
			if ent.IsNotFound(err) {
				r.Logger.Debug().Int("userID", userID).Str("deviceToken", req.DeviceToken).Msg("Device not found, attempting to create new device")
				newDevice, createErr := tx.UserDevice.
					Create().
					SetOwnerID(userID).
					SetToken(req.DeviceToken).
					SetIPAddress(req.IPAddress).
					SetBrowserName(req.BrowserName).
					SetBrowserVersion(req.BrowserVersion).
					SetOsName(req.OSName).
					SetOsVersion(req.OSVersion).
					SetUserAgent(req.UserAgent).
					Save(ctx)
				if createErr != nil {
					r.Logger.Error().Err(createErr).Int("userID", userID).Str("deviceToken", req.DeviceToken).Msg("Failed to create new device")
					err = createErr
					return createErr
				}
				device = newDevice
				r.Logger.Debug().Int("userID", userID).Int("deviceID", device.ID).Msg("New device created successfully")

				foundUser, err := GetUserByID(ctx, tx, userID)
				if err != nil {
					return errors.New("failed to fetch user")
				}

				if err := UpdateCache(ctx, foundUser, r); err != nil {
					return errors.New("failed to update user")
				}

				return nil
			}
			r.Logger.Error().Err(err).Int("userID", userID).Str("deviceToken", req.DeviceToken).Msg("Failed to query device in DB")
			return err
		}

		device = d
		r.Logger.Debug().Int("userID", userID).Int("deviceID", device.ID).Msg("Existing device found")
		return nil
	}); err != nil {
		r.Logger.Error().Err(err).Int("userID", userID).Str("deviceToken", req.DeviceToken).Msg("Transaction failed for FindOrCreateDevice")
		return nil, errors.New("failed to find or create device in transaction")
	}

	r.Logger.Debug().Int("userID", userID).Int("deviceID", device.ID).Msg("Device operation completed successfully")
	return device, nil
}

func (r *UsersRepository) UpdateEntrepreneurDetails(ctx context.Context, user *ent.User, req *requests.EntrepreneurDetails) (*ent.User, error) {
	r.Logger.Info().Int("userID", user.ID).Msg("Attempting to update entrepreneur details")
	var updatedUser *ent.User
	var err error

	if err = WithTransaction(ctx, r.DB, func(tx *ent.Tx) error {
		details, _ := tx.EntrepreneurDetails.
			Query().
			Where(entrepreneurdetails.HasUserDetailsWith(userdetails.IDEQ(user.Edges.UserDetails.ID))).
			Only(ctx)
		if details == nil {
			if _, err := tx.EntrepreneurDetails.
				Create().
				SetUserDetailsID(updatedUser.Edges.UserDetails.ID).
				SetBusinessName(req.BusinessName).
				SetNip(req.NIP).
				SetKrs(req.KRS).
				SetDescription(req.Description).
				SetIncome(req.Income).
				SetCosts(req.Costs).
				SetFundingCapital(req.FundingCapital).
				SetManagementCouncilMembers(req.ManagementCouncilMembers).
				SetDecisionMakers(req.DecisionMakers).
				SetBusinessPhoneNumber(req.BusinessPhoneNumber).
				SetBusinessMail(req.BusinessMail).
				SetWebsiteAddress(req.WebsiteAddress).
				Save(ctx); err != nil {
				return errors.New("failed to create entrepreneur details")
			}
		} else {
			if _, err := tx.EntrepreneurDetails.
				UpdateOneID(details.ID).
				SetUserDetailsID(updatedUser.Edges.UserDetails.ID).
				SetBusinessName(req.BusinessName).
				SetNip(req.NIP).
				SetKrs(req.KRS).
				SetDescription(req.Description).
				SetIncome(req.Income).
				SetCosts(req.Costs).
				SetFundingCapital(req.FundingCapital).
				SetManagementCouncilMembers(req.ManagementCouncilMembers).
				SetDecisionMakers(req.DecisionMakers).
				SetBusinessPhoneNumber(req.BusinessPhoneNumber).
				SetBusinessMail(req.BusinessMail).
				SetWebsiteAddress(req.WebsiteAddress).
				Save(ctx); err != nil {
				return errors.New("failed to update entrepreneur details")
			}
		}

		updatedUser, err = GetUserByID(ctx, tx, user.ID)
		if err != nil {
			return errors.New("failed to fetch updated user")
		}

		r.Logger.Debug().Int("userID", user.ID).Msg("User account marked as removed in DB")

		return nil
	}); err != nil {
		r.Logger.Error().Err(err).Int("userID", user.ID).Msg("Transaction failed for RemoveAccount")
		return nil, errors.New("failed to remove user")
	}

	if err := UpdateCache(ctx, updatedUser, r); err != nil {
		return nil, errors.New("failed to update user")
	}

	return updatedUser, nil
}

func (r *UsersRepository) UpdateLocation(ctx context.Context, user *ent.User, req *requests.Location) (*ent.User, error) {
	r.Logger.Debug().Int("userID", user.ID).Msg("Attempting to update user location")
	var updatedUser *ent.User
	var err error

	if err = WithTransaction(ctx, r.DB, func(tx *ent.Tx) error {
		foundLocation, _ := tx.Location.
			Query().
			Where(location.HasUserDetailsWith(userdetails.IDEQ(updatedUser.Edges.UserDetails.ID))).
			Only(ctx)
		if foundLocation != nil {
			if _, err := tx.Location.
				Update().
				Save(ctx); err != nil {
				return errors.New("failed to update user's location")
			}

		} else {
			if _, err := tx.Location.
				Create().
				Save(ctx); err != nil {
				return errors.New("failed to create new location")
			}
		}

		updatedUser, err = GetUserByID(ctx, tx, user.ID)
		if err != nil {
			return errors.New("failed to fetch updated user")
		}

		r.Logger.Debug().Int("userID", user.ID).Msg("User account marked as removed in DB")

		return nil
	}); err != nil {
		r.Logger.Error().Err(err).Int("userID", user.ID).Msg("Transaction failed for RemoveAccount")
		return nil, errors.New("failed to remove user")
	}

	if err := UpdateCache(ctx, updatedUser, r); err != nil {
		return nil, errors.New("failed to update user")
	}

	r.Logger.Debug().Int("userID", user.ID).Msg("User account successfully marked as removed")

	return nil, nil
}

func (r *UsersRepository) AddSubscriptions(ctx context.Context, user *ent.User, subIDs []int) error {
	r.Logger.Debug().Int("userID", user.ID).Msg("Attempting to add user's subscriptions")
	var updatedUser *ent.User
	var err error

	if err = WithTransaction(ctx, r.DB, func(tx *ent.Tx) error {
		toAppend := make([]int, 0)
		for _, newID := range subIDs {
			if !utils.ContainsInt(user.SubscriptionIds, newID) {
				toAppend = append(toAppend, newID)
			}
		}

		if len(toAppend) == 0 {
			return nil
		}

		newSubIDs := append(user.SubscriptionIds, toAppend...)

		if _, err := tx.User.
			UpdateOneID(user.ID).
			SetSubscriptionIds(newSubIDs).
			Save(ctx); err != nil {
			return errors.New("failed update user")
		}

		updatedUser, err = GetUserByID(ctx, tx, user.ID)
		if err != nil {
			return errors.New("failed to fetch updated user")
		}

		return nil

	}); err != nil {
		r.Logger.Error().Err(err).Int("userID", user.ID).Msg("Transaction failed for RemoveAccount")
		return errors.New("failed to remove user")
	}

	if err := UpdateCache(ctx, updatedUser, r); err != nil {
		return errors.New("failed to update user")
	}

	r.Logger.Debug().Int("userID", user.ID).Msg("User account successfully marked as removed")

	return nil
}

func (r *UsersRepository) RemoveSubscriptions(ctx context.Context, user *ent.User, subIDs []int) error {
	r.Logger.Debug().Int("userID", user.ID).Msg("Attempting to revoke user's subscriptions")
	var updatedUser *ent.User
	var err error

	if err = WithTransaction(ctx, r.DB, func(tx *ent.Tx) error {
		newSubIDs := utils.RemoveIDs(user.SubscriptionIds, subIDs)

		if _, err := tx.User.
			UpdateOneID(user.ID).
			SetSubscriptionIds(newSubIDs).
			Save(ctx); err != nil {
			return errors.New("failed update user")
		}

		updatedUser, err = GetUserByID(ctx, tx, user.ID)
		if err != nil {
			return errors.New("failed to fetch updated user")
		}

		return nil
	}); err != nil {
		r.Logger.Error().Err(err).Int("userID", user.ID).Msg("Transaction failed for RemoveAccount")
		return errors.New("failed to remove user")
	}

	if err := UpdateCache(ctx, updatedUser, r); err != nil {
		return errors.New("failed to update user")
	}

	r.Logger.Debug().Int("userID", user.ID).Msg("User account successfully marked as removed")

	return nil
}

func (r *UsersRepository) CreateUserAction(
	ctx context.Context,
	user *ent.User,
	action, originDevice, details string,
) error {
	r.Logger.Debug().Int("userID", user.ID).Msg("Attempting to revoke user's subscriptions")
	var updatedUser *ent.User
	var err error

	if err = WithTransaction(ctx, r.DB, func(tx *ent.Tx) error {
		if _, err := tx.UserAction.
			Create().
			SetAction(action).
			SetDetails(details).
			AddAuthorIDs(user.ID).
			Save(ctx); err != nil {
			return errors.New("failed to create user action")
		}

		updatedUser, err = GetUserByID(ctx, tx, user.ID)
		if err != nil {
			return errors.New("failed to fetch updated user")
		}

		return nil
	}); err != nil {
		r.Logger.Error().Err(err).Int("userID", user.ID).Msg("Transaction failed for RemoveAccount")
		return errors.New("failed to remove user")
	}

	if err := UpdateCache(ctx, updatedUser, r); err != nil {
		return errors.New("failed to update user")
	}

	r.Logger.Debug().Int("userID", user.ID).Msg("User account successfully marked as removed")
	return nil
}

func (r *UsersRepository) GetUsersPublic(ctx context.Context, page, pageSize int) ([]*ent.User, error) {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	r.Logger.Debug().Int("page", page).Int("pageSize", pageSize).Msg("Fetching public users with offset")

	foundUsers, err := r.DB.User.
		Query().
		Limit(pageSize).
		Offset(offset).
		Order(ent.Desc(user.FieldCreatedAt)).
		All(ctx)

	if err != nil {
		r.Logger.Error().Err(err).Msg("Database query failed for GetUsersPublic")
		return nil, errors.New("failed to fetch users public")
	}

	for _, user := range foundUsers {
		if err := UpdateCache(ctx, user, r); err != nil {
			r.Logger.Warn().Err(err).Int("userID", user.ID).Msg("Failed to update cache for a user, continuing")
		}
	}

	return foundUsers, nil
}

func (r *UsersRepository) Ping() error {
	r.Logger.Debug().Msg("Attempting to ping database")

	err := r.Driver.DB().Ping()
	if err != nil {
		r.Logger.Error().Err(err).Msg("Failed to ping database")
		return err
	}

	r.Logger.Debug().Msg("Database ping successful")
	return nil
}
