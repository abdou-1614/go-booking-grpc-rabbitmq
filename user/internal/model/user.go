package model

import (
	"fmt"
	"strings"
	"time"

	userService "Go-grpc/pb"

	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type User struct {
	ID        uuid.UUID  `json:"id"`
	FirstName string     `json:"first_name" validate:"required,min=3,max=25"`
	LastName  string     `json:"last_name" validate:"required,min=3,max=25"`
	Email     string     `json:"email" validate:"required,email"`
	Password  string     `json:"password" validate:"required,min=6,max=250"`
	CreatedAt *time.Time `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}

type CreateUserRequest struct {
	FirstName string `json:"first_name" validate:"required,min=3,max=25"`
	LastName  string `json:"last_name" validate:"required,min=3,max=25"`
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=6,max=250"`
}

type UserResponse struct {
	ID        uuid.UUID  `json:"id"`
	FirstName string     `json:"first_name" validate:"required,min=3,max=25"`
	LastName  string     `json:"last_name" validate:"required,min=3,max=25"`
	Email     string     `json:"email" validate:"required,email"`
	CreatedAt *time.Time `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}

type Role string

const (
	RoleGuest  Role = "guest"
	RoleMember Role = "member"
	RoleAdmin  Role = "admin"
)

func (e *Role) ToString() string {
	return string(*e)
}

func (e *Role) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = Role(s)
	case string:
		*e = Role(s)
	default:
		return fmt.Errorf("unsupported scan type for Role: %T", src)
	}
	return nil
}

type UserUpdate struct {
	UserID    uuid.UUID `json:"id"`
	FirstName string    `json:"first_name" validate:"omitempty,min=3,max=25" swaggertype:"string"`
	LastName  string    `json:"last_name" validate:"omitempty,min=3,max=25" swaggertype:"string"`
	Email     string    `json:"email" validate:"omitempty,email" swaggertype:"string"`
}

func (u *User) HashPassword() error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)

	if err != nil {
		return nil
	}

	u.Password = string(hashedPassword)

	return nil
}

func (u *User) ComparePassword(password string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)); err != nil {
		return err
	}

	return nil
}

func (u *User) PrepareToCreate() error {
	u.Email = strings.ToLower(strings.TrimSpace(u.Email))
	u.Password = strings.TrimSpace(u.Password)

	if err := u.HashPassword(); err != nil {
		return err
	}

	return nil
}

func (u *UserResponse) ToProto() *userService.User {
	return &userService.User{
		ID:        u.ID.String(),
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Email:     u.Email,
		CreatedAt: timestamppb.New(*u.CreatedAt),
		UpdatedAt: timestamppb.New(*u.UpdatedAt),
	}
}

type ErrorMessage struct {
	Offset    int64     `json:"offset"`
	Error     string    `json:"error"`
	Time      time.Time `json:"time"`
	Partition int32     `json:"partition"`
	Topic     string    `json:"topic"`
}
