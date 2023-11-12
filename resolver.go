package doorman

import (
	"context"
	"fmt"
	"reflect"
)

const concurrently = true

type ComputedTupleResolver struct {
	schema SchemaDef
	server *Server
}

func (r *ComputedTupleResolver) Check(ctx context.Context, tuple Tuple) (bool, error) {
	setdef, err := r.schema.GetDef(ctx, Tupleset{Object: tuple.Object, Relation: tuple.Relation})
	if err != nil {
		return false, fmt.Errorf("resolve failed: %w", err)
	}

	if setdef == nil {
		return false, nil
	}

	return r.check(ctx, setdef, tuple)
}

func (r *ComputedTupleResolver) check(ctx context.Context, setdef SetDef, tuple Tuple) (bool, error) {
	userset, err := setdef.ToUserset(ctx, tuple.Object)
	if err != nil {
		return false, fmt.Errorf("failed converting to userset: %w", err)
	}
	switch s := userset.(type) {
	case UsersetUnion:
		for _, def := range s.Args {
			success, err := r.check(ctx, def, tuple)
			if err != nil {
				return false, err
			}

			if success {
				return true, nil
			}
		}
		return false, nil

	case ComputedUserset:
		return r.server.Check(ctx, Tuple{Object: s.Tupleset.Object, Relation: s.Tupleset.Relation, Subject: tuple.Subject})

	case StaticTupleset:
		return r.server.Check(ctx, Tuple{Object: s.Tupleset.Object, Relation: s.Tupleset.Relation, Subject: tuple.Subject})

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
					tuple := Tuple{Object: obj, Relation: s.UsersetRelation, Subject: tuple.Subject}
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
				tuple := Tuple{Object: obj, Relation: s.UsersetRelation, Subject: tuple.Subject}
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
		panic("unhandled type: " + reflect.TypeOf(s).String())
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
