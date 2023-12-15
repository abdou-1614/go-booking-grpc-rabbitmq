package token

import (
	"fmt"
	"time"

	"github.com/aead/chacha20poly1305"
	"github.com/o1egl/paseto"
	"github.com/pkg/errors"
)

type Maker interface {
	// CreateToken creates a new token for a specific username and duration
	CreateToken(email string, role string, duration time.Duration) (string, *Payload, error)

	// VerifyToken checks if the token is valid or not
	VerifyToken(token string) (*Payload, error)
}

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

func (m *PasetoMaker) CreateToken(email string, role string, duration time.Duration) (string, *Payload, error) {
	payload, err := NewPayload(email, role, duration)

	if err != nil {
		return "", payload, err
	}

	token, err := m.paseto.Encrypt(m.symmetricKey, payload, nil)

	return token, payload, err
}

func (m *PasetoMaker) VerifyToken(token string) (*Payload, error) {
	payload := &Payload{}

	err := m.paseto.Decrypt(token, m.symmetricKey, payload, nil)

	if err != nil {
		return nil, errors.Wrap(err, "Invalid Token")
	}

	err = payload.Val()

	if err != nil {
		return nil, err
	}

	return payload, nil
}
