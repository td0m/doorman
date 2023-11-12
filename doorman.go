package doorman

import (
	"context"
	"errors"
	"fmt"
)

var ErrNoConnection = errors.New("no connection")

type WriteRequest struct {
	Add    []Tuple
	Remove []Tuple
}

type Server struct {
	schema SchemaDef
	tuples *TupleStore
}

func (s *Server) Check(ctx context.Context, tuple Tuple) (bool, error) {
	success, err := s.tuples.Exists(ctx, tuple)
	if err != nil {
		return false, fmt.Errorf("failed to read tuples in db: %w", err)
	}
	if success {
		return true, nil
	}

	// If connected to subject that is a tupleset then any members of that tupleset automatically become members of this one
	subjects, _ := s.tuples.ListSubjects(ctx, Tupleset{Object: tuple.Object, Relation: tuple.Relation})
	for _, subject := range subjects {
		tupleset, err := NewTupleset(subject)
		// TODO: consider more efficient way of doing this by having materialized col and indexing by it.
		if err != nil {
			continue
		}

		success, err := s.Check(ctx, Tuple{Object: tupleset.Object, Relation: tupleset.Relation, Subject: tuple.Subject})
		if err != nil {
			return false, fmt.Errorf("failed to check tupleset: %w", err)
		}
		if success {
			return true, nil
		}
	}

	computed := ComputedTupleResolver{server: s, schema: s.schema}
	success, err = computed.Check(ctx, tuple)
	if err != nil {
		return false, err
	}
	if success {
		return true, nil
	}
	return false, nil
}

func (s *Server) ListSubjects(ctx context.Context, tupleset Tupleset) ([]string, error) {
	return s.tuples.ListSubjects(ctx, tupleset)
}

func (s *Server) Write(ctx context.Context, request WriteRequest) (any, error) {
	for _, tuple := range request.Add {
		if err := s.tuples.Add(ctx, tuple); err != nil {
			return nil, fmt.Errorf("failed to add tuple '%s': %w", tuple, err)
		}
	}

	for _, tuple := range request.Remove {
		if err := s.tuples.Remove(ctx, tuple); err != nil {
			return nil, fmt.Errorf("failed to remove tuple '%s': %w", tuple, err)
		}
	}

	return nil, nil
}

func NewServer(schema SchemaDef, ts *TupleStore) *Server {
	return &Server{schema: schema, tuples: ts}
}
