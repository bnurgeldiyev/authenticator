package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type Tx interface {
	Exec(ctx context.Context, query string, args ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, query string, args ...interface{}) pgx.Row
}

type connTx struct {
	conn *pgxpool.Conn
	Tx   pgx.Tx
}

func (t *connTx) Exec(ctx context.Context, query string, args ...interface{}) (pgconn.CommandTag, error) {
	return t.Tx.Exec(ctx, query, args...)
}

func (t *connTx) Query(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error) {
	return t.Tx.Query(ctx, query, args...)
}

func (t *connTx) QueryRow(ctx context.Context, query string, args ...interface{}) pgx.Row {
	return t.Tx.QueryRow(ctx, query, args...)
}

func (p *Postgres) PgTxBegin(ctx context.Context) (id int, err error) {
	var conn *pgxpool.Conn
	conn, err = p.Pool.Acquire(ctx)
	if err != nil {
		log.Error().Err(err).Msg("error acquiring connection")
		return
	}
	var tx pgx.Tx
	tx, err = conn.Begin(ctx)
	if err != nil {
		log.Error().Err(err).Msg("error beginning tx")
		return
	}
	id = p.getTxIdAddMap(&connTx{
		conn: conn,
		Tx:   tx,
	})
	return id, nil
}

func (p *Postgres) PgTxEnd(ctx context.Context, txId int, errInTx error) error {
	var err error
	var pgTx *connTx
	pgTx, err = p.getConnTxById(txId)
	if err != nil {
		return err
	}
	tx := pgTx.Tx

	defer func() {
		if errInTx != nil && tx != nil {
			log.Debug().Msg("found error and rollback")
			rErr := tx.Rollback(ctx) // err is non-nil; don't change it
			if rErr != nil && err != pgx.ErrTxClosed {
				log.Error().Err(rErr).Msg("error at tx.Rollback")
			}
		}
		if !tx.Conn().IsClosed() {
			cErr := tx.Conn().Close(ctx)
			if cErr != nil {
				log.Error().Err(cErr).Msg("error at tx.Conn().Close")
			}
		}
		if pgTx.conn != nil {
			pgTx.conn.Release()
		}
		p.deleteTxFromMap(txId)
	}()

	if errInTx == nil {
		err = tx.Commit(ctx) // if Commit returns error update err with commit err
		if err != nil {
			log.Error().Err(err).Msg("error at tx.Commit")
			return err
		}
	}
	return err
}

func (p *Postgres) GetTxById(id int) (Tx, error) {
	p.txMutex.RLock()
	defer p.txMutex.RUnlock()
	if tx, ok := p.txMap[id]; ok {
		return tx, nil
	}
	return nil, errors.New("pgTx not found")
}

func (p *Postgres) getConnTxById(id int) (*connTx, error) {
	p.txMutex.RLock()
	defer p.txMutex.RUnlock()
	if tx, ok := p.txMap[id]; ok {
		return tx, nil
	}
	return nil, errors.New("pgTx not found")
}

func (p *Postgres) getTxIdAddMap(pgTx *connTx) int {
	p.txMutex.Lock()
	defer p.txMutex.Unlock()
	p.idTx += 1
	if len(p.txMap) == 0 {
		p.idTx = 0
	}
	p.txMap[p.idTx] = pgTx
	return p.idTx
}

func (p *Postgres) deleteTxFromMap(id int) {
	p.txMutex.Lock()
	defer p.txMutex.Unlock()
	delete(p.txMap, id)
}
