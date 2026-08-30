package parser

import (
	"cmp"
	"errors"
	"fmt"
	"slices"

	"github.com/pganalyze/pg_query_go/v6"
	"github.com/uthereal/scheme-generator-go/internal/ast"
	wasilibs "github.com/wasilibs/go-pgquery"
)

// PostgresAccumulator holds the parsed intermediate state for all schemas.
type PostgresAccumulator struct {
	Schemas               []*ast.Schema
	mapSchemaNameToSchema map[string]*ast.Schema
}

// defaultSchemaName defines default schema name used when none is specified.
const defaultSchemaName = "public"

// NewPostgresAccumulator initializes a new PostgresAccumulator.
func NewPostgresAccumulator() *PostgresAccumulator {
	p := &PostgresAccumulator{
		mapSchemaNameToSchema: make(map[string]*ast.Schema),
	}
	p.mapSchemaNameToSchema[defaultSchemaName] = &ast.Schema{
		Name:   defaultSchemaName,
		Enums:  make(map[string]*ast.Enum),
		Tables: make(map[string]*ast.Table),
	}
	return p
}

// ParseDDL parses PostgreSQL DDL into the PostgresAccumulator and builds
// sorted schemas and cross-table relations.
func (p *PostgresAccumulator) ParseDDL(ddl string) error {
	res, err := wasilibs.Parse(ddl)
	if err != nil {
		return fmt.Errorf("failed to parse DDL -> %w", err)
	}

	flatNodes := flattenStatements(res.Stmts)
	for _, node := range flatNodes {
		if node == nil {
			continue
		}

		err = p.processStatement(node)
		if err != nil {
			return err
		}
	}

	p.buildRelations()

	resSchemas := make([]*ast.Schema, 0, len(p.mapSchemaNameToSchema))
	for _, sc := range p.mapSchemaNameToSchema {
		resSchemas = append(resSchemas, sc)
	}

	slices.SortFunc(resSchemas, func(a *ast.Schema, b *ast.Schema) int {
		return cmp.Compare(a.Name, b.Name)
	})

	p.Schemas = resSchemas
	return nil
}

// flattenStatements unwraps nested schema statements and table elements into a
// flat sequence of top-level statement nodes.
func flattenStatements(stmts []*pg_query.RawStmt) []*pg_query.Node {
	var flat []*pg_query.Node

	for _, raw := range stmts {
		if raw == nil || raw.Stmt == nil {
			continue
		}

		flat = append(flat, flattenNode(raw.Stmt)...)
	}

	return flat
}

