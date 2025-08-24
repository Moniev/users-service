package repositories

import (
	"context"
	"errors"
	"strconv"
	"users-service/app/infrastructure"
	"users-service/app/models/ent"
	"users-service/app/models/ent/activationcode"
	"users-service/app/models/ent/entrepreneurdetails"
	"users-service/app/models/ent/location"
	"users-service/app/models/ent/resetcode"
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

var ()

type UsersRepository struct {
	CacheStore infrastructure.CacheStoreInterface
	DB         *ent.Client
	Driver     *sql.Driver
	Logger     zerolog.Logger
}

type UsersRepositoryInterface interface {
	AddRole(ctx context.Context, userID, roleID int) error
	RevokeRole(ctx context.Context, userID, roleID int) error

	CreateUser(ctx context.Context, req *requests.Register, hashedPassword string) (*ent.User, *ent.ActivationCode, error)
	CreateResetCode(ctx context.Context, user *ent.User) (*ent.ResetCode, error)
	CreateSecondFactorCode(ctx context.Context, user *ent.User, device *ent.UserDevice) (*ent.SecondFactorCode, error)
	CreateUserAction(ctx context.Context, user *ent.User, action, originDevice, details string) error

	GetUserByID(ctx context.Context, ID int) (*ent.User, error)
	GetUserFunctionalByID(ctx context.Context, ID int) (*ent.User, error)
	GetUserFunctionalDetailsByID(ctx context.Context, ID int) (*ent.User, error)
	GetUserPublicByID(ctx context.Context, ID int) (*ent.User, error)
	GetUserByMail(ctx context.Context, mail string) (*ent.User, error)
	GetUserByMailWithCodes(ctx context.Context, mail string) (*ent.User, error)
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
	r.Logger.Info().Int("user_id", ID).Msg("Attempting to get user by ID")
	var foundUser *ent.User
	var err error

	if err = WithTransaction(ctx, r.DB, func(tx *ent.Tx) error {
		foundUser, err = GetUserByID(ctx, tx, ID)
		if err != nil {
			r.Logger.Error().Err(err).Int("user_id", ID).Msg("Failed to get user by ID within transaction")
			return err
		}

		r.Logger.Debug().Int("user_id", ID).Msg("User found by ID within transaction")
		return nil
	}); err != nil {
		r.Logger.Error().Err(err).Int("user_id", ID).Msg("Transaction failed for GetUserByID")
		return nil, err
	}

	r.Logger.Info().Int("user_id", foundUser.ID).Msg("Successfully retrieved user by ID")
	return foundUser, nil
}

func (r *UsersRepository) GetUserFunctionalDetailsByID(ctx context.Context, ID int) (*ent.User, error) {
	r.Logger.Info().Int("user_id", ID).Msg("Attempting to get user by ID")
	var foundUser *ent.User
	var err error

	if err = WithTransaction(ctx, r.DB, func(tx *ent.Tx) error {
		foundUser, err = GetUserFunctionalByID(ctx, tx, ID)
		if err != nil {
			r.Logger.Error().Err(err).Int("user_id", ID).Msg("Failed to get user by ID within transaction")
			return err
		}

		r.Logger.Debug().Int("user_id", ID).Msg("User found by ID within transaction")
		return nil
	}); err != nil {
		r.Logger.Error().Err(err).Int("user_id", ID).Msg("Transaction failed for GetUserByID")
		return nil, err
	}

	r.Logger.Info().Int("user_id", foundUser.ID).Msg("Successfully retrieved user by ID")
	return foundUser, nil
}

func (r *UsersRepository) GetUserFunctionalByID(ctx context.Context, ID int) (*ent.User, error) {
	r.Logger.Info().Int("user_id", ID).Msg("Attempting to get user by ID")
	var foundUser *ent.User
	var err error

	if err = WithTransaction(ctx, r.DB, func(tx *ent.Tx) error {
		foundUser, err = GetUserFunctionalByID(ctx, tx, ID)
		if err != nil {
			r.Logger.Error().Err(err).Int("user_id", ID).Msg("Failed to get user by ID within transaction")
			return err
		}

		r.Logger.Debug().Int("user_id", ID).Msg("User found by ID within transaction")
		return nil
	}); err != nil {
		r.Logger.Error().Err(err).Int("user_id", ID).Msg("Transaction failed for GetUserByID")
		return nil, err
	}

	r.Logger.Info().Int("user_id", foundUser.ID).Msg("Successfully retrieved user by ID")
	return foundUser, nil
}

func (r *UsersRepository) GetUserPublicByID(ctx context.Context, ID int) (*ent.User, error) {
	r.Logger.Info().Int("user_id", ID).Msg("Attempting to get user by ID")
	var foundUser *ent.User
	var err error

	if err = WithTransaction(ctx, r.DB, func(tx *ent.Tx) error {
		foundUser, err = GetUserPublicByID(ctx, tx, ID)
		if err != nil {
			r.Logger.Error().Err(err).Int("user_id", ID).Msg("Failed to get user by ID within transaction")
			return err
		}

		r.Logger.Debug().Int("user_id", ID).Msg("User found by ID within transaction")
		return nil
	}); err != nil {
		r.Logger.Error().Err(err).Int("user_id", ID).Msg("Transaction failed for GetUserByID")
		return nil, err
	}

	r.Logger.Info().Int("user_id", foundUser.ID).Msg("Successfully retrieved user by ID")
	return foundUser, nil
}

func (r *UsersRepository) GetUserByMail(ctx context.Context, mail string) (*ent.User, error) {
	r.Logger.Info().Str("mail", mail).Msg("Attempting to get user by mail")
	var foundUser *ent.User
	var err error

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

	r.Logger.Info().Str("mail", foundUser.Mail).Msg("Successfully retrieved user by mail")
	return foundUser, nil
}

func (r *UsersRepository) GetUserByMailWithCodes(ctx context.Context, mail string) (*ent.User, error) {
	r.Logger.Info().Str("mail", mail).Msg("Attempting to get user by mail")
	var foundUser *ent.User
	var err error

	payload, _ := r.CacheStore.Get(ctx, "user:"+mail)
	if payload != nil {
		return r.CacheStore.DecacheUser(payload)
	}

	if err := WithTransaction(ctx, r.DB, func(tx *ent.Tx) error {
		foundUser, err = GetUserByMailWithCodes(ctx, tx, mail)
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

	r.Logger.Info().Str("mail", foundUser.Mail).Msg("Successfully retrieved user by mail")
	return foundUser, nil
}

func (r *UsersRepository) ActivateAccount(ctx context.Context, code string) (*ent.User, error) {
	r.Logger.Info().Str("activation_code", code).Msg("Attempting to activate account")
	var updatedUser *ent.User

	if err := WithTransaction(ctx, r.DB, func(tx *ent.Tx) error {
		foundUser, err := GetUserByActivationCode(ctx, tx, code)
		if err != nil {
			r.Logger.Error().Err(err).Msg("Failed to get user by activation code within transaction")
			return err
		}
		r.Logger.Debug().Int("user_id", foundUser.ID).Msg("User found for activation")

		if _, err = foundUser.
			Update().
			SetActive(true).
			Save(ctx); err != nil {
			r.Logger.Error().Err(err).Int("user_id", foundUser.ID).Msg("Failed to set user active")
			return err
		}
		r.Logger.Debug().Int("user_id", foundUser.ID).Msg("User account set to active")

		if _, err := tx.ActivationCode.
			Delete().
			Where(activationcode.CodeEQ(code)).
			Exec(ctx); err != nil {
			r.Logger.Error().Msg("Failed to delete activation code")
			return err
		}
		r.Logger.Debug().Str("activation_code", code).Msg("Activation code deleted")

		updatedUser, err = getUserWithSettings(ctx, tx, user.IDEQ(foundUser.ID))
		if err != nil {
			return errors.New("failed to find user's settings")
		}

		return nil
	}); err != nil {
		r.Logger.Error().Err(err).Str("activation_code", code).Msg("Transaction failed for ActivateAccount")
		return nil, err
	}

	InvalidateCache(ctx, updatedUser, r)

	r.Logger.Info().Int("user_id", updatedUser.ID).Msg("Account successfully activated")
	return updatedUser, nil
}

func (r *UsersRepository) VerifyAccount(ctx context.Context, code string) (*ent.User, error) {
	r.Logger.Info().Str("verification_code", code).Msg("Attempting to verify account")
	var foundUser *ent.User
	var err error

	if err := WithTransaction(ctx, r.DB, func(tx *ent.Tx) error {
		foundUser, err = GetUserByVerificationCode(ctx, tx, code)
		if err != nil {
			r.Logger.Error().Err(err).Str("verification_code", code).Msg("Failed to get user by verification code within transaction")
			return err
		}
		r.Logger.Debug().Int("user_id", foundUser.ID).Msg("User found for verification")

		foundUser, err = tx.User.
			UpdateOneID(foundUser.ID).
			SetVerified(true).
			Save(ctx)
		if err != nil {
			r.Logger.Error().Err(err).Int("user_id", foundUser.ID).Msg("Failed to set user verified")
			return err
		}
		r.Logger.Debug().Int("user_id", foundUser.ID).Msg("User account set to verified")

		if _, err := tx.VerificationCode.
			Delete().
			Where(verificationcode.CodeEQ(code)).
			Exec(ctx); err != nil {
			r.Logger.Error().Err(err).Str("verification_code", code).Msg("Failed to delete verification code")
			return err
		}

		foundUser, _ = getUserWithRoles(ctx, tx, user.IDEQ(foundUser.ID))

		r.Logger.Debug().Str("verification_code", code).Msg("Verification code deleted")

		return nil
	}); err != nil {
		r.Logger.Error().Err(err).Str("verification_code", code).Msg("Transaction failed for VerifyAccount")
		return nil, err
	}

	InvalidateCache(ctx, foundUser, r)

	r.Logger.Info().Int("user_id", foundUser.ID).Msg("Account successfully verified")
	return foundUser, nil
}

func (r *UsersRepository) GetUserBySecondFactor(ctx context.Context, code string) (*ent.User, error) {
	r.Logger.Info().Str("second_factor_id", code).Msg("Attempting to get user by second factor code")
	var foundUser *ent.User
	var err error

	if err := WithTransaction(ctx, r.DB, func(tx *ent.Tx) error {
		foundUser, err = GetUserBySecondFactor(ctx, tx, code)
		if err != nil {
			r.Logger.Error().Err(err).Str("second_factor_id", code).Msg("Failed to get user by second factor code within transaction")
			return err
		}

		r.Logger.Debug().Int("user_id", foundUser.ID).Msg("User found by second factor code within transaction")
		return nil
	}); err != nil {
		r.Logger.Error().Err(err).Str("second_factor_id", code).Msg("Transaction failed for GetUserBySecondFactor")
		return nil, err
	}

	r.Logger.Info().Int("user_id", foundUser.ID).Msg("Successfully retrieved user by second factor code")
	return foundUser, nil
}

func (r *UsersRepository) GetUserByVerificationCode(ctx context.Context, code string) (*ent.User, error) {
	r.Logger.Info().Str("verification_code", code).Msg("Attempting to get user by verification code")
	var foundUser *ent.User
	var err error

	if err := WithTransaction(ctx, r.DB, func(tx *ent.Tx) error {
		foundUser, err = GetUserByVerificationCode(ctx, tx, code)
		if err != nil {
			r.Logger.Error().Err(err).Str("verification_code", code).Msg("Failed to get user by verification code within transaction")
			return err
		}
		r.Logger.Debug().Int("user_id", foundUser.ID).Msg("User found by verification code within transaction")
		return nil
	}); err != nil {
		r.Logger.Error().Err(err).Str("verification_code", code).Msg("Transaction failed for GetUserByVerificationCode")
		return nil, err
	}

	if err := UpdateCache(ctx, foundUser, r); err != nil {
		return nil, err
	}

	r.Logger.Info().Int("user_id", foundUser.ID).Msg("Successfully retrieved user by verification code")
	return foundUser, nil
}

func (r *UsersRepository) GetUserByResetCode(ctx context.Context, code string) (*ent.User, error) {
	r.Logger.Info().Str("reset_code", code).Msg("Attempting to get user by reset code")
	var foundUser *ent.User
	var err error

	if err := WithTransaction(ctx, r.DB, func(tx *ent.Tx) error {
		foundUser, err = GetUserByResetCode(ctx, tx, code)
		if err != nil {
			r.Logger.Error().Err(err).Str("reset_code", code).Msg("Failed to get user by reset code within transaction")
			return err
		}
		r.Logger.Debug().Int("user_id", foundUser.ID).Msg("User found by reset code within transaction")
		return nil
	}); err != nil {
		r.Logger.Error().Err(err).Str("reset_code", code).Msg("Transaction failed for GetUserByResetCode")
		return nil, err
	}

	if err := UpdateCache(ctx, foundUser, r); err != nil {
		return nil, err
	}

	r.Logger.Info().Int("user_id", foundUser.ID).Msg("Successfully retrieved user by reset code")
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

	if _ = UpdateCache(ctx, foundUser, r); err != nil {
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
		newUser, err = tx.User.
			Create().
			SetMail(req.Mail).
			SetPassword(hashedPassword).
			Save(ctx)
		if err != nil {
			r.Logger.Error().Err(err).Str("mail", req.Mail).Msg("Failed to create user")
			return err
		}
		r.Logger.Debug().Int("user_id", newUser.ID).Msg("User created successfully")

		if _, err := tx.UserSettings.
			Create().
			SetUUID(uuid.New().String()).
			SetOwner(newUser).
			Save(ctx); err != nil {
			r.Logger.Error().Err(err).Int("user_id", newUser.ID).Msg("Failed to create user settings")
			return errors.New("failed to create user settings")
		}
		r.Logger.Debug().Int("user_id", newUser.ID).Msg("User settings created")

		if _, err := tx.UserDetails.
			Create().
			SetName(req.Name).
			SetOwner(newUser).
			Save(ctx); err != nil {
			r.Logger.Error().Err(err).Int("user_id", newUser.ID).Msg("Failed to create user details")
			return errors.New("failed to create user details")
		}
		r.Logger.Debug().Int("user_id", newUser.ID).Msg("User details created")

		if _, err := tx.UserDevice.
			Create().
			SetOwner(newUser).
			SetToken(req.DeviceToken).
			SetIPAddress(req.IPAddress).
			SetUserAgent(req.UserAgent).
			SetOsName(req.OSName).
			SetOsVersion(req.OSVersion).
			SetBrowserName(req.BrowserName).
			SetBrowserVersion(req.BrowserVersion).
			Save(ctx); err != nil {
			r.Logger.Error().Err(err).Int("user_id", newUser.ID).Msg("Failed to create user device")
			return errors.New("failed to create user device")
		}
		r.Logger.Debug().Int("user_id", newUser.ID).Str("device_token", req.DeviceToken).Msg("User device created")

		var codeStr string
		for {
			code := utils.GenerateRandomCode()
			codeStr = strconv.Itoa(code)

			if exists, _ := tx.ActivationCode.
				Query().
				Where(activationcode.CodeEQ(codeStr)).
				Exist(ctx); !exists {
				r.Logger.Debug().Str("activation_code", codeStr).Msg("Generated unique activation code")
				break
			}
			r.Logger.Debug().Str("activation_code", codeStr).Msg("Generated activation code already exists, retrying")
		}

		activationCode, err = tx.ActivationCode.
			Create().
			SetCode(codeStr).
			SetOwner(newUser).
			Save(ctx)
		if err != nil {
			r.Logger.Error().Err(err).Int("user_id", newUser.ID).Msg("Failed to create activation code")
			return err
		}
		r.Logger.Debug().Int("activation_code_id", activationCode.ID).Msg("Activation code created")

		return nil
	}); err != nil {
		r.Logger.Error().Err(err).Str("mail", req.Mail).Msg("Transaction failed for CreateUser")
		return nil, nil, err
	}

	r.Logger.Info().Int("user_id", newUser.ID).Msg("User and associated entities created successfully")
	return newUser, activationCode, nil
}

func (r *UsersRepository) CreateSecondFactorCode(
	ctx context.Context,
	user *ent.User,
	device *ent.UserDevice,
) (*ent.SecondFactorCode, error) {
	r.Logger.Info().Int("user_id", user.ID).Int("device_id", device.ID).Msg("Attempting to create second factor code")
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
				r.Logger.Debug().Str("second_factor_id", codeStr).Msg("Generated unique second factor code")
				break
			}
			r.Logger.Debug().Str("second_factor_id", codeStr).Msg("Generated second factor code already exists, retrying")
		}

		secondFactor, err = tx.SecondFactorCode.
			Create().
			SetCode(codeStr).
			SetOwner(user).
			SetTargetUserDeviceID(device.ID).
			Save(ctx)
		if err != nil {
			r.Logger.Error().Err(err).Int("user_id", user.ID).Msg("Failed to create second factor code")
			return err
		}
		r.Logger.Debug().Int("second_factor_id", secondFactor.ID).Msg("Second factor code created")

		return nil
	}); err != nil {
		r.Logger.Error().Err(err).Int("user_id", user.ID).Msg("Transaction failed for CreateSecondFactorCode")
		return nil, err
	}

	InvalidateCache(ctx, user, r)

	r.Logger.Info().Int("user_id", user.ID).Msg("Second factor code created successfully")
	return secondFactor, nil
}

