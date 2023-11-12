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

// func (s Server) CheckDirect(ctx context.Context, tuple Tuple) error {
// 	// Check direct connection
// 	ok, err := s.tuples.Exists(ctx, tuple)
// 	if err != nil {
// 		return false, nil
// 	}
// 	return ErrNoConnection
// }

func (s *Server) Check(ctx context.Context, tuple Tuple) (bool, error) {
	success, err := s.tuples.Exists(ctx, tuple)
	if err != nil {
		return false, fmt.Errorf("failed to read tuples in db: %w", err)
	}
	if success {
		return true, nil
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
	// // Check Direct
	// err := s.CheckDirect(ctx, tuple)
	// if err != nil && err != ErrNoConnection {
	// 	return fmt.Errorf("failed direct check: %w", err)
	// }
	// if err == nil {
	// 	return nil
	// }
	//
	// lazyUserset, err := s.schema.Resolve(ctx, tuple)
	// if err != nil {
	// 	return fmt.Errorf("failed resolving schema: %w", err)
	// }
	//
	// if false {
	// 	fmt.Println(tuple, lazyUserset)
	// }
	//
	// resolver := DirectResolver{
	// 	server: s,
	// }
	// // cached := NewCached(resolver)
	//
	// ok, err := resolver.Check(ctx, lazyUserset, tuple.Subject)
	// if err != nil {
	// 	return fmt.Errorf("failed computing lazy userset: %w", err)
	// }
	//
	// if ok {
	// 	return nil
	// }
	// return ErrNoConnection
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
