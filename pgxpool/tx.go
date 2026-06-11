package pgxpool

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type connTx struct {
	conn *pgconn.PgConn
	tx   pgx.Tx
}

func (ct *connTx) Begin(ctx context.Context) error {
	if ct.conn == nil {
		return nil
	}
	return ct.conn.Begin()
}

func (ct *connTx) Commit(ctx context.Context) error {
	if ct.conn == nil {
		return nil
	}
	err := ct.conn.Commit()
	ct.conn.Release()
	ct.conn = nil
	return err
}

func (ct *connTx) Rollback(ctx context.Context) error {
	if ct.conn == nil {
		return nil
	}
	err := ct.conn.Rollback()
	ct.conn.Release()
	ct.conn = nil
	return err
}