func (r *UsersRepository) CreateResetCode(ctx context.Context, user *ent.User) (*ent.ResetCode, error) {
	r.Logger.Info().Int("user_id", user.ID).Msg("Attempting to create reset code")
	var resetCode *ent.ResetCode
	var err error

	if err := WithTransaction(ctx, r.DB, func(tx *ent.Tx) error {
		var codeStr string

		for {
			code := utils.GenerateRandomCode()
			codeStr = strconv.Itoa(code)

			if exists, _ := tx.ResetCode.
				Query().
				Where(resetcode.CodeEQ(codeStr)).
				Exist(ctx); !exists {
				r.Logger.Debug().Str("reset_code", codeStr).Msg("Generated unique reset code")
				break
			}
			r.Logger.Debug().Str("reset_code", codeStr).Msg("Generated reset code already exists, retrying")
		}

		resetCode, err = tx.ResetCode.
			Create().
			SetCode(codeStr).
			SetOwner(user).
			Save(ctx)
		if err != nil {
			r.Logger.Error().Err(err).Int("user_id", user.ID).Msg("Failed to create reset code")
			return err
		}
		r.Logger.Debug().Int("reset_code_id", resetCode.ID).Msg("Reset code created")

		r.Logger.Debug().Int("user_id", user.ID).Msg("User re-fetched after reset code creation (for consistency check)")

		return nil
	}); err != nil {
		r.Logger.Error().Err(err).Int("user_id", user.ID).Msg("Transaction failed for CreateResetCode")
		return nil, err
	}

	InvalidateCache(ctx, user, r)

	r.Logger.Info().Int("user_id", user.ID).Msg("Reset code created successfully")
	return resetCode, nil
}

