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

	success, err := s.Check(ctx, MustNewTuple("team:admins#member@user:dom"))
	require.NoError(t, err)
	assert.False(t, success)

	_, err = s.Write(ctx, WriteRequest{
		Add: []Tuple{MustNewTuple("team:admins#member@user:dom")},
	})
	require.NoError(t, err)

	success, err = s.Check(ctx, MustNewTuple("team:admins#member@user:dom"))
	require.NoError(t, err)
	assert.True(t, success)
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

	success, err := s.Check(ctx, MustNewTuple("item:banana#can_retrieve@user:dom"))
	require.NoError(t, err)
	assert.False(t, success)

	_, err = s.Write(ctx, WriteRequest{
		Add: []Tuple{MustNewTuple("item:banana#owner@user:dom")},
	})
	require.NoError(t, err)

	success, err = s.Check(ctx, MustNewTuple("item:banana#can_retrieve@user:dom"))
	require.NoError(t, err)
	assert.True(t, success)
}

func TestComputedButSubjectIsTupleset(t *testing.T) {
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
					{Name: "child"},
					{Name: "member", Value: NewComputedVia("child", "member")},
				},
			},
			{
				Name: "shop",
				Relations: []SchemaRelationDef{
					{Name: "owner"},
				},
			},
		},
	}

	s := NewServer(schema, NewTupleStore(conn))

	t.Run("FailsInitially", func(t *testing.T) {
		success, err := s.Check(ctx, MustNewTuple("shop:a#owner@user:dom"))
		require.NoError(t, err)
		assert.False(t, success)
	})

	t.Run("SuccessWhenGroupMember", func(t *testing.T) {
		_, err = s.Write(ctx, WriteRequest{
			Add: []Tuple{
				MustNewTuple("shop:a#owner@group:foo#member"),
				MustNewTuple("group:foo#member@user:dom"),
			},
		})
		require.NoError(t, err)

		success, err := s.Check(ctx, MustNewTuple("shop:a#owner@user:dom"))
		require.NoError(t, err)
		assert.True(t, success)
	})

	t.Run("SuccessCheckingTuplesetMemberOfTupleset", func(t *testing.T) {
		success, err := s.Check(ctx, MustNewTuple("shop:a#owner@group:foo#member"))
		require.NoError(t, err)
		assert.True(t, success)
	})
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
					{Name: "can_write", Value: NewComputed("reader")},
					{Name: "can_retrieve", Value: NewUnion(NewComputed("owner"), NewComputed("can_write"))},
				},
			},
		},
	}

	s := NewServer(schema, NewTupleStore(conn))

	t.Run("FailsAtStart", func(t *testing.T) {
		success, err := s.Check(ctx, MustNewTuple("item:banana#can_retrieve@user:dom"))
		require.NoError(t, err)
		assert.False(t, success)
	})

	t.Run("SuccessWhenOwner", func(t *testing.T) {
		_, err = s.Write(ctx, WriteRequest{
			Add: []Tuple{MustNewTuple("item:banana#owner@user:dom")},
		})
		require.NoError(t, err)

		success, err := s.Check(ctx, MustNewTuple("item:banana#can_retrieve@user:dom"))
		require.NoError(t, err)
		assert.True(t, success)
	})

	t.Run("SuccessWhenOwnerAndReader", func(t *testing.T) {
		_, err = s.Write(ctx, WriteRequest{
			Add: []Tuple{MustNewTuple("item:banana#reader@user:dom")},
		})
		require.NoError(t, err)

		success, err := s.Check(ctx, MustNewTuple("item:banana#can_retrieve@user:dom"))
		require.NoError(t, err)
		assert.True(t, success)
	})

	t.Run("SuccessWhenReader", func(t *testing.T) {
		_, err = s.Write(ctx, WriteRequest{
			Remove: []Tuple{MustNewTuple("item:banana#owner@user:dom")},
		})
		require.NoError(t, err)

		success, err := s.Check(ctx, MustNewTuple("item:banana#can_retrieve@user:dom"))
		require.NoError(t, err)
		assert.True(t, success)
	})

	t.Run("FailureWhenNeither", func(t *testing.T) {
		_, err = s.Write(ctx, WriteRequest{
			Remove: []Tuple{MustNewTuple("item:banana#reader@user:dom")},
		})
		require.NoError(t, err)

		success, err := s.Check(ctx, MustNewTuple("item:banana#can_retrieve@user:dom"))
		require.NoError(t, err)
		assert.False(t, success)
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
		success, err := s.Check(ctx, MustNewTuple("item:banana#can_change_price@user:dom"))
		require.NoError(t, err)
		assert.False(t, success)
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
		success, err := s.Check(ctx, MustNewTuple("item:banana#can_change_price@user:dom"))
		require.NoError(t, err)
		assert.False(t, success)
	})

	t.Run("Success", func(t *testing.T) {
		_, err := s.Write(ctx, WriteRequest{
			Add: []Tuple{
				MustNewTuple("item:banana#seller@shop:asda"),
				MustNewTuple("shop:asda#owner@user:dom"),
			},
		})
		require.NoError(t, err)
		success, err := s.Check(ctx, MustNewTuple("item:banana#can_change_price@user:dom"))
		require.NoError(t, err)
		assert.True(t, success)

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
		success, err := s.Check(ctx, MustNewTuple("item:banana#can_change_price@user:dom"))
		require.NoError(t, err)
		assert.False(t, success)
	})

	t.Run("Success", func(t *testing.T) {
		_, err := s.Write(ctx, WriteRequest{
			Add: []Tuple{
				MustNewTuple("group:asda#member@user:dom"),
			},
		})
		require.NoError(t, err)
		success, err := s.Check(ctx, MustNewTuple("item:banana#can_change_price@user:dom"))
		require.NoError(t, err)
		assert.True(t, success)
	})
}
