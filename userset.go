package doorman

import (
	"context"
	"fmt"
)

type LazyResolver interface {
	Check(ctx context.Context, tuple Tuple) error
	CheckDirect(ctx context.Context, tuple Tuple) error
	ListSubjects(ctx context.Context, tupleset Tupleset) ([]string, error)
}

type Userset interface {
	String() string
}

type UsersetUnion struct {
	Args []SetDef
}

type NilSet struct{}

func (d NilSet) String() string {
	return "nil"
}

func (d UsersetUnion) String() string {
	return "union?"
	// argStrs := make([]string, len(d.Args))
	// for i, arg := range d.Args {
	// 	argStrs[i] = arg.String()
	// }
	// return fmt.Sprintf("(union %v)", strings.Join(argStrs, " "))
}

type ComputedUserset struct {
	Tupleset Tupleset
}

func (c ComputedUserset) String() string {
	return fmt.Sprintf("(computed %s)", c.Tupleset)
}

type StaticTupleset struct {
	Tupleset Tupleset
}

func (d StaticTupleset) String() string {
	return fmt.Sprintf("(static %s)", d.Tupleset)
}

type ComputedUsersetViaTupleset struct {
	Tupleset        Tupleset
	UsersetRelation string
}

func (d ComputedUsersetViaTupleset) String() string {
	return fmt.Sprintf("(computed_via %s#%s)", d.Tupleset, d.UsersetRelation)
}
