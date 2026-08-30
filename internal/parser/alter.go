package parser

import (
	"errors"
	"fmt"

	"github.com/pganalyze/pg_query_go/v6"
	"github.com/uthereal/scheme-generator-go/internal/ast"
)

// alterTableAddColumn appends a new column definition to the specified table.
func (p *PostgresAccumulator) alterTableAddColumn(
	atCmd *pg_query.AlterTableCmd,
	schemaName string,
	tableName string,
) error {
	if atCmd.Def == nil {
		return nil
	}

	tbl, err := p.getTable(schemaName, tableName)
	if err != nil {
		return err
	}

	colDefNode, ok := atCmd.Def.Node.(*pg_query.Node_ColumnDef)
	if !ok || colDefNode.ColumnDef == nil {
		return errors.New("expected column definition in add column")
	}


	colDef := colDefNode.ColumnDef
	col := &ast.Column{
		Name:       colDef.Colname,
		IsNullable: true,
	}

	if colDef.TypeName != nil {
		col.Type = extractTypeName(colDef.TypeName)
		col.IsArray = len(colDef.TypeName.ArrayBounds) > 0
	}

	for _, constrNode := range colDef.Constraints {
		if constrNode == nil {
			continue
		}

		constr, okConstr := constrNode.Node.(*pg_query.Node_Constraint)
		if !okConstr || constr.Constraint == nil {
			continue
		}

		err = p.applyColumnConstraint(constr.Constraint, col, tbl)
		if err != nil {
			return err
		}
	}

	tbl.Columns = append(tbl.Columns, col)
	return nil
}

// alterTableDropColumn removes a column from the specified table.
func (p *PostgresAccumulator) alterTableDropColumn(
	atCmd *pg_query.AlterTableCmd,
	schemaName string,
	tableName string,
) error {
	tbl, err := p.getTable(schemaName, tableName)
	if err != nil {
		return err
	}

	for i, col := range tbl.Columns {
		if col.Name != atCmd.Name {
			continue
		}
		tbl.Columns = append(tbl.Columns[:i], tbl.Columns[i+1:]...)
		break
	}

	return nil
}

// alterTableColumnDefault updates or clears a default value for a column.
func (p *PostgresAccumulator) alterTableColumnDefault(
	atCmd *pg_query.AlterTableCmd,
	schemaName string,
	tableName string,
) error {
	tbl, err := p.getTable(schemaName, tableName)
	if err != nil {
		return err
	}

	for idx := range tbl.Columns {
		if tbl.Columns[idx].Name != atCmd.Name {
			continue
		}

		switch atCmd.Def != nil {
		case true:
			val := "default"
			tbl.Columns[idx].Default = &val
		case false:
			tbl.Columns[idx].Default = nil
		}
		break
	}

	return nil
}

// alterTableColumnNullability alters the nullable constraint on a column.
func (p *PostgresAccumulator) alterTableColumnNullability(
	colName string,
	isNullable bool,
	schemaName string,
	tableName string,
) error {
	tbl, err := p.getTable(schemaName, tableName)
	if err != nil {
		return err
	}

	for idx := range tbl.Columns {
		if tbl.Columns[idx].Name != colName {
			continue
		}
		tbl.Columns[idx].IsNullable = isNullable
		break
	}

	return nil
}

// alterTableColumnType alters the data type of an existing column.
func (p *PostgresAccumulator) alterTableColumnType(
	atCmd *pg_query.AlterTableCmd,
	schemaName string,
	tableName string,
) error {
	if atCmd.Def == nil {
		return nil
	}

	tbl, err := p.getTable(schemaName, tableName)
	if err != nil {
		return err
	}

	colDefNode, ok := atCmd.Def.Node.(*pg_query.Node_ColumnDef)
	if !ok || colDefNode.ColumnDef == nil {
		return nil
	}

	colName := colDefNode.ColumnDef.Colname
	if colName == "" {
		colName = atCmd.Name
	}

	newType := ""
	if colDefNode.ColumnDef.TypeName != nil {
		newType = extractTypeName(colDefNode.ColumnDef.TypeName)
	}

	for idx := range tbl.Columns {
		if tbl.Columns[idx].Name != colName {
			continue
		}
		tbl.Columns[idx].Type = newType
		break
	}

	return nil
}

