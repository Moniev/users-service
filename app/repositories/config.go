package repositories

import (
	"context"
	"errors"
	"fmt"
	"users-service/app/models/ent"
	"users-service/app/models/ent/activationcode"
	"users-service/app/models/ent/resetcode"
	"users-service/app/models/ent/secondfactorcode"
	"users-service/app/models/ent/user"
	"users-service/app/models/ent/verificationcode"
)

func WithTransaction(ctx context.Context, client *ent.Client, fn func(tx *ent.Tx) error) error {
	tx, err := client.BeginTx(ctx, nil)

	if err != nil {
		return errors.New("failed to create transaction")
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("failed to rollback transaction %v, prime error: %w", rbErr, err)
		}

		return err
	}

	return tx.Commit()
}

func GetUserByID(ctx context.Context, tx *ent.Tx, ID int) (*ent.User, error) {
	return tx.User.
		Query().
		Where(user.IDEQ(ID), user.BlacklistedEQ(false), user.RemovedEQ(false)).
		WithUserSettings(func(usq *ent.UserSettingsQuery) {
			usq.WithOwner()
		}).
		WithUserDevices().
		WithUserDetails().
		WithUserRoles().
		WithUserActions().
		WithActivationCode().
		WithSecondFactorCode().
		WithResetCode().
		WithVerificationCode().
		Only(ctx)
}

func GetUserByMail(ctx context.Context, tx *ent.Tx, mail string) (*ent.User, error) {
	return tx.User.
		Query().
		Where(user.MailEQ(mail), user.BlacklistedEQ(false), user.RemovedEQ(false)).
		WithUserSettings(func(usq *ent.UserSettingsQuery) {
			usq.WithOwner()
		}).
		WithUserDevices().
		WithUserDetails().
		WithUserRoles().
		WithUserActions().
		WithActivationCode().
		WithSecondFactorCode().
		WithResetCode().
		WithVerificationCode().
		Only(ctx)
}

func GetUserByPhone(ctx context.Context, tx *ent.Tx, mail string) (*ent.User, error) {
	return tx.User.
		Query().
		Where(user.MailEQ(mail), user.BlacklistedEQ(false), user.RemovedEQ(false)).
		WithUserSettings(func(usq *ent.UserSettingsQuery) {
			usq.WithOwner()
		}).
		WithUserDevices().
		WithUserDetails().
		WithUserRoles().
		WithUserActions().
		WithActivationCode().
		WithSecondFactorCode().
		WithResetCode().
		WithVerificationCode().
		Only(ctx)
}

func GetUserBySecondFactor(ctx context.Context, tx *ent.Tx, code string) (*ent.User, error) {
	return tx.User.
		Query().
		Where(user.HasSecondFactorCodeWith(secondfactorcode.CodeEQ(code)), user.BlacklistedEQ(false), user.RemovedEQ(false)).
		WithUserSettings(func(usq *ent.UserSettingsQuery) {
			usq.WithOwner()
		}).
		WithUserDevices().
		WithUserDetails().
		WithUserRoles().
		WithUserActions().
		WithActivationCode().
		WithSecondFactorCode().
		WithResetCode().
		WithVerificationCode().
		Only(ctx)
}

func GetUserByActivationCode(ctx context.Context, tx *ent.Tx, code string) (*ent.User, error) {
	return tx.User.
		Query().
		Where(user.HasActivationCodeWith(activationcode.CodeEQ(code)), user.BlacklistedEQ(false), user.RemovedEQ(false)).
		WithUserSettings(func(usq *ent.UserSettingsQuery) {
			usq.WithOwner()
		}).
		WithUserDevices().
		WithUserDetails().
		WithUserRoles().
		WithUserActions().
		WithActivationCode().
		WithSecondFactorCode().
		WithResetCode().
		WithVerificationCode().
		Only(ctx)
}

func GetUserByVerificationCode(ctx context.Context, tx *ent.Tx, code string) (*ent.User, error) {
	return tx.User.
		Query().
		Where(user.HasVerificationCodeWith(verificationcode.CodeEQ(code)), user.BlacklistedEQ(false), user.RemovedEQ(false)).
		WithUserSettings(func(usq *ent.UserSettingsQuery) {
			usq.WithOwner()
		}).
		WithUserDevices().
		WithUserDetails().
		WithUserRoles().
		WithUserActions().
		WithActivationCode().
		WithSecondFactorCode().
		WithResetCode().
		WithVerificationCode().
		Only(ctx)
}

func GetUserByResetCode(ctx context.Context, tx *ent.Tx, code string) (*ent.User, error) {
	return tx.User.
		Query().
		Where(user.HasResetCodeWith(resetcode.CodeEQ(code)), user.BlacklistedEQ(false), user.RemovedEQ(false)).
		WithUserSettings(func(usq *ent.UserSettingsQuery) {
			usq.WithOwner()
		}).
		WithUserDevices().
		WithUserDetails().
		WithUserRoles().
		WithUserActions().
		WithActivationCode().
		WithSecondFactorCode().
		WithResetCode().
		WithVerificationCode().
		Only(ctx)
}
