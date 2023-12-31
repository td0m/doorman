package doorman

import (
	"errors"
	"strings"
)

type Tuple struct {
	Object   string
	Relation string
	Subject  string
}

func (t Tuple) String() string {
	return t.Object + "#" + t.Relation + "@" + t.Subject
}

func NewTuple(s string) (Tuple, error) {
	parts := strings.SplitN(s, "@", 2)
	if len(parts) != 2 {
		return Tuple{}, errors.New("expected to contain exactly one '@'")
	}

	setStr, sub := parts[0], parts[1]
	tupleset, err := NewTupleset(setStr)
	if err != nil {
		return Tuple{}, err
	}

	stype, _ := GetObjectTypeAndID(sub)
	if stype == "" {
		return Tuple{}, errors.New("subject must have a type")
	}

	return Tuple{Subject: sub, Object: tupleset.Object, Relation: tupleset.Relation}, nil
}

func MustNewTuple(s string) Tuple {
	t, err := NewTuple(s)
	if err != nil {
		panic(err)
	}

	return t
}
