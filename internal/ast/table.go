package ast

import (
	"github.com/ettle/strcase"
	"github.com/uthereal/scheme-generator-go/internal/inflection"
	"github.com/uthereal/scheme-generator-go/internal/pipe"
)

// Table represents intermediate table metadata.
type Table struct {
	Name              string
	Schema            *Schema
	Columns           []*Column
	PrimaryKey        []*Column
	ForeignKeys       []*ForeignKey
	UniqueConstraints []*UniqueConstraint
	BelongsTo         []*RelationBelongsTo
	BelongsToMany     []*RelationBelongsToMany
	HasOne            []*RelationHasOne
	HasMany           []*RelationHasMany
}

// ToGoModelName converts a table state to a Go model name.
func (t *Table) ToGoModelName() string {
	schemaName := pipe.NewPipe(t.Schema.Name).
		DoUnary(inflection.Singular).
		DoUnary(strcase.ToGoPascal).
		Unwrap()
	tableName := pipe.NewPipe(t.Name).
		DoUnary(inflection.Singular).
		DoUnary(strcase.ToGoPascal).
		Unwrap()

	return schemaName + tableName
}

// ColumnNames retrieves the names of all columns from the table.
func (t *Table) ColumnNames() []string {
	return ColumnNames(t.Columns)
}

// PrimaryKeyColumnNames returns a slice of column names that form the primary
// key of the table.
func (t *Table) PrimaryKeyColumnNames() []string {
	return ColumnNames(t.PrimaryKey)
}