// alterTableAddConstraint adds a primary key, unique, or foreign key
// constraint to an existing table.
func (p *PostgresAccumulator) alterTableAddConstraint(
	atCmd *pg_query.AlterTableCmd,
	schemaName string,
	tableName string,
) error {
	if atCmd.Def == nil {
		return nil
	}

	tbl, err := p.getTable(schemaName, tableName)
	if err != nil {
		return err
	}

	constrNode, ok := atCmd.Def.Node.(*pg_query.Node_Constraint)
	if !ok || constrNode.Constraint == nil {
		return errors.New("expected constraint in add constraint")
	}


	return p.applyTableConstraint(constrNode.Constraint, tbl)
}

// alterTableDropConstraint drops a named constraint from an existing table.
func (p *PostgresAccumulator) alterTableDropConstraint(
	atCmd *pg_query.AlterTableCmd,
	schemaName string,
	tableName string,
) error {
	tbl, err := p.getTable(schemaName, tableName)
	if err != nil {
		return err
	}

	var newPKs []*ast.Column
	constrName := atCmd.Name
	for _, pk := range tbl.PrimaryKey {
		if pk.Name == constrName {
			continue
		}
		newPKs = append(newPKs, pk)
	}
	tbl.PrimaryKey = newPKs

	return nil
}

// alterEnum appends a new enum value to an existing enum type.
func (p *PostgresAccumulator) alterEnum(
	n *pg_query.Node_AlterEnumStmt,
) error {
	stmt := n.AlterEnumStmt
	if len(stmt.TypeName) == 0 {
		return errors.New("missing type name in alter type")
	}


	scName, typeName := extractSchemaAndName(stmt.TypeName)
	schema, exists := p.mapSchemaNameToSchema[scName]
	if !exists {
		return fmt.Errorf("schema '%s' does not exist", scName)
	}
	enum, exists := schema.Enums[typeName]
	if !exists {
		return fmt.Errorf("enum '%s.%s' does not exist", scName, typeName)
	}

	if stmt.NewVal != "" {
		enum.Values = append(enum.Values, stmt.NewVal)
	}

	return nil
}

// alterObjectSchema moves an existing table to a new schema.
func (p *PostgresAccumulator) alterObjectSchema(
	n *pg_query.Node_AlterObjectSchemaStmt,
) error {
	stmt := n.AlterObjectSchemaStmt
	dstName := stmt.Newschema
	if dstName == "" {
		dstName = defaultSchemaName
	}
	dstSchema, err := p.getSchema(dstName)
	if err != nil {
		return err
	}

	if stmt.Relation != nil {
		srcName := stmt.Relation.Schemaname
		if srcName == "" {
			srcName = defaultSchemaName
		}
		objName := stmt.Relation.Relname
		srcSchema, errSrc := p.getSchema(srcName)
		if errSrc != nil {
			return errSrc
		}

		tbl, exists := srcSchema.Tables[objName]
		switch exists {
		case true:
			_, targetExists := dstSchema.Tables[objName]
			if targetExists {
				return fmt.Errorf(
					"table '%s.%s' already exists",
					dstName,
					objName,
				)
			}
			tbl.Schema = dstSchema
			dstSchema.Tables[objName] = tbl
			delete(srcSchema.Tables, objName)
		case false:
			if !stmt.MissingOk {
				return fmt.Errorf(
					"table '%s.%s' does not exist",
					srcName,
					objName,
				)
			}
		}
	}

	return nil
}

// resolveForeignKeyTarget retrieves or creates target schema/table for FKs.
func (p *PostgresAccumulator) resolveForeignKeyTarget(
	pktable *pg_query.RangeVar,
) (*ast.Schema, *ast.Table) {
	refScName := pktable.Schemaname
	if refScName == "" {
		refScName = defaultSchemaName
	}


	refSc, exists := p.mapSchemaNameToSchema[refScName]
	if !exists {
		refSc = &ast.Schema{
			Name:   refScName,
			Enums:  make(map[string]*ast.Enum),
			Tables: make(map[string]*ast.Table),
		}
		p.mapSchemaNameToSchema[refScName] = refSc
	}

	refTblName := pktable.Relname
	refTbl, existsTbl := refSc.Tables[refTblName]
	if !existsTbl {
		refTbl = &ast.Table{
			Name:   refTblName,
			Schema: refSc,
		}
		refSc.Tables[refTblName] = refTbl
	}

	return refSc, refTbl
}

