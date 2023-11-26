package repository

import (
	"Go-grpc/internal/model"
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
)

type userPGRepository struct {
	db *pgxpool.Pool
}

func NewUserPGRepository(db *pgxpool.Pool) *userPGRepository {
	return &userPGRepository{
		db: db,
	}
}

func (u *userPGRepository) Create(ctx context.Context, user *model.User) (*model.UserResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "userPGRepository.create")

	defer span.Finish()

	var created model.UserResponse

	if err := u.db.QueryRow(
		ctx,
		createUserQuery,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.Password,
	).Scan(&created.ID, &created.FirstName,
		&created.LastName,
		&created.Email,
		&created.CreatedAt,
		&created.UpdatedAt,
	); err != nil {
		return nil, errors.Wrap(err, "Scan")
	}

	return &created, nil
}

func (u *userPGRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.UserResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "userPGRepository.GetByID")

	defer span.Finish()

	var res model.UserResponse

	if err := u.db.QueryRow(ctx, getUserByIDQuery, id).Scan(
		&res.ID,
		&res.FirstName,
		&res.LastName,
		&res.Email,
		&res.Role,
		&res.CreatedAt,
		&res.UpdatedAt,
	); err != nil {
		return nil, errors.Wrap(err, "userPGRepository.GetByID.Scan")
	}

	return &res, nil
}