// flattenNode unrolls composite AST nodes (such as schemas, tables, and alter
// commands) into individual flattened statement nodes.
func flattenNode(node *pg_query.Node) []*pg_query.Node {
	var flat []*pg_query.Node

	switch n := node.Node.(type) {
	case *pg_query.Node_CreateSchemaStmt:
		schemaName := n.CreateSchemaStmt.Schemaname
		flat = append(flat, node)

		for _, elt := range n.CreateSchemaStmt.SchemaElts {
			if elt == nil {
				continue
			}
			propagateSchemaContext(elt, schemaName)
			flat = append(flat, flattenNode(elt)...)
		}


	case *pg_query.Node_CreateStmt:
		stmt := n.CreateStmt
		tableElts := stmt.TableElts
		stmt.TableElts = nil
		flat = append(flat, node)

		for _, elt := range tableElts {
			if elt == nil {
				continue
			}

			switch elt.Node.(type) {
			case *pg_query.Node_ColumnDef:
				flat = append(flat, &pg_query.Node{
					Node: &pg_query.Node_AlterTableStmt{
						AlterTableStmt: &pg_query.AlterTableStmt{
							Relation: stmt.Relation,
							Cmds: []*pg_query.Node{
								{
									Node: &pg_query.Node_AlterTableCmd{
										AlterTableCmd: &pg_query.AlterTableCmd{
											Subtype: pg_query.AlterTableType_AT_AddColumn,
											Def:     elt,
										},
									},
								},
							},
						},
					},
				})

			case *pg_query.Node_Constraint:
				flat = append(flat, &pg_query.Node{
					Node: &pg_query.Node_AlterTableStmt{
						AlterTableStmt: &pg_query.AlterTableStmt{
							Relation: stmt.Relation,
							Cmds: []*pg_query.Node{
								{
									Node: &pg_query.Node_AlterTableCmd{
										AlterTableCmd: &pg_query.AlterTableCmd{
											Subtype: pg_query.AlterTableType_AT_AddConstraint,
											Def:     elt,
										},
									},
								},
							},
						},
					},
				})
			}
		}

	case *pg_query.Node_AlterTableStmt:
		stmt := n.AlterTableStmt
		for _, cmd := range stmt.Cmds {
			if cmd == nil {
				continue
			}
			flat = append(flat, &pg_query.Node{
				Node: &pg_query.Node_AlterTableStmt{
					AlterTableStmt: &pg_query.AlterTableStmt{
						Relation:  stmt.Relation,
						MissingOk: stmt.MissingOk,
						Cmds:      []*pg_query.Node{cmd},
					},
				},
			})
		}

	case *pg_query.Node_DropStmt:
		stmt := n.DropStmt
		for _, obj := range stmt.Objects {
			if obj == nil {
				continue
			}
			flat = append(flat, &pg_query.Node{
				Node: &pg_query.Node_DropStmt{
					DropStmt: &pg_query.DropStmt{
						RemoveType: stmt.RemoveType,
						MissingOk:  stmt.MissingOk,
						Objects:    []*pg_query.Node{obj},
					},
				},
			})
		}

	default:
		flat = append(flat, node)
	}

	return flat
}

// propagateSchemaContext assigns parent schema name to child statement nodes
// if the child statement does not explicitly define one.
func propagateSchemaContext(node *pg_query.Node, schemaName string) {
	if schemaName == "" {
		return
	}

	switch n := node.Node.(type) {
	case *pg_query.Node_CreateStmt:
		if n.CreateStmt.Relation != nil &&
			n.CreateStmt.Relation.Schemaname == "" {
			n.CreateStmt.Relation.Schemaname = schemaName
		}
	case *pg_query.Node_CreateEnumStmt:
		if len(n.CreateEnumStmt.TypeName) == 1 {
			origName := n.CreateEnumStmt.TypeName[0]
			n.CreateEnumStmt.TypeName = []*pg_query.Node{
				{
					Node: &pg_query.Node_String_{
						String_: &pg_query.String{
							Sval: schemaName,
						},
					},
				},
				origName,
			}
		}
	}
}

