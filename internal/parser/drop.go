package parser

import (
	"errors"
	"fmt"

	"github.com/pganalyze/pg_query_go/v6"
)

// dropSchema removes a schema from PostgresAccumulator.
func (p *PostgresAccumulator) dropSchema(
	stmt *pg_query.Node_DropStmt,
) error {
	for _, objNode := range stmt.DropStmt.Objects {
		if objNode == nil {
			continue
		}

		schemaName := pgQueryObjectToString(objNode)
		if schemaName == "" {
			continue
		}

		_, exists := p.mapSchemaNameToSchema[schemaName]
		if !exists {
			if stmt.DropStmt.MissingOk {
				continue
			}
			return fmt.Errorf("schema '%s' does not exist", schemaName)
		}

		delete(p.mapSchemaNameToSchema, schemaName)
	}

	return nil
}

// dropTable removes a table from its schema.
func (p *PostgresAccumulator) dropTable(
	stmt *pg_query.Node_DropStmt,
) error {
	for _, objNode := range stmt.DropStmt.Objects {
		if objNode == nil {
			continue
		}

		names := pgQueryObjectToStrings(objNode)
		scName := defaultSchemaName
		tblName := ""

		switch true {
		case len(names) == 1:
			tblName = names[0]
		case len(names) > 1:
			scName = names[0]
			tblName = names[1]
		}

		if tblName == "" {
			continue
		}

		schema, exists := p.mapSchemaNameToSchema[scName]
		if !exists {
			if stmt.DropStmt.MissingOk {
				continue
			}
			return fmt.Errorf(
				"table '%s.%s' does not exist",
				scName,
				tblName,
			)
		}

		_, exists = schema.Tables[tblName]
		if !exists {
			if stmt.DropStmt.MissingOk {
				continue
			}
			return fmt.Errorf(
				"table '%s.%s' does not exist",
				scName,
				tblName,
			)
		}

		delete(schema.Tables, tblName)
	}

	return nil
}

// dropEnum removes an enum type from its schema.
func (p *PostgresAccumulator) dropEnum(
	stmt *pg_query.Node_DropStmt,
) error {
	for _, objNode := range stmt.DropStmt.Objects {
		if objNode == nil {
			continue
		}

		names := pgQueryObjectToStrings(objNode)
		scName := defaultSchemaName
		typeName := ""

		switch true {
		case len(names) == 1:
			typeName = names[0]
		case len(names) > 1:
			scName = names[0]
			typeName = names[1]
		}

		if typeName == "" {
			continue
		}

		schema, exists := p.mapSchemaNameToSchema[scName]
		if !exists {
			if stmt.DropStmt.MissingOk {
				continue
			}
			return fmt.Errorf(
				"enum '%s.%s' does not exist",
				scName,
				typeName,
			)
		}

		delete(schema.Enums, typeName)
	}

	return nil
}

// dropCast returns an error as DROP CAST is not supported.
func (p *PostgresAccumulator) dropCast(
	stmt *pg_query.Node_DropStmt,
) error {
	return errors.New("drop cast is not supported")
}

// dropColumn returns an error as DROP COLUMN is only supported via ALTER TABLE.
func (p *PostgresAccumulator) dropColumn(
	stmt *pg_query.Node_DropStmt,
) error {
	return errors.New(
		"drop column is not supported as a top-level statement",
	)
}

// dropConstraint returns an error as DROP CONSTRAINT is only supported via
// ALTER TABLE.
func (p *PostgresAccumulator) dropConstraint(
	stmt *pg_query.Node_DropStmt,
) error {
	return errors.New(
		"drop constraint is not supported as a top-level statement",
	)
}
