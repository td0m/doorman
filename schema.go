package doorman

import (
	"context"
	"fmt"
)

type SchemaDef struct {
	Types []SchemaTypeDef
}

func (schema SchemaDef) Resolve(ctx context.Context, tuple Tuple) (LazyUserset, error) {
	otype, _ := GetObjectTypeAndID(tuple.Object)
	for _, typ := range schema.Types {
		if typ.Name == otype {
			for _, rel := range typ.Relations {
				if rel.Name == tuple.Relation {
					direct := LazyDirect{Tupleset: Tupleset{Object: tuple.Object, Relation: tuple.Relation}}
					if rel.Value == nil {
						return direct, nil
					}
					val, err := rel.Value.ToSet(ctx, tuple.Object)
					if err != nil {
						return nil, err
					}
					union := LazyUnionUserset{Args: []LazyUserset{direct, val}}
					return union, nil
				}
			}
			return nil, fmt.Errorf("type found but not the relation '%s' in %s", tuple.Relation, tuple)
		}
	}
	return nil, fmt.Errorf("failed to find type '%s' in: %s", otype, tuple)
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
	ToSet(ctx context.Context, atObject string) (LazyUserset, error)
}

type UnionDef struct {
	Args []SetDef
}

func (d UnionDef) ToSet(ctx context.Context, atObject string) (LazyUserset, error) {
	usersets := make([]LazyUserset, len(d.Args))
	for i, def := range d.Args {
		userset, err := def.ToSet(ctx, atObject)
		if err != nil {
			return nil, fmt.Errorf("failed: %w", err)
		}
		usersets[i] = userset
	}

	return LazyUnionUserset{Args: usersets}, nil
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

func (c ComputedUsersetDef) ToSet(ctx context.Context, atObject string) (LazyUserset, error) {
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

func (d StaticComputedUsersetDef) ToSet(ctx context.Context, atObject string) (LazyUserset, error) {
	return LazyDirect{Tupleset: d.Userset}, nil
}

func NewStatic(tupleset Tupleset) StaticComputedUsersetDef {
	return StaticComputedUsersetDef{Userset: tupleset}
}

func (c ComputedUsersetViaTuplesetDef) ToSet(ctx context.Context, atObject string) (LazyUserset, error) {
	return ComputedUsersetViaTupleset{
		Tupleset:        Tupleset{Object: atObject, Relation: c.TuplesetRelation},
		UsersetRelation: c.UsersetRelation,
	}, nil
}
