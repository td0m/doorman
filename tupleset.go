package doorman

import (
	"errors"
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
	tupleset, err := NewTupleset(s)
	if err != nil {
		panic(err)
	}

	return tupleset
}

func GetObjectTypeAndID(obj string) (string, string) {
	parts := strings.SplitN(obj, ":", 2)
	if len(parts) != 2 {
		return "", parts[0]
	}
	return parts[0], parts[1]
}
