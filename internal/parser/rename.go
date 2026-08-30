package parser

import (
	"fmt"

	"github.com/pganalyze/pg_query_go/v6"
)

// renameSchema renames an existing schema in PostgresAccumulator.
func (p *PostgresAccumulator) renameSchema(
	stmt *pg_query.Node_RenameStmt,
) error {
	oldName := stmt.RenameStmt.Subname
	if oldName == "" && stmt.RenameStmt.Relation != nil {
		oldName = stmt.RenameStmt.Relation.Relname
	}
	if oldName == "" && stmt.RenameStmt.Object != nil {
		oldName = pgQueryObjectToString(stmt.RenameStmt.Object)
	}

	newName := stmt.RenameStmt.Newname
	sc, exists := p.mapSchemaNameToSchema[oldName]
	if !exists {
		if stmt.RenameStmt.MissingOk {
			return nil
		}
		return fmt.Errorf("schema '%s' does not exist", oldName)
	}

	_, targetExists := p.mapSchemaNameToSchema[newName]
	if targetExists {
		return fmt.Errorf("schema '%s' already exists", newName)
	}

	sc.Name = newName
	p.mapSchemaNameToSchema[newName] = sc
	delete(p.mapSchemaNameToSchema, oldName)
	return nil
}

// renameTable renames a table in its schema.
func (p *PostgresAccumulator) renameTable(
	stmt *pg_query.Node_RenameStmt,
) error {
	scName := defaultSchemaName
	if stmt.RenameStmt.Relation != nil &&
		stmt.RenameStmt.Relation.Schemaname != "" {
		scName = stmt.RenameStmt.Relation.Schemaname
	}

	schema, exists := p.mapSchemaNameToSchema[scName]
	if !exists {
		if stmt.RenameStmt.MissingOk {
			return nil
		}
		return fmt.Errorf("schema '%s' does not exist", scName)
	}
	oldName := stmt.RenameStmt.Relation.Relname
	newName := stmt.RenameStmt.Newname

	tbl, exists := schema.Tables[oldName]
	if !exists {
		if stmt.RenameStmt.MissingOk {
			return nil
		}
		return fmt.Errorf(
			"table '%s.%s' does not exist",
			scName,
			oldName,
		)
	}

	_, targetExists := schema.Tables[newName]
	if targetExists {
		return fmt.Errorf(
			"table '%s.%s' already exists",
			scName,
			newName,
		)
	}

	tbl.Name = newName
	schema.Tables[newName] = tbl
	delete(schema.Tables, oldName)
	return nil
}

// renameColumn renames a column on the specified table.
func (p *PostgresAccumulator) renameColumn(
	stmt *pg_query.Node_RenameStmt,
) error {
	scName := defaultSchemaName
	if stmt.RenameStmt.Relation != nil &&
		stmt.RenameStmt.Relation.Schemaname != "" {
		scName = stmt.RenameStmt.Relation.Schemaname
	}

	schema, exists := p.mapSchemaNameToSchema[scName]
	if !exists {
		if stmt.RenameStmt.MissingOk {
			return nil
		}
		return fmt.Errorf("schema '%s' does not exist", scName)
	}
	tblName := stmt.RenameStmt.Relation.Relname
	tbl, exists := schema.Tables[tblName]
	if !exists {
		if stmt.RenameStmt.MissingOk {
			return nil
		}
		return fmt.Errorf(
			"table '%s.%s' does not exist",
			scName,
			tblName,
		)
	}

	oldCol := stmt.RenameStmt.Subname
	newCol := stmt.RenameStmt.Newname

	for _, col := range tbl.Columns {
		if col.Name == newCol {
			return fmt.Errorf(
				"column '%s' already exists in table '%s.%s'",
				newCol,
				scName,
				tblName,
			)
		}
	}

	found := false
	for _, col := range tbl.Columns {
		if col.Name == oldCol {
			col.Name = newCol
			found = true
			break
		}
	}

	if !found && !stmt.RenameStmt.MissingOk {
		return fmt.Errorf(
			"column '%s' does not exist in table '%s.%s'",
			oldCol,
			scName,
			tblName,
		)
	}

	return nil
}

// renameType renames an enum type in the specified schema.
func (p *PostgresAccumulator) renameType(
	stmt *pg_query.Node_RenameStmt,
) error {
	scName := defaultSchemaName
	if stmt.RenameStmt.Relation != nil &&
		stmt.RenameStmt.Relation.Schemaname != "" {
		scName = stmt.RenameStmt.Relation.Schemaname
	}

	oldName := stmt.RenameStmt.Subname
	if oldName == "" && stmt.RenameStmt.Object != nil {
		oldName = pgQueryObjectToString(stmt.RenameStmt.Object)
	}
	newName := stmt.RenameStmt.Newname

	schema, exists := p.mapSchemaNameToSchema[scName]
	if !exists {
		if stmt.RenameStmt.MissingOk {
			return nil
		}
		return fmt.Errorf("schema '%s' does not exist", scName)
	}
	enum, exists := schema.Enums[oldName]
	if !exists {
		if stmt.RenameStmt.MissingOk {
			return nil
		}
		return fmt.Errorf("enum '%s.%s' does not exist", scName, oldName)
	}

	_, targetExists := schema.Enums[newName]
	if targetExists {
		return fmt.Errorf("enum '%s.%s' already exists", scName, newName)
	}

	enum.Name = newName
	schema.Enums[newName] = enum
	delete(schema.Enums, oldName)

	return nil
}
