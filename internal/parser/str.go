package parser

import (
	"github.com/pganalyze/pg_query_go/v6"
)

// pgQueryNodeToString extracts the string value from a PostgreSQL node.
func pgQueryNodeToString(node *pg_query.Node) string {
	str, ok := node.Node.(*pg_query.Node_String_)
	if !ok || str.String_ == nil {
		return ""
	}

	return str.String_.Sval
}

// pgQueryObjectToString extracts a single string identifier or list item.
func pgQueryObjectToString(obj *pg_query.Node) string {
	names := pgQueryObjectToStrings(obj)
	if len(names) == 0 {
		return ""
	}
	return names[0]
}

// pgQueryObjectToStrings extracts string identifiers from an object node.
func pgQueryObjectToStrings(objNode *pg_query.Node) []string {
	var names []string

	list, ok := objNode.Node.(*pg_query.Node_List)
	if ok && list.List != nil {
		for _, item := range list.List.Items {
			if item == nil {
				continue
			}
			names = append(names, pgQueryNodeToString(item))
		}
		return names
	}

	str, ok := objNode.Node.(*pg_query.Node_String_)
	if ok && str.String_ != nil {
		names = append(names, str.String_.Sval)
		return names
	}

	return names
}

// extractTypeName returns the data type name from a TypeName node.
func extractTypeName(
	typeName *pg_query.TypeName,
) string {
	names := typeName.Names
	if len(names) == 0 {
		return ""
	}

	lastNode := names[len(names)-1]
	return pgQueryNodeToString(lastNode)
}

// extractSchemaAndName extracts schema and object names using switch true.
func extractSchemaAndName(
	names []*pg_query.Node,
) (string, string) {
	scName := defaultSchemaName
	objName := ""

	switch true {
	case len(names) == 1:
		objName = pgQueryNodeToString(names[0])
	case len(names) > 1:
		scName = pgQueryNodeToString(names[0])
		objName = pgQueryNodeToString(names[1])
	}

	return scName, objName
}
