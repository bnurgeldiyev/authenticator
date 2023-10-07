package repo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	"authenticator/internal/model"
	"authenticator/pkg/postgres"
)

// UserRepo -.
type UserRepo struct {
	*postgres.Postgres
}

// NewUser -.
func NewUser(pg *postgres.Postgres) *UserRepo {
	return &UserRepo{pg}
}

func (r *UserRepo) Create(ctx context.Context, in *model.User, txId int) error {

	zLog := zerolog.Ctx(ctx).With().
		Str("unit", "internal.usecase.repo.UserRepo").
		Str("method", "Create").
		Str("username", in.Username).Logger()

	tx, err := r.GetTxById(txId)
	if err != nil {
		zLog.Err(err).Msgf("UserRepo - Create - r.GetTxById")
		return err
	}

	query, args, err := r.Builder.
		Insert(model.UserTableName).
		Columns("username",
			"password",
			"state",
			"create_ts",
			"update_ts").
		Values(in.Username,
			in.Password,
			in.State,
			in.CreateTs,
			in.UpdateTs).
		Suffix("RETURNING id").
		ToSql()
	if err != nil {
		zLog.Err(err).Msgf("UserRepo - Create - r.Builder")
		return err
	}

	err = tx.QueryRow(ctx, query, args...).Scan(&in.Id)
	if err != nil {
		zLog.Err(err).Msgf("UserRepo - Create - tx.QueryRow - query: %s", query)
		return err
	}

	return nil
}

func (r *UserRepo) GetById(ctx context.Context, id uuid.UUID) (*model.User, error) {
	zLog := zerolog.Ctx(ctx).With().
		Str("unit", "internal.usecase.repo.UserRepo").
		Str("method", "GetById").
		Str("id", id.String()).Logger()

	query, args, err := r.Builder.
		Select("id",
			"username",
			"state",
			"create_ts",
			"update_ts",
			"version").
		From(model.UserTableName).
		Where("id = ?", id).
		Where("state != ?", model.Deleted).
		ToSql()
	if err != nil {
		zLog.Err(err).Msgf("UserRepo - GetById - r.Builder")
		return nil, err
	}
	var data model.User
	err = scanUser(r.Pool.QueryRow(ctx, query, args...), &data)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			zLog.Debug().Msgf("id: %d no results", id)
			return nil, nil
		}
		zLog.Err(err).Msgf("UserRepo - GetById - r.Pool.QueryRow - query: %s", query)
		return nil, err
	}
	return &data, nil
}

func (r *UserRepo) GetPasswordById(ctx context.Context, id uuid.UUID) (*model.User, error) {
	zLog := zerolog.Ctx(ctx).With().
		Str("unit", "internal.usecase.repo.UserRepo").
		Str("method", "GetById").
		Str("id", id.String()).Logger()

	query, args, err := r.Builder.
		Select("id",
			"username",
			"password",
			"state",
			"create_ts",
			"update_ts",
			"version").
		From(model.UserTableName).
		Where("id = ?", id).
		Where("state != ?", model.Deleted).
		ToSql()
	if err != nil {
		zLog.Err(err).Msgf("UserRepo - GetById - r.Builder")
		return nil, err
	}

	var data model.User
	err = scanDetailUser(r.Pool.QueryRow(ctx, query, args...), &data)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			zLog.Debug().Msgf("id: %d no results", id)
			return nil, nil
		}
		zLog.Err(err).Msgf("UserRepo - GetById - r.Pool.QueryRow - query: %s", query)
		return nil, err
	}
	return &data, nil
}

func (r *UserRepo) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	zLog := zerolog.Ctx(ctx).With().
		Str("unit", "internal.usecase.repo.UserRepo").
		Str("method", "GetById").
		Str("username", username).Logger()

	query, args, err := r.Builder.
		Select("id",
			"username",
			"password",
			"state",
			"create_ts",
			"update_ts",
			"version").
		From(model.UserTableName).
		Where("username = ?", username).
		Where("state != ?", model.Deleted).
		ToSql()
	if err != nil {
		zLog.Err(err).Msgf("UserRepo - GetById - r.Builder")
		return nil, err
	}

	var data model.User
	err = scanDetailUser(r.Pool.QueryRow(ctx, query, args...), &data)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			zLog.Debug().Msgf("username: %d no results", username)
			return nil, nil
		}
		zLog.Err(err).Msgf("UserRepo - GetById - r.Pool.QueryRow - query: %s", query)
		return nil, err
	}
	return &data, nil
}

func (r *UserRepo) ChangeState(ctx context.Context, old, new *model.User, txId int) error {
	zLog := zerolog.Ctx(ctx).With().
		Str("unit", "internal.usecase.repo.UserRepo").
		Str("method", "ChangeState").
		Str("id", old.Id.String()).Logger()

	tx, err := r.GetTxById(txId)
	if err != nil {
		zLog.Err(err).Msgf("UserRepo - ChangeState - r.GetTxById")
		return err
	}

	query, args, err := r.Builder.
		Update(model.UserTableName).
		Where("id = ?", old.Id).
		Where("version = ?", old.Version).
		SetMap(map[string]interface{}{
			"state":     new.State,
			"update_ts": new.UpdateTs,
			"version":   new.Version,
		}).
		ToSql()
	if err != nil {
		zLog.Err(err).Msgf("UserRepo - Update - r.Builder")
		return err
	}

	var cmdTag pgconn.CommandTag
	cmdTag, err = tx.Exec(ctx, query, args...)
	if err != nil {
		zLog.Err(err).Msgf("UserRepo - Update - tx.Exec - query: %s", query)
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		zLog.Error().Msgf("UserRepo - Update - tx.Exec - no rows affected - query: %s", query)
		return model.ErrNoRowsAffected
	}

	return nil
}

func scanUser(row pgx.Row, item *model.User) (err error) {
	// id, username, state, create_ts, update_ts, version

	err = row.Scan(&item.Id, &item.Username, &item.State, &item.CreateTs, &item.UpdateTs, &item.Version)
	if err == nil {
		item.CreateTs = item.CreateTs.In(time.UTC)
		item.UpdateTs = item.UpdateTs.In(time.UTC)
	}
	return
}

func scanDetailUser(row pgx.Row, item *model.User) (err error) {
	// id, username, password, state, create_ts, update_ts, version

	err = row.Scan(&item.Id, &item.Username, &item.Password, &item.State, &item.CreateTs, &item.UpdateTs, &item.Version)
	if err == nil {
		item.CreateTs = item.CreateTs.In(time.UTC)
		item.UpdateTs = item.UpdateTs.In(time.UTC)
	}
	return
}
