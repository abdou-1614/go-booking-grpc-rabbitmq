package repository

import (
	"auth/internal/model"
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
)

type sessionPGRepository struct {
	db *pgxpool.Pool
}

func NewSessionPGReposity(db *pgxpool.Pool) *sessionPGRepository {
	return &sessionPGRepository{db: db}
}

func (l *sessionPGRepository) CreateSession(ctx context.Context, arg *model.CreateSessionParams) (*model.SessionResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "sessionPGRepository.CreateSession")

	defer span.Finish()

	var i model.SessionResponse

	if err := l.db.QueryRow(ctx, createSession,
		arg.ID,
		arg.Email,
		arg.RefreshToken,
		arg.ClientIP,
		arg.IsBlocked,
		arg.UserAgent,
		arg.ExpiresAt,
	).Scan(
		&i.ID,
		&i.Email,
		&i.RefreshToken,
		&i.ClientIP,
		&i.IsBlocked,
		&i.UserAgent,
		&i.ExpiresAt,
		&i.CreatedAt,
	); err != nil {
		return nil, errors.Wrap(err, "Scan")
	}

	return &i, nil
}
