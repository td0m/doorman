package doorman

import "fmt"

type Tuple struct {
	Subject Object
	Role    string
	Object  Object
}

type TupleWithPath struct {
	Tuple
	Path  Path
}

func (t Tuple) Equal(r Tuple) bool {
	return r.Object == t.Object && r.Role == t.Role && r.Subject == t.Subject
}

func (t Tuple) String() string {
	return fmt.Sprintf("(%s, %s, %s)", t.Subject, t.Role, t.Object)
}

func NewTuple(sub Object, role string, obj Object) Tuple {
	return Tuple{sub, role, obj}
}

type Connection struct {
	Role   string
	Object Object
}

type Path []Connection
