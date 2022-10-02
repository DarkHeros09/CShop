package token

import (
	"fmt"
	"time"

	"github.com/o1egl/paseto"
	"golang.org/x/crypto/chacha20poly1305"
)

// Paseto is a PASETO token maker
type PasetoMaker struct {
	paseto       *paseto.V2
	symmetricKey []byte
}

// NewPasetoMaker creates a new PasetoMaker
func NewPasetoMaker(symmetricKey string) (Maker, error) {
	if len(symmetricKey) != chacha20poly1305.KeySize {
		return nil, fmt.Errorf("invalid key size: must be exactly %d characters", chacha20poly1305.KeySize)
	}

	maker := &PasetoMaker{
		paseto:       paseto.NewV2(),
		symmetricKey: []byte(symmetricKey),
	}

	return maker, nil
}

// CreateToken creates a new token for specific username and duration
func (maker *PasetoMaker) CreateTokenForUser(userID int64, username string, duration time.Duration) (string, *UserPayload, error) {
	payload, err := NewPayloadForUser(userID, username, duration)
	if err != nil {
		return "", payload, err
	}

	token, err := maker.paseto.Encrypt(maker.symmetricKey, payload, nil)
	if err != nil {
		return "", payload, err
	}
	return token, payload, err
}

// VerifyToken checks if the token is valid or not
func (maker *PasetoMaker) VerifyTokenForUser(token string) (*UserPayload, error) {
	userPayload := &UserPayload{}

	err := maker.paseto.Decrypt(token, maker.symmetricKey, userPayload, nil)
	if err != nil {
		return nil, ErrInvalidToken
	}

	err = userPayload.ValidUser()
	if err != nil {
		return nil, err
	}

	return userPayload, nil
}

// CreateToken creates a new admin token for specific admin and duration
func (maker *PasetoMaker) CreateTokenForAdmin(adminID int64, username string, type_id int64, active bool, duration time.Duration) (string, *AdminPayload, error) {
	payload, err := NewPayloadForAdmin(adminID, username, type_id, active, duration)
	if err != nil {
		return "", payload, err
	}

	token, err := maker.paseto.Encrypt(maker.symmetricKey, payload, nil)
	if err != nil {
		return "", payload, err
	}
	return token, payload, err
}

// VerifyToken checks if the token is valid or not
func (maker *PasetoMaker) VerifyTokenForAdmin(token string) (*AdminPayload, error) {
	adminPayload := &AdminPayload{}

	err := maker.paseto.Decrypt(token, maker.symmetricKey, adminPayload, nil)
	if err != nil {
		return nil, ErrInvalidToken
	}

	err = adminPayload.ValidAdmin()
	if err != nil {
		return nil, err
	}

	return adminPayload, nil
}
