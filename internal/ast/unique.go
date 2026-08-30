package ast

// UniqueConstraint represents an intermediate unique constraint state.
// It now references columns directly.
type UniqueConstraint struct {
	Columns []*Column
}

// ColumnNames returns a slice of column names from the unique constraint
// columns.
func (uc *UniqueConstraint) ColumnNames() []string {
	return ColumnNames(uc.Columns)
}
