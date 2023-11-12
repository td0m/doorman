package doorman

import (
	"context"
	"fmt"
)

type SchemaDef struct {
	Types []SchemaTypeDef
}

func (schema SchemaDef) Resolve(ctx context.Context, resolver Resolver, tuple Tuple) (LazyUserset, error) {
	otype, _ := GetObjectTypeAndID(tuple.Object)
	for _, typ := range schema.Types {
		if typ.Name == otype {
			for _, rel := range typ.Relations {
				if rel.Name == tuple.Relation {
					direct := LazyDirect{Tupleset: Tupleset{Object: tuple.Object, Relation: tuple.Relation}}
					if rel.Value == nil {
						return direct, nil
					}
					val, err := rel.Value.ToSet(ctx, resolver, tuple.Object)
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

type Resolver interface {
	ListSubjects(ctx context.Context, tupleset Tupleset) ([]string, error)
}

type SetDef interface {
	ToSet(ctx context.Context, r Resolver, atObject string) (LazyUserset, error)
}

type UnionDef struct {
	Args []SetDef
}

func (d UnionDef) ToSet(ctx context.Context, r Resolver, atObject string) (LazyUserset, error) {
	usersets := make([]LazyUserset, len(d.Args))
	for i, def := range d.Args {
		userset, err := def.ToSet(ctx, r, atObject)
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

func NewUnion(args ...SetDef) UnionDef {
	return UnionDef{Args: args}
}

func (c ComputedUsersetDef) ToSet(ctx context.Context, r Resolver, atObject string) (LazyUserset, error) {
	return ComputedUserset{
		Tupleset: Tupleset{Object: atObject, Relation: c.Relation},
	}, nil
}
