package token

import "time"

// Maker is an interface for manging tokens
type Maker interface {
	// CreateToken creates a new user token for specific username and duration
	CreateTokenForUser(userID int64, username string, duration time.Duration) (string, *UserPayload, error)

	// VerifyTokenForUser checks if the user token is valid or not
	VerifyTokenForUser(token string) (*UserPayload, error)

	// CreateToken creates a new admin token for specific admin and duration
	CreateTokenForAdmin(userID int64, username string, type_id int64, active bool, duration time.Duration) (string, *AdminPayload, error)

	// VerifyTokenForAdmin checks if the admin token is valid or not
	VerifyTokenForAdmin(token string) (*AdminPayload, error)
}
