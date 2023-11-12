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

	assert.ErrorIs(t, s.Check(ctx, MustNewTuple("team:admins#member@user:dom")), ErrNoConnection)

	_, err = s.Write(ctx, WriteRequest{
		Add: []Tuple{MustNewTuple("team:admins#member@user:dom")},
	})
	require.NoError(t, err)

	assert.Nil(t, s.Check(ctx, MustNewTuple("team:admins#member@user:dom")))
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

	assert.ErrorIs(t, s.Check(ctx, MustNewTuple("item:banana#can_retrieve@user:dom")), ErrNoConnection)

	_, err = s.Write(ctx, WriteRequest{
		Add: []Tuple{MustNewTuple("item:banana#owner@user:dom")},
	})
	require.NoError(t, err)

	assert.Nil(t, s.Check(ctx, MustNewTuple("item:banana#can_retrieve@user:dom")))
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
		assert.ErrorIs(t, s.Check(ctx, MustNewTuple("item:banana#can_retrieve@user:dom")), ErrNoConnection)
	})

	t.Run("SuccessWhenOwner", func(t *testing.T) {
		_, err = s.Write(ctx, WriteRequest{
			Add: []Tuple{MustNewTuple("item:banana#owner@user:dom")},
		})
		require.NoError(t, err)

		assert.Nil(t, s.Check(ctx, MustNewTuple("item:banana#can_retrieve@user:dom")))
	})

	t.Run("SuccessWhenOwnerAndReader", func(t *testing.T) {
		_, err = s.Write(ctx, WriteRequest{
			Add: []Tuple{MustNewTuple("item:banana#reader@user:dom")},
		})
		require.NoError(t, err)

		assert.Nil(t, s.Check(ctx, MustNewTuple("item:banana#can_retrieve@user:dom")))
	})

	t.Run("SuccessWhenReader", func(t *testing.T) {
		_, err = s.Write(ctx, WriteRequest{
			Remove: []Tuple{MustNewTuple("item:banana#owner@user:dom")},
		})
		require.NoError(t, err)

		assert.Nil(t, s.Check(ctx, MustNewTuple("item:banana#can_retrieve@user:dom")))
	})

	t.Run("FailureWhenNeither", func(t *testing.T) {
		_, err = s.Write(ctx, WriteRequest{
			Remove: []Tuple{MustNewTuple("item:banana#reader@user:dom")},
		})
		require.NoError(t, err)

		assert.ErrorIs(t, s.Check(ctx, MustNewTuple("item:banana#can_retrieve@user:dom")), ErrNoConnection)
	})
}

func TestComputedViaTupleset(t *testing.T) {
	ctx := context.Background()

	conn, err := pgxpool.New(ctx, "")
	if err != nil {
		panic(err)
	}
	cleanup(ctx, conn)

	schema := SchemaDef{
		Types: []SchemaTypeDef{
			{
				Name: "shop",
				Relations: []SchemaRelationDef{
					{Name: "owner"},
				},
			},
			{
				Name: "item",
				Relations: []SchemaRelationDef{
					{Name: "seller"},
					{Name: "can_change_price", Value: NewComputedVia("seller", "owner")},
				},
			},
		},
	}

	s := NewServer(schema, NewTupleStore(conn))

	t.Run("FailsAtStart", func(t *testing.T) {
		assert.ErrorIs(t, s.Check(ctx, MustNewTuple("item:banana#can_change_price@user:dom")), ErrNoConnection)
	})

	t.Run("FailsOnUnrelated", func(t *testing.T) {
		_, err := s.Write(ctx, WriteRequest{
			Add: []Tuple{
				MustNewTuple("item:banana#seller@shop:asda"),
				MustNewTuple("item:banana#seller@shop:wallmart"),
				MustNewTuple("item:banana#seller@shop:lidl"),
			},
		})
		require.NoError(t, err)
		assert.ErrorIs(t, s.Check(ctx, MustNewTuple("item:banana#can_change_price@user:dom")), ErrNoConnection)
	})

	t.Run("Success", func(t *testing.T) {
		_, err := s.Write(ctx, WriteRequest{
			Add: []Tuple{
				MustNewTuple("item:banana#seller@shop:asda"),
				MustNewTuple("shop:asda#owner@user:dom"),
			},
		})
		require.NoError(t, err)
		assert.Nil(t, s.Check(ctx, MustNewTuple("item:banana#can_change_price@user:dom")))
	})
}

func TestStatic(t *testing.T) {
	ctx := context.Background()

	conn, err := pgxpool.New(ctx, "")
	if err != nil {
		panic(err)
	}
	cleanup(ctx, conn)

	schema := SchemaDef{
		Types: []SchemaTypeDef{
			{
				Name: "group",
				Relations: []SchemaRelationDef{
					{Name: "member"},
				},
			},
			{
				Name: "item",
				Relations: []SchemaRelationDef{
					{Name: "can_change_price", Value: NewStatic(MustNewTupleset("group:asda#member"))},
				},
			},
		},
	}

	s := NewServer(schema, NewTupleStore(conn))

	t.Run("FailsAtStart", func(t *testing.T) {
		assert.ErrorIs(t, s.Check(ctx, MustNewTuple("item:banana#can_change_price@user:dom")), ErrNoConnection)
	})

	t.Run("Success", func(t *testing.T) {
		_, err := s.Write(ctx, WriteRequest{
			Add: []Tuple{
				MustNewTuple("group:asda#member@user:dom"),
			},
		})
		require.NoError(t, err)
		assert.Nil(t, s.Check(ctx, MustNewTuple("item:banana#can_change_price@user:dom")))
	})
}
