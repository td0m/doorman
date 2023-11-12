package doorman

import (
	"context"
	"fmt"
	"reflect"
)

const concurrently = true

type ComputedTupleResolver struct {
	schema Schema
	server *Server
}

func (r *ComputedTupleResolver) Check(ctx context.Context, tuple Tuple) (bool, error) {
	expr, err := r.schema.GetExpr(ctx, Tupleset{Object: tuple.Object, Relation: tuple.Relation})
	if err != nil {
		return false, fmt.Errorf("resolve failed: %w", err)
	}

	if expr == nil {
		return false, nil
	}

	return r.check(ctx, expr, tuple)
}

func (r *ComputedTupleResolver) check(ctx context.Context, expr Expr, tuple Tuple) (bool, error) {
	switch expr := expr.(type) {
	case Union:
		for _, e := range expr {
			success, err := r.check(ctx, e, tuple)
			if err != nil {
				return false, err
			}

			if success {
				return true, nil
			}
		}
		return false, nil

	case Tupleset:
		return r.server.Check(ctx, Tuple{Object: expr.Object, Relation: expr.Relation, Subject: tuple.Subject})

	case Sibling:
		return r.server.Check(ctx, Tuple{Object: tuple.Object, Relation: string(expr), Subject: tuple.Subject})

	case ViaSibling:
		subs, err := r.server.ListSubjects(ctx, Tupleset{Object: tuple.Object, Relation: expr.SiblingRelation})
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
					tuple := Tuple{Object: obj, Relation: expr.Relation, Subject: tuple.Subject}
					yes, err := r.server.Check(ctx, tuple)
					if err != nil {
						jobs <- err
						return
					}
					if yes {
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
				tuple := Tuple{Object: obj, Relation: expr.Relation, Subject: tuple.Subject}
				yes, err := r.server.Check(ctx, tuple)
				if err != nil {
					return false, err
				}
				if yes {
					return true, nil
				}
			}
		}
		return false, nil
	case NilSet:
		return false, nil
	default:
		panic("unhandled type: " + reflect.TypeOf(expr).String())
	}
}

//
// type OrderedTuplesets map[Tupleset]bool
//
// func (m OrderedTuplesets) Add(tupleset Tupleset) {
// 	m[tupleset] = true
// }
//
// func (m OrderedTuplesets) Remove(tupleset Tupleset) {
// 	delete(m, tupleset)
// }
//
// func (a OrderedTuplesets) Intersects(b OrderedTuplesets) bool {
// 	for a, aconn := range a {
// 		if aconn && b[a] {
// 			return true
// 		}
// 	}
// 	return false
// }
//
// type CachedResolver struct {
// 	resolver          UsersetResolver
// 	subject2tupleset  map[string]OrderedTuplesets
// 	tupleset2tupleset map[Tupleset]OrderedTuplesets
// }
//
// func NewCached(resolver UsersetResolver) CachedResolver {
// 	return CachedResolver{
// 		resolver:          resolver,
// 		subject2tupleset:  map[string]OrderedTuplesets{},
// 		tupleset2tupleset: map[Tupleset]OrderedTuplesets{},
// 	}
// }
//
// func (r CachedResolver) Check(ctx context.Context, tupleset Tupleset, userset Userset, subject string) (bool, error) {
// 	subject2tupleset, ok1 := r.subject2tupleset[subject]
// 	tupleset2tupleset, ok2 := r.tupleset2tupleset[tupleset]
//
// 	switch s := userset.(type) {
// 	case DirectUserset:
// 		return r.resolver.Check(ctx, userset, subject)
//
// 	case UsersetUnion:
// 		return r.resolver.Check(ctx, userset, subject)
//
// 	case ComputedUserset:
// 		return r.resolver.Check(ctx, userset, subject)
//
// 	case ComputedUsersetViaTupleset:
// 		return r.resolver.Check(ctx, userset, subject)
//
// 	default:
// 		panic("unhandled type" + s.String())
// 	}
// }
