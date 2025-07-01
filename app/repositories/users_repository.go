package repositories

import (
	"context"
	"errors"
	"strconv"
	"time"
	"users-service/app/infrastructure"
	"users-service/app/models/ent"
	"users-service/app/models/ent/activationcode"
	"users-service/app/models/ent/secondfactorcode"
	"users-service/app/models/ent/user"
	"users-service/app/models/ent/userdevice"
	"users-service/app/models/ent/verificationcode"
	"users-service/app/models/requests"
	"users-service/app/utils"

	"entgo.io/ent/dialect/sql"
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

	GetUserByID(ctx context.Context, ID int) (*ent.User, error)
	GetUserByMail(ctx context.Context, mail string) (*ent.User, error)
	GetUserByPhone(ctx context.Context, phone string) (*ent.User, error)
	GetUserBySecondFactor(ctx context.Context, code string) (*ent.User, error)
	GetUserByVerificationCode(ctx context.Context, code string) (*ent.User, error)
	GetUserByResetCode(ctx context.Context, code string) (*ent.User, error)

	FindOrCreateDevice(ctx context.Context, userID int, req *requests.Device) (*ent.UserDevice, error)

	ActivateAccount(ctx context.Context, code string) (*ent.User, error)
	VerifyAccount(ctx context.Context, code string) (*ent.User, error)
	UpdateUser(ctx context.Context, user *ent.User, req *requests.User) (*ent.User, error)
	UpdateUsersPassword(ctx context.Context, user *ent.User, hashedPassword string) (*ent.User, error)
	UpdateUsersDetails(ctx context.Context, user *ent.User, req *requests.Details) (*ent.User, error)
	UpdateUsersSettings(ctx context.Context, user *ent.User, req *requests.Settings) (*ent.User, error)

	RemoveResetCode(ctx context.Context, user *ent.User) (*ent.User, error)
	RemoveSecondFactorCode(ctx context.Context, user *ent.User) (*ent.User, error)
	RemoveAccount(ctx context.Context, user *ent.User) error
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
	var foundUser *ent.User
	var err error

	if err := WithTransaction(ctx, r.DB, func(tx *ent.Tx) error {
		foundUser, err = GetUserByID(ctx, tx, ID)
		if err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return foundUser, nil
}

func (r *UsersRepository) GetUserByMail(ctx context.Context, mail string) (*ent.User, error) {
	var foundUser *ent.User
	var err error

	if err := WithTransaction(ctx, r.DB, func(tx *ent.Tx) error {
		foundUser, err = GetUserByMail(ctx, tx, mail)
		if err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return foundUser, nil
}

func (r *UsersRepository) ActivateAccount(ctx context.Context, code string) (*ent.User, error) {
	var foundUser *ent.User
	var err error

	if err := WithTransaction(ctx, r.DB, func(tx *ent.Tx) error {
		foundUser, err = GetUserByActivationCode(ctx, tx, code)
		if err != nil {
			return err
		}

		if _, err := foundUser.
			Update().
			SetActive(true).
			Save(ctx); err != nil {
			return err
		}

		if _, err := tx.ActivationCode.
			Delete().
			Where(activationcode.CodeEQ(code)).
			Exec(ctx); err != nil {
			return err
		}

		foundUser, err = GetUserByID(ctx, tx, foundUser.ID)
		if err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return foundUser, nil
}

func (r *UsersRepository) VerifyAccount(ctx context.Context, code string) (*ent.User, error) {
	var foundUser *ent.User
	var err error

	if err := WithTransaction(ctx, r.DB, func(tx *ent.Tx) error {
		foundUser, err = GetUserByVerificationCode(ctx, tx, code)
		if err != nil {
			return err
		}

		if _, err := foundUser.
			Update().
			SetVerified(true).
			Save(ctx); err != nil {
			return err
		}

		if _, err := tx.VerificationCode.
			Delete().
			Where(verificationcode.CodeEQ(code)).
			Exec(ctx); err != nil {
			return err
		}

		foundUser, err = GetUserByID(ctx, tx, foundUser.ID)
		if err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return foundUser, nil
}

func (r *UsersRepository) GetUserBySecondFactor(ctx context.Context, code string) (*ent.User, error) {
	var foundUser *ent.User
	var err error

	if err := WithTransaction(ctx, r.DB, func(tx *ent.Tx) error {
		foundUser, err = GetUserBySecondFactor(ctx, tx, code)
		if err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return foundUser, nil
}

func (r *UsersRepository) GetUserByVerificationCode(ctx context.Context, code string) (*ent.User, error) {
	var foundUser *ent.User
	var err error

	if err := WithTransaction(ctx, r.DB, func(tx *ent.Tx) error {
		foundUser, err = GetUserByVerificationCode(ctx, tx, code)
		if err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return foundUser, nil
}

func (r *UsersRepository) GetUserByResetCode(ctx context.Context, code string) (*ent.User, error) {
	var foundUser *ent.User
	var err error

	if err := WithTransaction(ctx, r.DB, func(tx *ent.Tx) error {
		foundUser, err = GetUserByResetCode(ctx, tx, code)
		if err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return foundUser, nil
}

func (r *UsersRepository) GetUserByPhone(ctx context.Context, phone string) (*ent.User, error) {
	var foundUser *ent.User
	var err error

	if err := WithTransaction(ctx, r.DB, func(tx *ent.Tx) error {
		foundUser, err = GetUserByPhone(ctx, tx, phone)
		if err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return foundUser, nil
}

func (r *UsersRepository) CreateUser(ctx context.Context, req *requests.Register, hashedPassword string) (*ent.User, *ent.ActivationCode, error) {
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
			return err
		}

		if _, err := tx.UserSettings.
			Create().
			SetOwner(newUser).
			Save(ctx); err != nil {
			return errors.New("failed to create user settings")
		}

		if _, err := tx.UserDetails.
			Create().
			SetName(req.Name).
			SetOwner(newUser).
			Save(ctx); err != nil {
			return errors.New("failed to create user details")
		}

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
			return errors.New("failed to create user device")
		}

		var codeStr string

		for {
			code := utils.GenerateRandomCode()
			codeStr = strconv.Itoa(code)

			if exists, _ := tx.ActivationCode.
				Query().
				Where(activationcode.CodeEQ(codeStr)).
				Exist(ctx); !exists {
				break
			}
		}

		activationCode, err = tx.ActivationCode.
			Create().
			SetCode(codeStr).
			SetOwner(newUser).
			Save(ctx)
		if err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, nil, err
	}

	return newUser, activationCode, nil
}

func (r *UsersRepository) CreateSecondFactorCode(ctx context.Context, user *ent.User, device *ent.UserDevice) (*ent.User, *ent.SecondFactorCode, error) {
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
				return err
			}
			if !exists {
				break
			}
		}

		secondFactor, err = tx.SecondFactorCode.
			Create().
			SetCode(codeStr).
			SetOwner(user).
			SetTargetUserDeviceID(device.ID).
			Save(ctx)
		if err != nil {
			return err
		}

		updatedUser, err = GetUserByID(ctx, tx, user.ID)
		if err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, nil, err
	}

	return updatedUser, secondFactor, nil
}

func (r *UsersRepository) CreateResetCode(ctx context.Context, user *ent.User) (*ent.ResetCode, error) {
	var resetCode *ent.ResetCode
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
				break
			}
		}

		resetCode, err = tx.ResetCode.
			Create().
			SetCode(codeStr).
			SetOwner(user).
			Save(ctx)
		if err != nil {
			return err
		}

		_, err := GetUserByID(ctx, tx, user.ID)
		if err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return resetCode, nil
}

func (r *UsersRepository) RemoveResetCode(ctx context.Context, user *ent.User) (*ent.User, error) {
	var updatedUser *ent.User
	var err error

	if err := WithTransaction(ctx, r.DB, func(tx *ent.Tx) error {
		if user.Edges.ResetCode == nil {
			return errors.New("user has no reset code")
		}

		if err := tx.ResetCode.
			DeleteOneID(user.Edges.ResetCode.ID).
			Exec(ctx); err != nil {
			return err
		}

		updatedUser, err = GetUserByID(ctx, tx, user.ID)
		if err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return updatedUser, nil
}

func (r *UsersRepository) RemoveSecondFactorCode(ctx context.Context, user *ent.User) (*ent.User, error) {
	var updatedUser *ent.User
	var err error

	if err := WithTransaction(ctx, r.DB, func(tx *ent.Tx) error {
		if user.Edges.SecondFactorCode == nil {
			return errors.New("user has no reset code")
		}

		if err := tx.SecondFactorCode.
			DeleteOneID(user.Edges.SecondFactorCode.ID).
			Exec(ctx); err != nil {
			return errors.New("failed to remove second factor code")
		}

		updatedUser, err = GetUserByID(ctx, tx, user.ID)
		if err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}

	cachedUser, err := r.CacheStore.CacheUser(updatedUser)
	r.CacheStore.Set(ctx, "", cachedUser, time.Minute*5)

	return updatedUser, nil
}

func (r *UsersRepository) UpdateUsersPassword(ctx context.Context, user *ent.User, hashedPassword string) (*ent.User, error) {
	var updatedUser *ent.User
	var err error

	if err := WithTransaction(ctx, r.DB, func(tx *ent.Tx) error {
		if _, err = tx.User.
			UpdateOneID(user.ID).
			SetPassword(hashedPassword).
			Save(ctx); err != nil {
			return err
		}

		updatedUser, err = GetUserByID(ctx, tx, user.ID)
		if err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return updatedUser, nil
}

func (r *UsersRepository) UpdateUser(ctx context.Context, user *ent.User, req *requests.User) (*ent.User, error) {
	var updatedUser *ent.User
	var err error

	if err := WithTransaction(ctx, r.DB, func(tx *ent.Tx) error {
		if req.Phone != user.Phone {
			if _, err := tx.User.
				UpdateOneID(user.ID).
				SetPhone(req.Phone).
				Save(ctx); err != nil {
				return errors.New("failed to update user's phone")
			}

		}

		if req.Mail != user.Mail {
			if _, err := tx.User.
				UpdateOneID(user.ID).
				SetMail(req.Mail).
				Save(ctx); err != nil {
				return errors.New("failed to update user's mail")
			}
		}

		updatedUser, err = GetUserByID(ctx, tx, user.ID)
		if err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, errors.New("failed to update user")
	}

	return updatedUser, nil
}

func (r *UsersRepository) UpdateUsersDetails(ctx context.Context, user *ent.User, req *requests.Details) (*ent.User, error) {
	var updatedUser *ent.User
	var err error

	if err := WithTransaction(ctx, r.DB, func(tx *ent.Tx) error {
		if _, err = tx.UserDetails.
			UpdateOneID(user.Edges.UserDetails.ID).
			Save(ctx); err != nil {
			return errors.New("failed to update user's details")
		}

		updatedUser, err = GetUserByID(ctx, tx, user.ID)
		if err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, errors.New("failed to update user")
	}

	return updatedUser, nil
}

func (r *UsersRepository) UpdateUsersSettings(ctx context.Context, user *ent.User, req *requests.Settings) (*ent.User, error) {
	var updatedUser *ent.User
	var err error

	if err := WithTransaction(ctx, r.DB, func(tx *ent.Tx) error {
		if _, err = tx.UserSettings.
			UpdateOneID(user.Edges.UserSettings.ID).
			SetNightMode(req.NightMode).
			SetTwoFactor(req.TwoFactor).
			SetSecondFactorTargetID(req.SecondFactorTargetID).
			ClearNotificationTargetDevices().
			AddNotificationTargetDeviceIDs(req.NotificationsTargetDeviceIDs...).
			Save(ctx); err != nil {
			return errors.New("failed to update user's settings")
		}

		updatedUser, err = GetUserByID(ctx, tx, user.ID)
		if err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, errors.New("failed to update user")
	}

	return updatedUser, nil
}

func (r *UsersRepository) RemoveAccount(ctx context.Context, user *ent.User) error {
	var err error

	if err = WithTransaction(ctx, r.DB, func(tx *ent.Tx) error {
		if _, err := tx.User.
			UpdateOneID(user.ID).
			SetRemoved(true).
			Save(ctx); err != nil {
			err = errors.New("failed to remove user")
			return err
		}

		return nil
	}); err != nil {
		return errors.New("failed to remove user")
	}

	return err
}

func (r *UsersRepository) FindOrCreateDevice(ctx context.Context, userID int, req *requests.Device) (*ent.UserDevice, error) {
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
					err = createErr
					return createErr
				}
				device = newDevice
				return nil
			}
			return err
		}

		device = d
		return nil
	}); err != nil {
		return nil, errors.New("failed to find or create device in transaction")
	}

	return device, nil
}