func (r *UsersRepository) RemoveResetCode(ctx context.Context, user *ent.User) (*ent.User, error) {
	r.Logger.Info().Int("user_id", user.ID).Msg("Attempting to remove reset code")
	var updatedUser *ent.User
	var err error

	if err := WithTransaction(ctx, r.DB, func(tx *ent.Tx) error {
		if user.Edges.ResetCode == nil {
			r.Logger.Warn().Int("user_id", user.ID).Msg("User has no reset code to remove")
			return errors.New("user has no reset code")
		}
		r.Logger.Debug().Int("reset_code_id", user.Edges.ResetCode.ID).Msg("Reset code found for deletion")

		if err := tx.ResetCode.
			DeleteOneID(user.Edges.ResetCode.ID).
			Exec(ctx); err != nil {
			r.Logger.Error().Err(err).Int("user_id", user.ID).Int("reset_code_id", user.Edges.ResetCode.ID).Msg("Failed to delete reset code")
			return err
		}
		r.Logger.Debug().Int("reset_code_id", user.Edges.ResetCode.ID).Msg("Reset code deleted")

		updatedUser, err = GetUserByID(ctx, tx, user.ID)
		if err != nil {
			r.Logger.Error().Err(err).Int("user_id", user.ID).Msg("Failed to re-fetch user after reset code removal")
			return err
		}
		r.Logger.Debug().Int("user_id", updatedUser.ID).Msg("User re-fetched after reset code removal")

		return nil
	}); err != nil {
		r.Logger.Error().Err(err).Int("user_id", user.ID).Msg("Transaction failed for RemoveResetCode")
		return nil, err
	}

	InvalidateCache(ctx, user, r)

	r.Logger.Info().Int("user_id", updatedUser.ID).Msg("Reset code removed successfully")
	return updatedUser, nil
}

