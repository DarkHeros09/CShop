package token

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidToken = errors.New("token is invalid")

	ErrExpiredToken = errors.New("token has expired")
)

// Payload contains the user payload data of the token
type UserPayload struct {
	ID        uuid.UUID `json:"id"`
	UserID    int64     `json:"user_id"`
	Username  string    `json:"username"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiredAt time.Time `json:"expired_at"`
}

// NewPayload creates a new token payload with a specific username and duration
func NewPayloadForUser(userID int64, username string, duration time.Duration) (*UserPayload, error) {
	tokenID, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	payload := &UserPayload{
		ID:        tokenID,
		UserID:    userID,
		Username:  username,
		IssuedAt:  time.Now(),
		ExpiredAt: time.Now().Add(duration),
	}

	return payload, nil
}

func (payload *UserPayload) ValidUser() error {
	if time.Now().After(payload.ExpiredAt) {
		return ErrExpiredToken
	}
	return nil
}

// Payload contains the admin payload data of the token
type AdminPayload struct {
	ID        uuid.UUID `json:"id"`
	AdminID   int64     `json:"admin_id"`
	Username  string    `json:"username"`
	TypeID    int64     `json:"type_id"`
	Active    bool      `json:"active"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiredAt time.Time `json:"expired_at"`
}

// NewPayload creates a new token payload with a specific Admin and duration
func NewPayloadForAdmin(adminID int64, username string, type_id int64, active bool, duration time.Duration) (*AdminPayload, error) {
	tokenID, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	payload := &AdminPayload{
		ID:        tokenID,
		AdminID:   adminID,
		Username:  username,
		TypeID:    type_id,
		Active:    active,
		IssuedAt:  time.Now(),
		ExpiredAt: time.Now().Add(duration),
	}

	return payload, nil
}

func (payload *AdminPayload) ValidAdmin() error {
	if time.Now().After(payload.ExpiredAt) {
		return ErrExpiredToken
	}
	return nil
}
