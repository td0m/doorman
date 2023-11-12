package doorman

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TupleStore struct {
	conn *pgxpool.Pool
}

func (t *TupleStore) ListSubjects(ctx context.Context, userset Userset) ([]string, error) {
	query := `
		select subject
		from tuples
		where
			object = $1 and
			relation = $2
	`

	users := []string{}
	rows, err := t.conn.Query(ctx, query, userset.Object, userset.Relation)
	if err != nil {
		return nil, fmt.Errorf("failed to query: %w", err)
	}

	for rows.Next() {
		user := ""
		if err := rows.Scan(ctx, &user); err != nil {
			return nil, fmt.Errorf("failed to scan: %w", err)
		}
		users = append(users, user)
	}

	return users, nil
}

func (t *TupleStore) Add(ctx context.Context, tuple Tuple) error {
	query := `
		insert into tuples(object, relation, subject)
		values($1, $2, $3)
		on conflict do nothing
	`

	if _, err := t.conn.Exec(ctx, query, tuple.Object, tuple.Relation, tuple.Subject); err != nil {
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
				subject = $3
		)
	`

	var exists bool
	if err := t.conn.QueryRow(ctx, query, tuple.Object, tuple.Relation, tuple.Subject).Scan(&exists); err != nil {
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
			subject = $3
	`

	if _, err := t.conn.Exec(ctx, query, tuple.Object, tuple.Relation, tuple.Subject); err != nil {
		return fmt.Errorf("exec failed: %w", err)
	}

	return nil
}

func NewTupleStore(c *pgxpool.Pool) *TupleStore {
	return &TupleStore{c}
}
