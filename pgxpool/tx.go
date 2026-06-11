package pgxpool

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	realpgxpool "github.com/jackc/pgx/v5/pgxpool"
)

type connTx struct {
	tx   pgx.Tx
	conn *realpgxpool.Conn
}

func (ct *connTx) Begin(ctx context.Context) (pgx.Tx, error) {
	return ct.tx.Begin(ctx)
}

func (ct *connTx) Commit(ctx context.Context) error {
	if ct.conn == nil {
		return pgx.ErrTxClosed
	}

	err := ct.tx.Commit(ctx)
	ct.conn.Release()
	ct.conn = nil
	return err
}

func (ct *connTx) Rollback(ctx context.Context) error {
	if ct.conn == nil {
		return pgx.ErrTxClosed
	}

	err := ct.tx.Rollback(ctx)
	if err != nil {
		ct.conn.Conn().Close(context.Background())
		ct.conn.Release()
		ct.conn = nil
		return err
	}

	ct.conn.Release()
	ct.conn = nil
	return nil
}

func (ct *connTx) CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error) {
	return ct.tx.CopyFrom(ctx, tableName, columnNames, rowSrc)
}

func (ct *connTx) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults {
	return ct.tx.SendBatch(ctx, b)
}

func (ct *connTx) LargeObjects() pgx.LargeObjects {
	return ct.tx.LargeObjects()
}

func (ct *connTx) Prepare(ctx context.Context, name, sql string) (*pgconn.StatementDescription, error) {
	return ct.tx.Prepare(ctx, name, sql)
}

func (ct *connTx) Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error) {
	return ct.tx.Exec(ctx, sql, arguments...)
}

func (ct *connTx) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	return ct.tx.Query(ctx, sql, args...)
}

func (ct *connTx) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	return ct.tx.QueryRow(ctx, sql, args...)
}

func (ct *connTx) Conn() *pgx.Conn {
	return ct.conn.Conn()
}
