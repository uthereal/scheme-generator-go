package parser

import (
	"errors"
	"fmt"

	"github.com/pganalyze/pg_query_go/v6"
	"github.com/uthereal/scheme-generator-go/internal/ast"
)

// createSchema initializes and adds a new schema to PostgresAccumulator.
func (p *PostgresAccumulator) createSchema(
	n *pg_query.Node_CreateSchemaStmt,
) error {
	schemaName := n.CreateSchemaStmt.Schemaname
	_, exists := p.mapSchemaNameToSchema[schemaName]
	if exists {
		if n.CreateSchemaStmt.IfNotExists {
			return nil
		}
		return fmt.Errorf("schema '%s' already exists", schemaName)
	}

	p.mapSchemaNameToSchema[schemaName] = &ast.Schema{
		Name:   schemaName,
		Enums:  make(map[string]*ast.Enum),
		Tables: make(map[string]*ast.Table),
	}

	return nil
}

// createTable adds a new table metadata entry in the specified schema.
func (p *PostgresAccumulator) createTable(
	n *pg_query.Node_CreateStmt,
) error {
	stmt := n.CreateStmt
	if stmt.Relation == nil {
		return errors.New("missing relation in create table")
	}

	schemaName := stmt.Relation.Schemaname
	if schemaName == "" {
		schemaName = defaultSchemaName
	}

	tableName := stmt.Relation.Relname
	schema, err := p.getSchema(schemaName)
	if err != nil {
		return err
	}

	_, exists := schema.Tables[tableName]
	if exists {
		if stmt.IfNotExists {
			return nil
		}
		return fmt.Errorf(
			"table '%s.%s' already exists",
			schemaName,
			tableName,
		)
	}

	schema.Tables[tableName] = &ast.Table{
		Name:   tableName,
		Schema: schema,
	}

	return nil
}

// createEnum parses and registers an ENUM type in the specified schema.
func (p *PostgresAccumulator) createEnum(
	n *pg_query.Node_CreateEnumStmt,
) error {
	stmt := n.CreateEnumStmt
	if len(stmt.TypeName) == 0 {
		return errors.New("missing type name in create type")
	}


	scName, typeName := extractSchemaAndName(stmt.TypeName)
	schema, err := p.getSchema(scName)
	if err != nil {
		return err
	}

	var enumVals []string
	for _, valNode := range stmt.Vals {
		if valNode == nil {
			continue
		}
		switch v := valNode.Node.(type) {
		case *pg_query.Node_String_:
			if v.String_ != nil {
				enumVals = append(enumVals, v.String_.Sval)
			}
		case *pg_query.Node_AConst:
			strVal, ok := v.AConst.Val.(*pg_query.A_Const_Sval)
			if ok {
				enumVals = append(enumVals, strVal.Sval.Sval)
			}
		}

	}

	schema.Enums[typeName] = &ast.Enum{
		Name:   typeName,
		Schema: schema,
		Values: enumVals,
	}

	return nil
}