func (r *UsersRepository) RemoveSecondFactorCode(ctx context.Context, targetUser *ent.User) (*ent.User, error) {
	r.Logger.Info().Int("user_id", targetUser.ID).Msg("Attempting to remove second factor code")
	var updatedUser *ent.User
	var err error

	if err := WithTransaction(ctx, r.DB, func(tx *ent.Tx) error {
		if targetUser.Edges.SecondFactorCode == nil {
			r.Logger.Warn().Int("user_id", targetUser.ID).Msg("User has no second factor code to remove")
			return errors.New("user has no reset code")
		}

		r.Logger.Debug().Int("second_factor_code_id", targetUser.Edges.SecondFactorCode.ID).Msg("Second factor code found for deletion")

		updatedUser, err = tx.User.
			UpdateOne(targetUser).
			ClearSecondFactorCode().
			Save(ctx)
		if err != nil {
			r.Logger.Error().Err(err).Int("user_id", targetUser.ID).Int("second_factor_code_id", targetUser.Edges.SecondFactorCode.ID).Msg("Failed to remove second factor code")
			return errors.New("failed to remove second factor code")
		}
		r.Logger.Debug().Int("second_factor_code_id", targetUser.Edges.SecondFactorCode.ID).Msg("Second factor code deleted")

		updatedUser, err = getUserFunctional(ctx, tx, user.IDEQ(targetUser.ID))

		return nil
	}); err != nil {
		r.Logger.Error().Err(err).Int("user_id", targetUser.ID).Msg("Transaction failed for RemoveSecondFactorCode")
		return nil, err
	}

	InvalidateCache(ctx, targetUser, r)

	return updatedUser, nil
}

