package doorman

import (
	"context"
	"fmt"
)

type SchemaDef struct {
	Types []SchemaTypeDef
}

func (schema SchemaDef) GetDef(ctx context.Context, tupleset Tupleset) (SetDef, error) {
	otype, _ := GetObjectTypeAndID(tupleset.Object)
	for _, typ := range schema.Types {
		if typ.Name == otype {
			for _, rel := range typ.Relations {
				if rel.Name == tupleset.Relation {
					return rel.Value, nil
				}
			}
			return nil, fmt.Errorf("type found but not the relation '%s' in %s", tupleset.Relation, tupleset)
		}
	}
	return nil, fmt.Errorf("failed to find type '%s' in: %s", otype, tupleset)
}

type SchemaRelationDef struct {
	Name  string
	Value SetDef
}

type SchemaTypeDef struct {
	Name      string
	Relations []SchemaRelationDef
}

type SetDef interface {
	ToUserset(ctx context.Context, atObject string) (Userset, error)
}

type UnionDef struct {
	Args []SetDef
}

func (d UnionDef) ToUserset(ctx context.Context, atObject string) (Userset, error) {
	// usersets := make([]Tupleset, len(d.Args))
	// for i, def := range d.Args {
	// 	userset, err := def.ToUserset(ctx, atObject)
	// 	if err != nil {
	// 		return nil, fmt.Errorf("failed: %w", err)
	// 	}
	// 	usersets[i] = userset
	// }

	return UsersetUnion{Args: d.Args}, nil
}

type ComputedUsersetDef struct {
	Relation string
}

func NewComputed(rel string) ComputedUsersetDef {
	return ComputedUsersetDef{rel}
}

func NewComputedVia(tuplerel, usersetrel string) ComputedUsersetViaTuplesetDef {
	return ComputedUsersetViaTuplesetDef{TuplesetRelation: tuplerel, UsersetRelation: usersetrel}
}

func NewUnion(args ...SetDef) UnionDef {
	return UnionDef{Args: args}
}

func (c ComputedUsersetDef) ToUserset(ctx context.Context, atObject string) (Userset, error) {
	return ComputedUserset{
		Tupleset: Tupleset{Object: atObject, Relation: c.Relation},
	}, nil
}

type ComputedUsersetViaTuplesetDef struct {
	TuplesetRelation string
	UsersetRelation  string
}

type StaticComputedUsersetDef struct {
	Userset Tupleset
}

func (d StaticComputedUsersetDef) ToUserset(ctx context.Context, atObject string) (Userset, error) {
	return StaticTupleset{Tupleset: d.Userset}, nil
}

func NewStatic(tupleset Tupleset) StaticComputedUsersetDef {
	return StaticComputedUsersetDef{Userset: tupleset}
}

func (c ComputedUsersetViaTuplesetDef) ToUserset(ctx context.Context, atObject string) (Userset, error) {
	return ComputedUsersetViaTupleset{
		Tupleset:        Tupleset{Object: atObject, Relation: c.TuplesetRelation},
		UsersetRelation: c.UsersetRelation,
	}, nil
}
