package doorman

import (
	"context"
	"fmt"
)

const concurrently = true

type UsersetResolver interface {
	Check(ctx context.Context, userset Userset, subject string) (bool, error)
}

type DirectResolver struct {
	server *Server
}

func (r DirectResolver) Check(ctx context.Context, userset Userset, subject string) (bool, error) {
	switch s := userset.(type) {

	case DirectUserset:
		err := r.server.CheckDirect(ctx, Tuple{Object: s.Tupleset.Object, Relation: s.Tupleset.Relation, Subject: subject})
		if err != nil && err != ErrNoConnection {
			return false, fmt.Errorf("failed direct check: %w", err)
		}
		return err != ErrNoConnection, nil

	case UsersetUnion:
		for _, arg := range s.Args {
			has, err := r.Check(ctx, arg, subject)
			if err != nil {
				return false, fmt.Errorf("has failed: %w", err)
			}
			if has {
				return true, nil
			}
		}
		return false, nil

	case ComputedUserset:
		err := r.server.Check(ctx, Tuple{Object: s.Tupleset.Object, Relation: s.Tupleset.Relation, Subject: subject})
		if err != nil && err != ErrNoConnection {
			return false, err
		}
		return err != ErrNoConnection, nil

	case ComputedUsersetViaTupleset:
		subs, err := r.server.ListSubjects(ctx, s.Tupleset)
		if err != nil {
			return false, fmt.Errorf("failed to list subjects: %w", err)
		}

		if len(subs) == 0 {
			return false, nil
		}

		if concurrently {
			jobs := make(chan error)
			success := make(chan bool, 1)
			for _, obj := range subs {
				obj := obj
				go func() {
					tuple := Tuple{Object: obj, Relation: s.UsersetRelation, Subject: subject}
					err := r.server.Check(ctx, tuple)
					if err != nil && err != ErrNoConnection {
						jobs <- err
						return
					}
					if err == nil {
						success <- true
					}
					jobs <- nil
				}()
			}
			for i := 0; i < len(subs); i++ {
				select {
				case err := <-jobs:
					if err != nil {
						return false, err
					}
				case <-success:
					return true, nil
				}
			}
		} else {
			for _, obj := range subs {
				tuple := Tuple{Object: obj, Relation: s.UsersetRelation, Subject: subject}
				err := r.server.Check(ctx, tuple)
				if err != nil && err != ErrNoConnection {
					return false, err
				}
				if err == nil {
					return true, nil
				}
			}
		}
		return false, nil

	default:
		panic("unhandled type")
	}
}