func (r *UsersRepository) UpdateUsersPassword(ctx context.Context, user *ent.User, hashedPassword string) (*ent.User, error) {
	r.Logger.Info().Int("user_id", user.ID).Msg("Attempting to update user password")
	var updatedUser *ent.User
	var err error

	if err := WithTransaction(ctx, r.DB, func(tx *ent.Tx) error {
		updatedUser, err = tx.User.
			UpdateOneID(user.ID).
			SetPassword(hashedPassword).
			Save(ctx)
		if err != nil {
			r.Logger.Error().Err(err).Int("user_id", user.ID).Msg("Failed to update user password in DB")
			return errors.New("failed to update user")
		}
		r.Logger.Debug().Int("user_id", user.ID).Msg("User password updated in DB")

		return nil
	}); err != nil {
		r.Logger.Error().Err(err).Int("user_id", user.ID).Msg("Transaction failed for UpdateUsersPassword")
		return nil, err
	}

	InvalidateCache(ctx, user, r)

	r.Logger.Info().Int("user_id", updatedUser.ID).Msg("User password updated successfully")
	return updatedUser, nil
}

func (r *UsersRepository) UpdateUser(ctx context.Context, targetUser *ent.User, req *requests.User) (*ent.User, error) {
	r.Logger.Info().Int("user_id", targetUser.ID).Msg("Attempting to update user mail/phone")
	var updatedUser *ent.User

	if err := WithTransaction(ctx, r.DB, func(tx *ent.Tx) error {
		userUpdater := tx.User.UpdateOneID(targetUser.ID)
		updated := false

		if req.Phone != targetUser.Phone {
			if exists, _ := tx.User.
				Query().
				Where(user.PhoneEQ(req.Phone)).
				Exist(ctx); exists {
				return errors.New("phone already exists")
			}

			userUpdater.SetPhone(req.Phone)
			updated = true
			r.Logger.Debug().Int("user_id", targetUser.ID).Str("old_phone", targetUser.Phone).Str("new_phone", req.Phone).Msg("Updating user phone")
		}

		if req.Mail != targetUser.Mail {
			if exists, _ := tx.User.
				Query().
				Where(user.MailEQ(req.Mail)).
				Exist(ctx); exists {
				return errors.New("mail already exists")
			}

			userUpdater.SetMail(req.Mail)
			updated = true
			r.Logger.Debug().Int("user_id", targetUser.ID).Str("old_mail", targetUser.Mail).Str("new_mail", req.Mail).Msg("Updating user mail")
		}

		if updated {
			if _, err := userUpdater.Save(ctx); err != nil {
				r.Logger.Error().Err(err).Int("user_id", targetUser.ID).Msg("Failed to update user mail/phone in DB")
				return errors.New("failed to update user's mail/phone")
			}

			r.Logger.Debug().Int("user_id", targetUser.ID).Msg("User mail/phone updated in DB")
		} else {

			r.Logger.Debug().Int("user_id", targetUser.ID).Msg("No changes detected for user mail/phone, skipping update")
		}

		updatedUser, _ = getUserFunctionalWithDetails(ctx, tx, user.IDEQ(targetUser.ID))

		return nil
	}); err != nil {
		r.Logger.Error().Err(err).Int("user_id", targetUser.ID).Msg("Transaction failed for UpdateUser")
		return nil, errors.New("failed to update user")
	}

	InvalidateCache(ctx, targetUser, r)

	r.Logger.Info().Int("user_id", updatedUser.ID).Msg("User mail/phone updated successfully")
	return updatedUser, nil
}

