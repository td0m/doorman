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

func (s Server) CheckDirect(ctx context.Context, tuple Tuple) error {
	// Check direct connection
	err := s.tuples.Exists(ctx, tuple)
	if err == nil {
		return nil
	}
	if err != ErrNoConnection {
		return fmt.Errorf("failed checking tuple store: %w", err)
	}

	return ErrNoConnection
}

func (s *Server) Check(ctx context.Context, tuple Tuple) error {
	lazyUserset, err := s.schema.Resolve(ctx, s.tuples, tuple)
	if err != nil {
		return fmt.Errorf("failed resolving schema: %w", err)
	}

	ok, err := lazyUserset.Has(ctx, s, tuple.UserID)
	if err != nil {
		return fmt.Errorf("failed computing lazy userset: %w", err)
	}

	if ok {
		return nil
	}

	return ErrNoConnection
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
