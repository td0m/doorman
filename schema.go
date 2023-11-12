package doorman

import (
	"context"
	"fmt"
)

type Schema struct {
	Types []SchemaType
}

func (schema Schema) GetExpr(ctx context.Context, tupleset Tupleset) (Expr, error) {
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

type SchemaType struct {
	Name      string
	Relations []SchemaRelation
}

type SchemaRelation struct {
	Name  string
	Value Expr
}
