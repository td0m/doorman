package doorman

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

type Tupleset struct {
	Object   string
	Relation string
}

func (t Tupleset) String() string {
	return t.Object + "#" + t.Relation
}

func NewTupleset(s string) (Tupleset, error) {
	parts := strings.SplitN(s, "#", 2)
	if len(parts) != 2 {
		return Tupleset{}, errors.New("expected to contain exactly one '#'")
	}

	otype, _ := GetObjectTypeAndID(parts[0])
	if otype == "" {
		return Tupleset{}, errors.New("object must have a type")
	}
	return Tupleset{Object: parts[0], Relation: parts[1]}, nil
}

func MustNewTupleset(s string) Tupleset {
	userset, err := NewTupleset(s)
	if err != nil {
		panic(err)
	}

	return userset
}

type LazyResolver interface {
	Check(ctx context.Context, tuple Tuple) error
	CheckDirect(ctx context.Context, tuple Tuple) error
	ListSubjects(ctx context.Context, tupleset Tupleset) ([]string, error)
}

type LazyUserset interface {
	Has(ctx context.Context, resolver LazyResolver, sub string) (bool, error)
	String() string
}

type LazyUnionUserset struct {
	Args []LazyUserset
}

func (s LazyUnionUserset) Has(ctx context.Context, resolver LazyResolver, sub string) (bool, error) {
	for _, arg := range s.Args {
		has, err := arg.Has(ctx, resolver, sub)
		if err != nil {
			return false, fmt.Errorf("has failed: %w", err)
		}

		if has {
			return true, nil
		}
	}
	return false, nil
}

func (d LazyUnionUserset) String() string {
	argStrs := make([]string, len(d.Args))
	for i, arg := range d.Args {
		argStrs[i] = arg.String()
	}
	return fmt.Sprintf("(union %v)", strings.Join(argStrs, " "))
}

type ComputedUserset struct {
	Tupleset Tupleset
}

func (c ComputedUserset) Has(ctx context.Context, resolver LazyResolver, sub string) (bool, error) {
	err := resolver.Check(ctx, Tuple{Object: c.Tupleset.Object, Relation: c.Tupleset.Relation, Subject: sub})
	if err != nil && err != ErrNoConnection {
		return false, err
	}

	return err != ErrNoConnection, nil
}

func (c ComputedUserset) String() string {
	return fmt.Sprintf("(computed %s)", c.Tupleset)
}

func GetObjectTypeAndID(obj string) (string, string) {
	parts := strings.SplitN(obj, ":", 2)
	if len(parts) != 2 {
		return "", parts[0]
	}
	return parts[0], parts[1]
}

type LazyDirect struct {
	Tupleset Tupleset
}

func (u LazyDirect) Has(ctx context.Context, resolver LazyResolver, sub string) (bool, error) {
	err := resolver.CheckDirect(ctx, Tuple{Object: u.Tupleset.Object, Relation: u.Tupleset.Relation, Subject: sub})
	if err != nil && err != ErrNoConnection {
		return false, fmt.Errorf("failed direct check: %w", err)
	}
	return err != ErrNoConnection, nil
}

func (d LazyDirect) String() string {
	return fmt.Sprintf("(direct %s)", d.Tupleset)
}

type ComputedUsersetViaTupleset struct {
	Tupleset        Tupleset
	UsersetRelation string
}

func (u ComputedUsersetViaTupleset) Has(ctx context.Context, resolver LazyResolver, sub string) (bool, error) {
	subs, err := resolver.ListSubjects(ctx, u.Tupleset)
	if err != nil {
		return false, fmt.Errorf("failed to list subjects: %w", err)
	}

	for _, obj := range subs {
		tuple := Tuple{Object: obj, Relation: u.UsersetRelation, Subject: sub}
		err := resolver.Check(ctx, tuple)
		if err != nil && err != ErrNoConnection {
			return false, err
		}
		if err == nil {
			return true, nil
		}
	}

	return false, nil
}

func (d ComputedUsersetViaTupleset) String() string {
	return fmt.Sprintf("(computed_via %s#%s)", d.Tupleset, d.UsersetRelation)
}
