package db

import (
	"context"
	"time"

	"github.com/cshop/v3/util"
	"github.com/guregu/null/v6"
)

// SignUpTx contains the input parameters of the purchase transaction
type SignUpTxParams struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type SignUpTxResult struct {
	ID              int64     `json:"id"`
	Username        string    `json:"username"`
	Email           string    `json:"email"`
	IsBlocked       bool      `json:"is_blocked"`
	IsEmailVerified bool      `json:"is_email_verified"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	SecretCode      string    `json:"secret_code"`
}

/*
SignUpTx performs a shop order item delete from DB, and update the new total price in shop order table
*/
func (store *SQLStore) SignUpTx(ctx context.Context, arg SignUpTxParams) (*SignUpTxResult, error) {
	var result *SignUpTxResult
	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		arg1 := CreateUserParams{
			Username: arg.Username,
			Email:    arg.Email,
			Password: arg.Password,
			// Telephone: arg.Telephone,
		}

		user, err := q.CreateUser(ctx, arg1)
		if err != nil {
			return err
		}

		arg2 := CreateVerifyEmailParams{
			UserID: null.IntFromPtr(&user.ID),
			// Email:      user.Email,
			SecretCode: util.GenerateOTP(),
		}

		verifyEmail, err := q.CreateVerifyEmail(ctx, arg2)
		if err != nil {
			return err
		}

		result = &SignUpTxResult{
			ID:              user.ID,
			Username:        user.Username,
			Email:           user.Email,
			IsBlocked:       user.IsBlocked,
			IsEmailVerified: user.IsEmailVerified,
			CreatedAt:       user.CreatedAt,
			UpdatedAt:       user.UpdatedAt,
			SecretCode:      verifyEmail.SecretCode,
		}

		return nil
	})

	return result, err
}
