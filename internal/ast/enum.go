package ast

import (
	"github.com/ettle/strcase"
)

// Enum represents intermediate PostgreSQL enum metadata.
type Enum struct {
	Name   string
	Schema *Schema
	Values []string
}

// ToGoPascalName returns the PascalCase representation of the enum name.
func (e *Enum) ToGoPascalName() string {
	return strcase.ToGoPascal(e.Name)
}

// ValueToGoPascalName returns the PascalCase representation of an enum value.
func (e *Enum) ValueToGoPascalName(val string) string {
	return strcase.ToGoPascal(val)
}
