package doorman

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TupleStore struct {
	conn *pgxpool.Pool
}

func (t *TupleStore) Add(ctx context.Context, tuple Tuple) error {
	query := `
		insert into tuples(object, relation, user_id)
		values($1, $2, $3)
		on conflict do nothing
	`

	if _, err := t.conn.Exec(ctx, query, tuple.Object, tuple.Relation, tuple.UserID); err != nil {
		return fmt.Errorf("exec failed: %w", err)
	}

	return nil
}

func (t *TupleStore) Exists(ctx context.Context, tuple Tuple) error {
	query := `
		select exists(
			select 1
			from tuples
			where
				object = $1 and
				relation = $2 and
				user_id = $3
		)
	`

	var exists bool
	if err := t.conn.QueryRow(ctx, query, tuple.Object, tuple.Relation, tuple.UserID).Scan(&exists); err != nil {
		return fmt.Errorf("query failed: %w", err)
	}

	if !exists {
		return ErrNoConnection
	}
	return nil
}

func (t *TupleStore) Remove(ctx context.Context, tuple Tuple) error {
	query := `
		delete from tuples
		where
			object = $1 and
			relation = $2 and
			user_id = $3
	`

	if _, err := t.conn.Exec(ctx, query, tuple.Object, tuple.Relation, tuple.UserID); err != nil {
		return fmt.Errorf("exec failed: %w", err)
	}

	return nil
}

func NewTupleStore(c *pgxpool.Pool) *TupleStore {
	return &TupleStore{c}
}
