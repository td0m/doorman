package doorman

import (
	"errors"
	"strings"
)

type Tuple struct {
	Object   string
	Relation string
	UserID   string
}

func (t Tuple) String() string {
	return t.Object + "#" + t.Relation + "@" + t.UserID
}

func NewTuple(s string) (Tuple, error) {
	parts := strings.SplitN(s, "@", 2)
	if len(parts) != 2 {
		return Tuple{}, errors.New("expected to contain exactly one '@'")
	}

	setStr, userID := parts[0], parts[1]
	userset, err := NewUserset(setStr)
	if err != nil {
		return Tuple{}, err
	}
	return Tuple{UserID: userID, Object: userset.Object, Relation: userset.Relation}, nil
}

func MustNewTuple(s string) Tuple {
	t, err := NewTuple(s)
	if err != nil {
		panic(err)
	}

	return t
}
