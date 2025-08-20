package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"
	"users-service/app/models/ent"
	"users-service/app/models/ent/activationcode"
	"users-service/app/models/ent/predicate"
	"users-service/app/models/ent/resetcode"
	"users-service/app/models/ent/secondfactorcode"
	"users-service/app/models/ent/user"
	"users-service/app/models/ent/verificationcode"
	"users-service/app/utils"
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

func UpdateCache(ctx context.Context, user *ent.User, r *UsersRepository) error {
	payload, err := r.CacheStore.CacheUser(user)
	if err != nil {
		return errors.New("failed to cache user")
	}

	cacheKeys := utils.GetUserKeys(user)
	for _, key := range cacheKeys {
		if err := r.CacheStore.Set(ctx, key, payload, time.Minute*5); err != nil {
			return errors.New("failed to set user cache")
		}
	}

	return nil
}

func withFullUserData(q *ent.UserQuery) *ent.UserQuery {
	return q.
		Where(user.BlacklistedEQ(false), user.RemovedEQ(false)).
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
		WithVerificationCode()
}

func getUser(ctx context.Context, tx *ent.Tx, where ...predicate.User) (*ent.User, error) {
	q := tx.User.Query()
	q = withFullUserData(q)
	q.Where(where...)
	return q.Only(ctx)
}

func GetUserByID(ctx context.Context, tx *ent.Tx, ID int) (*ent.User, error) {
	return getUser(ctx, tx, user.IDEQ(ID))
}

func GetUserPublicByID(ctx context.Context, tx *ent.Tx, ID int) (*ent.User, error) {
	return tx.User.
		Query().
		Where(user.IDEQ(ID), user.BlacklistedEQ(false), user.RemovedEQ(false)).
		Only(ctx)
}

func GetUserByMail(ctx context.Context, tx *ent.Tx, mail string) (*ent.User, error) {
	return getUser(ctx, tx, user.MailEQ(mail))
}

func GetUserByPhone(ctx context.Context, tx *ent.Tx, phone string) (*ent.User, error) {
	return getUser(ctx, tx, user.PhoneEQ(phone))
}

func GetUserBySecondFactor(ctx context.Context, tx *ent.Tx, code string) (*ent.User, error) {
	return getUser(ctx, tx, user.HasSecondFactorCodeWith(secondfactorcode.CodeEQ(code)))
}

func GetUserByActivationCode(ctx context.Context, tx *ent.Tx, code string) (*ent.User, error) {
	return getUser(ctx, tx, user.HasActivationCodeWith(activationcode.CodeEQ(code)))
}

func GetUserByVerificationCode(ctx context.Context, tx *ent.Tx, code string) (*ent.User, error) {
	return getUser(ctx, tx, user.HasVerificationCodeWith(verificationcode.CodeEQ(code)))
}

func GetUserByResetCode(ctx context.Context, tx *ent.Tx, code string) (*ent.User, error) {
	return getUser(ctx, tx, user.HasResetCodeWith(resetcode.CodeEQ(code)))
}