// processStatement routes a PostgreSQL AST node to its dedicated receiver.
func (p *PostgresAccumulator) processStatement(
	node *pg_query.Node,
) error {
	var err error

	switch n := node.Node.(type) {
	case *pg_query.Node_CreateSchemaStmt:
		err = p.createSchema(n)

	case *pg_query.Node_CreateStmt:
		err = p.createTable(n)

	case *pg_query.Node_CreateEnumStmt:
		err = p.createEnum(n)

	case *pg_query.Node_RenameStmt:
		switch n.RenameStmt.RenameType {
		case pg_query.ObjectType_OBJECT_SCHEMA:
			err = p.renameSchema(n)
		case pg_query.ObjectType_OBJECT_TABLE:
			err = p.renameTable(n)
		case pg_query.ObjectType_OBJECT_COLUMN:
			err = p.renameColumn(n)
		case pg_query.ObjectType_OBJECT_TYPE:
			err = p.renameType(n)
		}

	case *pg_query.Node_DropStmt:
		switch n.DropStmt.RemoveType {
		case pg_query.ObjectType_OBJECT_SCHEMA:
			err = p.dropSchema(n)
		case pg_query.ObjectType_OBJECT_TABLE:
			err = p.dropTable(n)
		case pg_query.ObjectType_OBJECT_TYPE:
			err = p.dropEnum(n)
		case pg_query.ObjectType_OBJECT_CAST:
			err = p.dropCast(n)
		case pg_query.ObjectType_OBJECT_COLUMN:
			err = p.dropColumn(n)
		case pg_query.ObjectType_OBJECT_TABCONSTRAINT:
			err = p.dropConstraint(n)
		}

	case *pg_query.Node_AlterTableStmt:
		stmt := n.AlterTableStmt
		if stmt.Relation == nil {
			return errors.New("missing relation in alter table")
		}


		schemaName := stmt.Relation.Schemaname
		if schemaName == "" {
			schemaName = defaultSchemaName
		}

		tableName := stmt.Relation.Relname
		_, err = p.getTable(schemaName, tableName)
		if err != nil {
			if stmt.MissingOk {
				return nil
			}
			return err
		}

		for _, cmdNode := range stmt.Cmds {
			if cmdNode == nil {
				continue
			}

			cmd, ok := cmdNode.Node.(*pg_query.Node_AlterTableCmd)
			if !ok || cmd.AlterTableCmd == nil {
				return errors.New("expected alter table command node")
			}

			atCmd := cmd.AlterTableCmd

			switch atCmd.Subtype {
			case pg_query.AlterTableType_AT_AddColumn:
				err = p.alterTableAddColumn(atCmd, schemaName, tableName)

			case pg_query.AlterTableType_AT_DropColumn:
				err = p.alterTableDropColumn(atCmd, schemaName, tableName)

			case pg_query.AlterTableType_AT_ColumnDefault:
				err = p.alterTableColumnDefault(
					atCmd,
					schemaName,
					tableName,
				)

			case pg_query.AlterTableType_AT_DropNotNull:
				err = p.alterTableColumnNullability(
					atCmd.Name,
					true,
					schemaName,
					tableName,
				)

			case pg_query.AlterTableType_AT_SetNotNull:
				err = p.alterTableColumnNullability(
					atCmd.Name,
					false,
					schemaName,
					tableName,
				)

			case pg_query.AlterTableType_AT_AlterColumnType:
				err = p.alterTableColumnType(
					atCmd,
					schemaName,
					tableName,
				)

			case pg_query.AlterTableType_AT_AddConstraint:
				err = p.alterTableAddConstraint(
					atCmd,
					schemaName,
					tableName,
				)

			case pg_query.AlterTableType_AT_DropConstraint:
				err = p.alterTableDropConstraint(
					atCmd,
					schemaName,
					tableName,
				)
			}

			if err != nil {
				return err
			}
		}

	case *pg_query.Node_AlterEnumStmt:
		err = p.alterEnum(n)

	case *pg_query.Node_AlterObjectSchemaStmt:
		err = p.alterObjectSchema(n)

	default:
		// Unknown or unhandled statement node type.
	}

	return err
}

// getSchema retrieves an existing schema or returns an error.
func (p *PostgresAccumulator) getSchema(
	schemaName string,
) (*ast.Schema, error) {
	if schemaName == "" {
		schemaName = defaultSchemaName
	}

	sc, exists := p.mapSchemaNameToSchema[schemaName]
	if !exists {
		return nil, fmt.Errorf("schema '%s' does not exist", schemaName)
	}

	return sc, nil
}

// getTable retrieves an existing table from specified schema or returns error.
func (p *PostgresAccumulator) getTable(
	schemaName string,
	tableName string,
) (*ast.Table, error) {
	sc, err := p.getSchema(schemaName)
	if err != nil {
		return nil, err
	}

	tbl, exists := sc.Tables[tableName]
	if !exists {
		return nil, fmt.Errorf(
			"table '%s.%s' does not exist",
			sc.Name,
			tableName,
		)
	}

	return tbl, nil
}
