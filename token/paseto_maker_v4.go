package token

// import (
// 	"fmt"
// 	"time"

// 	"aidanwoods.dev/go-paseto"
// 	"golang.org/x/crypto/chacha20poly1305"
// )

// // Paseto is a PASETO token maker
// type PasetoMaker struct {
// 	paseto       *paseto.Token
// 	parser       *paseto.Parser
// 	symmetricKey paseto.V4SymmetricKey
// }

// // NewPasetoMaker creates a new PasetoMaker
// func NewPasetoMaker(symmetricKey string) (Maker, error) {
// 	if len(symmetricKey) != chacha20poly1305.KeySize {
// 		return nil, fmt.Errorf("invalid key size: must be exactly %d characters", chacha20poly1305.KeySize)
// 	}

// 	v4SymmetricKey, err := paseto.V4SymmetricKeyFromBytes([]byte(symmetricKey))
// 	if err != nil {
// 		return nil, err
// 	}

// 	pasetoToken := paseto.NewToken()

// 	parser := paseto.NewParserWithoutExpiryCheck()

// 	maker := &PasetoMaker{
// 		paseto:       &pasetoToken,
// 		parser:       &parser,
// 		symmetricKey: v4SymmetricKey,
// 	}

// 	return maker, nil
// }

// // CreateToken creates a new token for specific username and duration
// func (maker *PasetoMaker) CreateTokenForUser(userID int64, username string, duration time.Duration) (string, *UserPayload, error) {

// 	payload, err := NewPayloadForUser(userID, username, duration)
// 	if err != nil {
// 		return "", payload, err
// 	}

// 	maker.paseto.Set("payload", payload)

// 	token := maker.paseto.V4Encrypt(maker.symmetricKey, nil)

// 	return token, payload, err
// }

// // VerifyToken checks if the token is valid or not
// func (maker *PasetoMaker) VerifyTokenForUser(signedToken string) (*UserPayload, error) {

// 	userPayload := &UserPayload{}

// 	token, err := maker.parser.ParseV4Local(maker.symmetricKey, signedToken, nil)
// 	if err != nil {
// 		return nil, ErrInvalidToken
// 	}

// 	err = token.Get("payload", userPayload)
// 	if err != nil {
// 		return nil, err
// 	}

// 	err = userPayload.ValidUser()
// 	if err != nil {
// 		return nil, err
// 	}

// 	return userPayload, nil
// }

// // CreateToken creates a new admin token for specific admin and duration
// func (maker *PasetoMaker) CreateTokenForAdmin(adminID int64, username string, type_id int64, active bool, duration time.Duration) (string, *AdminPayload, error) {
// 	payload, err := NewPayloadForAdmin(adminID, username, type_id, active, duration)
// 	if err != nil || payload == nil {
// 		return "", payload, err
// 	}

// 	maker.paseto.Set("payload", payload)

// 	token := maker.paseto.V4Encrypt(maker.symmetricKey, nil)

// 	return token, payload, err
// }

// // VerifyToken checks if the token is valid or not
// func (maker *PasetoMaker) VerifyTokenForAdmin(signedToken string) (*AdminPayload, error) {
// 	adminPayload := &AdminPayload{}

// 	token, err := maker.parser.ParseV4Local(maker.symmetricKey, signedToken, nil)
// 	if err != nil {
// 		return nil, ErrInvalidToken
// 	}

// 	err = token.Get("payload", adminPayload)
// 	if err != nil {
// 		return nil, err
// 	}

// 	err = adminPayload.ValidAdmin()
// 	if err != nil {
// 		return nil, err
// 	}

// 	return adminPayload, nil
// }