// appendForeignKeyReferences populates referenced column pointers on an FK.
func appendForeignKeyReferences(
	fk *ast.ForeignKey,
	pkAttrs []*pg_query.Node,
) {
	for _, pkNode := range pkAttrs {
		refColName := pgQueryNodeToString(pkNode)
		var refColPtr *ast.Column
		if fk.RefTable != nil {
			for _, rc := range fk.RefTable.Columns {
				if rc.Name == refColName {
					refColPtr = rc
					break
				}
			}
			if refColPtr == nil {
				refColPtr = &ast.Column{Name: refColName}
				fk.RefTable.Columns = append(
					fk.RefTable.Columns,
					refColPtr,
				)
			}
		} else {
			refColPtr = &ast.Column{Name: refColName}
		}
		fk.RefColumns = append(fk.RefColumns, refColPtr)
	}
}

// applyColumnConstraint applies an inline column constraint to column/table.
func (p *PostgresAccumulator) applyColumnConstraint(
	c *pg_query.Constraint,
	col *ast.Column,
	table *ast.Table,
) error {
	switch c.Contype {
	case pg_query.ConstrType_CONSTR_PRIMARY:
		table.PrimaryKey = append(table.PrimaryKey, col)

	case pg_query.ConstrType_CONSTR_NOTNULL:
		col.IsNullable = false

	case pg_query.ConstrType_CONSTR_NULL:
		col.IsNullable = true

	case pg_query.ConstrType_CONSTR_DEFAULT:
		val := "default"
		col.Default = &val

	case pg_query.ConstrType_CONSTR_UNIQUE:
		table.UniqueConstraints = append(
			table.UniqueConstraints,
			&ast.UniqueConstraint{
				Columns: []*ast.Column{col},
			},
		)

	case pg_query.ConstrType_CONSTR_FOREIGN:
		fk := ast.ForeignKey{
			FkColumns: []*ast.Column{col},
		}
		if c.Pktable != nil {
			fk.RefSchema, fk.RefTable = p.resolveForeignKeyTarget(c.Pktable)
		}
		appendForeignKeyReferences(&fk, c.PkAttrs)
		table.ForeignKeys = append(table.ForeignKeys, &fk)
	}

	return nil
}

// applyTableConstraint applies a table-level constraint to table metadata.
func (p *PostgresAccumulator) applyTableConstraint(
	c *pg_query.Constraint,
	table *ast.Table,
) error {
	switch c.Contype {
	case pg_query.ConstrType_CONSTR_PRIMARY:
		for _, keyNode := range c.Keys {
			colName := pgQueryNodeToString(keyNode)
			var colPtr *ast.Column
			for _, col := range table.Columns {
				if col.Name == colName {
					colPtr = col
					break
				}
			}
			if colPtr != nil {
				table.PrimaryKey = append(table.PrimaryKey, colPtr)
			}
		}

	case pg_query.ConstrType_CONSTR_UNIQUE:
		var cols []*ast.Column
		for _, keyNode := range c.Keys {
			colName := pgQueryNodeToString(keyNode)
			var colPtr *ast.Column
			for _, col := range table.Columns {
				if col.Name == colName {
					colPtr = col
					break
				}
			}
			if colPtr != nil {
				cols = append(cols, colPtr)
			}
		}
		table.UniqueConstraints = append(
			table.UniqueConstraints,
			&ast.UniqueConstraint{
				Columns: cols,
			},
		)

	case pg_query.ConstrType_CONSTR_FOREIGN:
		fk := ast.ForeignKey{}
		if c.Pktable != nil {
			fk.RefSchema, fk.RefTable = p.resolveForeignKeyTarget(c.Pktable)
		}
		for _, fkNode := range c.FkAttrs {

			colName := pgQueryNodeToString(fkNode)
			var colPtr *ast.Column
			for _, col := range table.Columns {
				if col.Name == colName {
					colPtr = col
					break
				}
			}
			if colPtr == nil {
				colPtr = &ast.Column{Name: colName}
			}
			fk.FkColumns = append(fk.FkColumns, colPtr)
		}
		appendForeignKeyReferences(&fk, c.PkAttrs)
		table.ForeignKeys = append(table.ForeignKeys, &fk)
	}

	return nil
}