func (r *UsersRepository) UpdateUsersDetails(
	ctx context.Context,
	targetUser *ent.User,
	req *requests.Details,
) (*ent.User, error) {
	r.Logger.Debug().Int("user_id", targetUser.ID).Msg("Attempting to update user details")
	var updatedUser *ent.User
	var err error

	if err := WithTransaction(ctx, r.DB, func(tx *ent.Tx) error {
		detailsUpdated := tx.UserDetails.
			UpdateOneID(targetUser.Edges.UserDetails.ID).
			SetFirstName(req.FirstName).
			SetLastName(req.LastName)

		if _, err = detailsUpdated.Save(ctx); err != nil {
			r.Logger.Error().Err(err).Int("user_id", targetUser.ID).Msg("Failed to update user's details in DB")
			return errors.New("failed to update user's details")
		}
		r.Logger.Debug().Int("user_id", targetUser.ID).Msg("User details updated in DB")

		updatedUser, err = getUserFunctionalWithDetails(ctx, tx, user.IDEQ(targetUser.ID))

		return nil
	}); err != nil {
		r.Logger.Error().Err(err).Int("user_id", targetUser.ID).Msg("Transaction failed for UpdateUsersDetails")
		return nil, errors.New("failed to update user")
	}

	InvalidateCache(ctx, targetUser, r)

	r.Logger.Debug().Int("user_id", targetUser.ID).Msg("User details updated successfully")
	return updatedUser, nil
}

func (r *UsersRepository) UpdateUsersSettings(ctx context.Context, targetUser *ent.User, req *requests.Settings) (*ent.User, error) {
	r.Logger.Debug().Int("user_id", targetUser.ID).Msg("Attempting to update user settings")
	var updatedUser *ent.User
	var err error

	if err := WithTransaction(ctx, r.DB, func(tx *ent.Tx) error {
		settingsUpdater := tx.UserSettings.
			UpdateOneID(targetUser.Edges.UserSettings.ID).
			SetNightMode(req.NightMode).
			SetTwoFactor(req.TwoFactor).
			SetSecondFactorTargetID(req.SecondFactorTargetID)

		r.Logger.Debug().Int("user_id", targetUser.ID).
			Bool("night_mode", req.NightMode).
			Bool("two_factor", req.TwoFactor).
			Int("second_factor_target_id", req.SecondFactorTargetID).
			Msg("Updating user settings fields")

		settingsUpdater.ClearNotificationTargetDevices().
			AddNotificationTargetDeviceIDs(req.NotificationsTargetDeviceIDs...)
		r.Logger.Debug().Int("user_id", targetUser.ID).
			Ints("notification_target_device_ids", req.NotificationsTargetDeviceIDs).
			Msg("Updating notification target devices")

		if _, err = settingsUpdater.Save(ctx); err != nil {
			r.Logger.Error().Err(err).Int("user_id", targetUser.ID).Msg("Failed to update user's settings in DB")
			return errors.New("failed to update user's settings")
		}
		r.Logger.Debug().Int("user_id", targetUser.ID).Msg("User settings updated in DB")

		updatedUser, _ = getUserWithSettings(ctx, tx, user.IDEQ(targetUser.ID))

		return nil
	}); err != nil {
		r.Logger.Error().Err(err).Int("user_id", targetUser.ID).Msg("Transaction failed for UpdateUsersSettings")
		return nil, errors.New("failed to update user")
	}

	InvalidateCache(ctx, targetUser, r)

	r.Logger.Debug().Int("user_id", targetUser.ID).Msg("User settings updated successfully")
	return updatedUser, nil
}

