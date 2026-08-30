package ast

import (
	"cmp"
	"slices"

	"github.com/ettle/strcase"
)

// Schema represents intermediate schema metadata.
type Schema struct {
	Name   string
	Tables map[string]*Table
	Enums  map[string]*Enum
}

// ToGoPascalName returns the PascalCase representation of the schema name.
func (s *Schema) ToGoPascalName() string {
	return strcase.ToGoPascal(s.Name)
}

// SortedTables returns the schema tables sorted by name.
func (s *Schema) SortedTables() []*Table {
	res := make([]*Table, 0, len(s.Tables))
	for _, t := range s.Tables {
		res = append(res, t)
	}

	slices.SortFunc(res, func(a *Table, b *Table) int {
		return cmp.Compare(a.Name, b.Name)
	})

	return res
}
