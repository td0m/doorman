package doorman

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func cleanup(ctx context.Context, conn *pgxpool.Pool) {
	query := `
		delete from tuples;
	`
	if _, err := conn.Exec(ctx, query); err != nil {
		panic(err)
	}
}

func TestDirect(t *testing.T) {
	ctx := context.Background()

	conn, err := pgxpool.New(ctx, "")
	if err != nil {
		panic(err)
	}
	cleanup(ctx, conn)

	s := NewServer(NewTupleStore(conn))

	assert.ErrorIs(t, s.Check(ctx, MustNewTuple("team:admins#member@dom")), ErrNoConnection)

	_, err = s.Write(ctx, WriteRequest{
		Add: []Tuple{MustNewTuple("team:admins#member@dom")},
	})
	require.NoError(t, err)

	assert.Nil(t, s.Check(ctx, MustNewTuple("team:admins#member@dom")))
}