func (r *UsersRepository) RemoveAccount(
	ctx context.Context,
	user *ent.User,
) error {
	r.Logger.Info().Int("user_id", user.ID).Msg("Attempting to soft delete user account")
	var err error

	if err = WithTransaction(ctx, r.DB, func(tx *ent.Tx) error {
		if user, err := tx.User.
			UpdateOneID(user.ID).
			SetRemoved(true).
			Save(ctx); err != nil {
			r.Logger.Error().Err(err).Int("user_id", user.ID).Msg("Failed to mark user as removed in DB")
			err = errors.New("failed to remove user")
			return err
		}
		r.Logger.Debug().Int("user_id", user.ID).Msg("User account marked as removed in DB")

		return nil
	}); err != nil {
		r.Logger.Error().Err(err).Int("user_id", user.ID).Msg("Transaction failed for RemoveAccount")
		return errors.New("failed to remove user")
	}

	InvalidateCache(ctx, user, r)

	r.Logger.Debug().Int("user_id", user.ID).Msg("User account successfully marked as removed")
	return err
}

func (r *UsersRepository) FindOrCreateDevice(
	ctx context.Context,
	userID int,
	req *requests.Device,
) (*ent.UserDevice, error) {

	r.Logger.Debug().Int("user_id", userID).Str("device_token", req.DeviceToken).Msg("Attempting to find or create device")
	var device *ent.UserDevice

	if err := WithTransaction(ctx, r.DB, func(tx *ent.Tx) error {
		d, err := tx.UserDevice.Query().
			Where(
				userdevice.HasOwnerWith(user.IDEQ(userID)),
				userdevice.Token(req.DeviceToken),
			).
			Only(ctx)

		if err != nil {
			if ent.IsNotFound(err) {
				r.Logger.Debug().Int("user_id", userID).Str("device_token", req.DeviceToken).Msg("Device not found, attempting to create new device")
				newDevice, err := tx.UserDevice.
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
				if err != nil {
					r.Logger.Error().Err(err).Int("user_id", userID).Str("device_token", req.DeviceToken).Msg("Failed to create new device")
					return err
				}
				device = newDevice
				r.Logger.Debug().Int("user_id", userID).Int("device_id", device.ID).Msg("New device created successfully")

				user, _ := GetUserPublicByID(ctx, tx, userID)
				InvalidateCache(ctx, user, r)

				return nil
			}
			r.Logger.Error().Err(err).Int("user_id", userID).Str("device_token", req.DeviceToken).Msg("Failed to query device in DB")
			return err
		}

		device = d
		r.Logger.Debug().Int("user_id", userID).Int("device_id", device.ID).Msg("Existing device found")
		return nil
	}); err != nil {
		r.Logger.Error().Err(err).Int("user_id", userID).Str("device_token", req.DeviceToken).Msg("Transaction failed for FindOrCreateDevice")
		return nil, errors.New("failed to find or create device in transaction")
	}

	r.Logger.Debug().Int("user_id", userID).Int("device_id", device.ID).Msg("Device operation completed successfully")
	return device, nil
}

func (r *UsersRepository) UpdateEntrepreneurDetails(
	ctx context.Context,
	targetUser *ent.User,
	req *requests.EntrepreneurDetails,
) (*ent.User, error) {
	r.Logger.Info().Int("user_id", targetUser.ID).Msg("Attempting to update entrepreneur details")
	var updatedUser *ent.User
	var err error

	if err = WithTransaction(ctx, r.DB, func(tx *ent.Tx) error {
		details, _ := tx.EntrepreneurDetails.
			Query().
			Where(entrepreneurdetails.HasUserDetailsWith(userdetails.IDEQ(targetUser.Edges.UserDetails.ID))).
			Only(ctx)
		if details == nil {
			if _, err := tx.EntrepreneurDetails.
				Create().
				SetUserDetailsID(targetUser.Edges.UserDetails.ID).
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
				SetUserDetailsID(targetUser.Edges.UserDetails.ID).
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

		updatedUser, err = getUserFunctionalWithDetails(ctx, tx, user.IDEQ(targetUser.ID))

		return nil
	}); err != nil {
		r.Logger.Error().Err(err).Int("user_id", targetUser.ID).Msg("Transaction failed for RemoveAccount")
		return nil, errors.New("failed to remove user")
	}

	InvalidateCache(ctx, targetUser, r)

	return updatedUser, nil
}

func (r *UsersRepository) UpdateLocation(
	ctx context.Context,
	targetUser *ent.User,
	req *requests.Location,
) (*ent.User, error) {
	r.Logger.Debug().Int("user_id", targetUser.ID).Msg("Attempting to update user location")
	var updatedUser *ent.User
	var err error

	if err = WithTransaction(ctx, r.DB, func(tx *ent.Tx) error {
		foundLocation, _ := tx.Location.
			Query().
			Where(location.HasUserDetailsWith(userdetails.IDEQ(targetUser.Edges.UserDetails.ID))).
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

		r.Logger.Debug().Int("user_id", targetUser.ID).Msg("User account marked as removed in DB")

		updatedUser, _ = getUserFunctionalWithDetails(ctx, tx, user.IDEQ(targetUser.ID))

		return nil
	}); err != nil {
		r.Logger.Error().Err(err).Int("user_id", targetUser.ID).Msg("Transaction failed for RemoveAccount")
		return nil, errors.New("failed to remove user")
	}

	InvalidateCache(ctx, targetUser, r)

	r.Logger.Debug().Int("user_id", targetUser.ID).Msg("User account successfully marked as removed")

	return updatedUser, nil
}

func (r *UsersRepository) AddSubscriptions(ctx context.Context, user *ent.User, subIDs []int) error {
	r.Logger.Debug().Int("user_id", user.ID).Msg("Attempting to add user's subscriptions")
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

		return nil

	}); err != nil {
		r.Logger.Error().Err(err).Int("user_id", user.ID).Msg("Transaction failed for AddSubscriptions")
		return errors.New("failed to remove user")
	}

	InvalidateCache(ctx, user, r)

	r.Logger.Debug().Int("user_id", user.ID).Msg("Successfully added subscriptions to user account")

	return nil
}

