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

	schema := SchemaDef{
		Types: []SchemaTypeDef{
			{
				Name: "team",
				Relations: []SchemaRelationDef{
					{Name: "member"},
				},
			},
		},
	}

	s := NewServer(schema, NewTupleStore(conn))

	assert.ErrorIs(t, s.Check(ctx, MustNewTuple("team:admins#member@dom")), ErrNoConnection)

	_, err = s.Write(ctx, WriteRequest{
		Add: []Tuple{MustNewTuple("team:admins#member@dom")},
	})
	require.NoError(t, err)

	assert.Nil(t, s.Check(ctx, MustNewTuple("team:admins#member@dom")))
}

func TestComputed(t *testing.T) {
	ctx := context.Background()

	conn, err := pgxpool.New(ctx, "")
	if err != nil {
		panic(err)
	}
	cleanup(ctx, conn)

	schema := SchemaDef{
		Types: []SchemaTypeDef{
			{
				Name: "item",
				Relations: []SchemaRelationDef{
					{Name: "owner"},
					{Name: "can_retrieve", Value: NewComputed("owner")},
				},
			},
		},
	}

	s := NewServer(schema, NewTupleStore(conn))

	assert.ErrorIs(t, s.Check(ctx, MustNewTuple("item:banana#can_retrieve@dom")), ErrNoConnection)

	_, err = s.Write(ctx, WriteRequest{
		Add: []Tuple{MustNewTuple("item:banana#owner@dom")},
	})
	require.NoError(t, err)

	assert.Nil(t, s.Check(ctx, MustNewTuple("item:banana#can_retrieve@dom")))
}

func TestUnion(t *testing.T) {
	ctx := context.Background()

	conn, err := pgxpool.New(ctx, "")
	if err != nil {
		panic(err)
	}
	cleanup(ctx, conn)

	schema := SchemaDef{
		Types: []SchemaTypeDef{
			{
				Name: "item",
				Relations: []SchemaRelationDef{
					{Name: "owner"},
					{Name: "reader"},
					{Name: "can_retrieve", Value: NewUnion(NewComputed("reader"), NewComputed("owner"))},
				},
			},
		},
	}

	s := NewServer(schema, NewTupleStore(conn))

	t.Run("FailsAtStart", func(t *testing.T) {
		assert.ErrorIs(t, s.Check(ctx, MustNewTuple("item:banana#can_retrieve@dom")), ErrNoConnection)
	})

	t.Run("SuccessWhenOwner", func(t *testing.T) {
		_, err = s.Write(ctx, WriteRequest{
			Add: []Tuple{MustNewTuple("item:banana#owner@dom")},
		})
		require.NoError(t, err)

		assert.Nil(t, s.Check(ctx, MustNewTuple("item:banana#can_retrieve@dom")))
	})

	t.Run("SuccessWhenOwnerAndReader", func(t *testing.T) {
		_, err = s.Write(ctx, WriteRequest{
			Add: []Tuple{MustNewTuple("item:banana#reader@dom")},
		})
		require.NoError(t, err)

		assert.Nil(t, s.Check(ctx, MustNewTuple("item:banana#can_retrieve@dom")))
	})

	t.Run("SuccessWhenReader", func(t *testing.T) {
		_, err = s.Write(ctx, WriteRequest{
			Remove: []Tuple{MustNewTuple("item:banana#owner@dom")},
		})
		require.NoError(t, err)

		assert.Nil(t, s.Check(ctx, MustNewTuple("item:banana#can_retrieve@dom")))
	})

	t.Run("FailureWhenNeither", func(t *testing.T) {
		_, err = s.Write(ctx, WriteRequest{
			Remove: []Tuple{MustNewTuple("item:banana#reader@dom")},
		})
		require.NoError(t, err)

		assert.ErrorIs(t, s.Check(ctx, MustNewTuple("item:banana#can_retrieve@dom")), ErrNoConnection)
	})
}
