package repo

import (
	"context"

	"authenticator/pkg/postgres"
)

// TxRepo -.
type TxRepo struct {
	*postgres.Postgres
}

// NewTx -.
func NewTx(pg *postgres.Postgres) *TxRepo {
	return &TxRepo{pg}
}

func (t *TxRepo) NewTxId(ctx context.Context) (int, error) {
	id, err := t.PgTxBegin(ctx)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (t *TxRepo) TxEnd(ctx context.Context, txId int, err error) error {
	return t.PgTxEnd(ctx, txId, err)
}
