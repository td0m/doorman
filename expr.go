package doorman

type Expr interface{}

type NilSet struct{}

type Union []Expr

type Sibling string

type ViaSibling struct {
	SiblingRelation string
	Relation        string
}
