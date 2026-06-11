package pgxpool_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTxCommit(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	if os.Getenv("PGX_TEST_DATABASE") == "" {
		t.Skip("Skipping test because PGX_TEST_DATABASE is not set")
	}

	pool, err := pgxpool.New(ctx, os.Getenv("PGX_TEST_DATABASE"))
	require.NoError(t, err)
	defer pool.Close()

	tx, err := pool.Begin(ctx)
	require.NoError(t, err)

	_, err = tx.Exec(ctx, "create temporary table foo(id integer)")
	require.NoError(t, err)

	err = tx.Commit(ctx)
	require.NoError(t, err)

	conn, err := pool.Acquire(ctx)
	require.NoError(t, err)
	defer conn.Release()

	var exists bool
	err = conn.QueryRow(ctx, "select exists (select 1 from pg_class where relname = 'foo')").Scan(&exists)
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestTxRollback(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	if os.Getenv("PGX_TEST_DATABASE") == "" {
		t.Skip("Skipping test because PGX_TEST_DATABASE is not set")
	}

	pool, err := pgxpool.New(ctx, os.Getenv("PGX_TEST_DATABASE"))
	require.NoError(t, err)
	defer pool.Close()

	tx, err := pool.Begin(ctx)
	require.NoError(t, err)

	_, err = tx.Exec(ctx, "create temporary table foo(id integer)")
	require.NoError(t, err)

	err = tx.Rollback(ctx)
	require.NoError(t, err)

	conn, err := pool.Acquire(ctx)
	require.NoError(t, err)
	defer conn.Release()

	var exists bool
	err = conn.QueryRow(ctx, "select exists (select 1 from pg_class where relname = 'foo')").Scan(&exists)
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestTxRollbackFailureDestroy(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if os.Getenv("PGX_TEST_DATABASE") == "" {
		t.Skip("Skipping test because PGX_TEST_DATABASE is not set")
	}

	config, err := pgxpool.ParseConfig(os.Getenv("PGX_TEST_DATABASE"))
	require.NoError(t, err)
	config.MaxConns = 1

	pool, err := pgxpool.NewWithConfig(ctx, config)
	require.NoError(t, err)
	defer pool.Close()

	tx, err := pool.Begin(ctx)
	require.NoError(t, err)

	canceledCtx, cancelRollback := context.WithCancel(ctx)
	cancelRollback()

	err = tx.Rollback(canceledCtx)
	require.Error(t, err)

	// Try to acquire a new connection with a short timeout.
	// If the connection was leaked, this will fail/timeout.
	acquireCtx, cancelAcquire := context.WithTimeout(ctx, 1*time.Second)
	defer cancelAcquire()

	conn, err := pool.Acquire(acquireCtx)
	require.NoError(t, err)
	conn.Release()

	// Subsequent calls to Rollback or Commit must return ErrTxClosed
	err = tx.Rollback(ctx)
	require.ErrorIs(t, err, pgx.ErrTxClosed)

	err = tx.Commit(ctx)
	require.ErrorIs(t, err, pgx.ErrTxClosed)
}
