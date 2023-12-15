package token

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Payload struct {
	ID uuid.UUID `json:"id"`

	Email     string    `json:"email"`
	Role      string    `json:"role"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiredAt time.Time `json:"expired_at"`
}

func NewPayload(email string, role string, duration time.Duration) (*Payload, error) {
	tokenID, err := uuid.NewRandom()

	if err != nil {
		return nil, err
	}

	payload := Payload{
		ID:        tokenID,
		Email:     email,
		Role:      role,
		IssuedAt:  time.Now(),
		ExpiredAt: time.Now().Add(duration),
	}

	return &payload, nil
}

func (p *Payload) Val() error {
	if time.Now().After(p.ExpiredAt) {
		return errors.New("token has expired")
	}

	return nil
}
