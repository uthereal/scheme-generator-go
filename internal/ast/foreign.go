package ast

// ForeignKey represents an intermediate foreign key state.
// It now references columns directly.
type ForeignKey struct {
	FkColumns  []*Column
	RefTable   *Table
	RefSchema  *Schema
	RefColumns []*Column
}

// FkColumnNames returns a slice of column names from the foreign key
// columns.
func (fk *ForeignKey) FkColumnNames() []string {
	return ColumnNames(fk.FkColumns)
}

// RefColumnNames returns a slice of column names from the foreign key
// reference columns.
func (fk *ForeignKey) RefColumnNames() []string {
	return ColumnNames(fk.RefColumns)
}