func (r *UsersRepository) RemoveSubscriptions(ctx context.Context, user *ent.User, subIDs []int) error {
	r.Logger.Debug().Int("user_id", user.ID).Msg("Attempting to revoke user's subscriptions")
	var err error

	if err = WithTransaction(ctx, r.DB, func(tx *ent.Tx) error {
		newSubIDs := utils.RemoveIDs(user.SubscriptionIds, subIDs)

		if _, err := tx.User.
			UpdateOneID(user.ID).
			SetSubscriptionIds(newSubIDs).
			Save(ctx); err != nil {
			return errors.New("failed update user")
		}

		return nil
	}); err != nil {
		r.Logger.Error().Err(err).Int("user_id", user.ID).Msg("Transaction failed for RemoveSubscriptions")
		return errors.New("failed to remove user")
	}

	InvalidateCache(ctx, user, r)

	r.Logger.Debug().Int("user_id", user.ID).Msg("Successfully removed subscriptions from user account")

	return nil
}

func (r *UsersRepository) CreateUserAction(
	ctx context.Context,
	user *ent.User,
	action, originDevice, details string,
) error {
	r.Logger.Debug().Int("user_id", user.ID).Msg("Attempting to create user action")
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

		return nil
	}); err != nil {
		r.Logger.Error().Err(err).Int("user_id", user.ID).Msg("Transaction failed for CreateUserAction")
		return errors.New("failed to remove user")
	}

	InvalidateCache(ctx, user, r)

	r.Logger.Debug().Int("user_id", user.ID).Msg("Successfully created UserAction")
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

	r.Logger.Debug().Int("page", page).Int("page_size", pageSize).Msg("Fetching public users with offset")

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
			r.Logger.Warn().Err(err).Int("user_id", user.ID).Msg("Failed to update cache for a user, continuing")
		}
	}

	return foundUsers, nil
}

func (r *UsersRepository) AddRole(ctx context.Context, userID, roleID int) error {
	r.Logger.Debug().Int("user_id", userID).Msg("Attempting to add role for user")
	var updatedUser *ent.User
	var err error

	if err = WithTransaction(ctx, r.DB, func(tx *ent.Tx) error {
		updatedUser, err = tx.User.
			UpdateOneID(userID).
			AddUserRoleIDs(roleID).
			Save(ctx)
		if err != nil {
			return errors.New("failed to add role for user")
		}

		return nil
	}); err != nil {
		r.Logger.Error().Err(err).Int("user_id", userID).Msg("Transaction failed for AddRole")
		return errors.New("failed to remove user")
	}

	InvalidateCache(ctx, updatedUser, r)

	r.Logger.Debug().Int("user_id", userID).Msg("Successfully added role for user")
	return nil
}

func (r *UsersRepository) RevokeRole(ctx context.Context, userID, roleID int) error {
	r.Logger.Debug().Int("user_id", userID).Msg("Attempting to revoke role for user")
	var updatedUser *ent.User
	var err error

	if err = WithTransaction(ctx, r.DB, func(tx *ent.Tx) error {
		updatedUser, err = tx.User.
			UpdateOneID(userID).
			RemoveUserRoleIDs(roleID).
			Save(ctx)
		if err != nil {
			return errors.New("failed to revoke role for user")
		}

		return nil
	}); err != nil {
		r.Logger.Error().Err(err).Int("user_id", userID).Msg("Transaction failed for AddRole")
		return errors.New("failed to remove user")
	}

	InvalidateCache(ctx, updatedUser, r)

	r.Logger.Debug().Int("user_id", userID).Msg("Successfully added role for user")
	return nil
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
