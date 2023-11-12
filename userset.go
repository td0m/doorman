package doorman

import (
	"errors"
	"strings"
)

type Userset struct {
	Object   string
	Relation string
}

func (u Userset) String() string {
	return u.Object + "#" + u.Relation
}

func NewUserset(s string) (Userset, error) {
	parts := strings.SplitN(s, "#", 2)
	if len(parts) != 2 {
		return Userset{}, errors.New("expected to contain exactly one '#'")
	}
	return Userset{Object: parts[0], Relation: parts[1]}, nil
}

func MustNewUserset(s string) Userset {
	userset, err := NewUserset(s)
	if err != nil {
		panic(err)
	}

	return userset
}
