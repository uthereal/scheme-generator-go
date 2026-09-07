package ast

import (
	"strings"

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

// ToGoTableName converts a table state to a Go table name for scoped
// accessors.
func (t *Table) ToGoTableName() string {
	rawName := t.Name
	if t.Schema == nil || t.Schema.Name == "" {
		return formatTableIdentifier(rawName)
	}

	prefix := strings.ToLower(t.Schema.Name) + "_"
	hasPrefix := strings.HasPrefix(strings.ToLower(rawName), prefix)
	if !hasPrefix || len(rawName) <= len(prefix) {
		return formatTableIdentifier(rawName)
	}

	candidate := rawName[len(prefix):]
	if t.Schema.Tables != nil {
		if _, exists := t.Schema.Tables[candidate]; exists {
			return formatTableIdentifier(rawName)
		}
	}

	return formatTableIdentifier(candidate)
}

// formatTableIdentifier inflects and formats a table identifier into
// PascalCase.
func formatTableIdentifier(rawName string) string {
	return pipe.NewPipe(rawName).
		DoUnary(inflection.Singular).
		DoUnary(strcase.ToGoPascal).
		Unwrap()
}

// ToGoModelName converts a table state to a Go model name.
func (t *Table) ToGoModelName() string {
	tblName := t.ToGoTableName()
	if t.Schema == nil || t.Schema.Name == "" {
		return tblName
	}

	schemaName := pipe.NewPipe(t.Schema.Name).
		DoUnary(inflection.Singular).
		DoUnary(strcase.ToGoPascal).
		Unwrap()

	return schemaName + tblName
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
