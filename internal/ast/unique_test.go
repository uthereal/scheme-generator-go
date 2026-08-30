package ast_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/uthereal/scheme-generator-go/internal/ast"
)

func TestUniqueConstraint_ColumnNames(t *testing.T) {
	tests := []struct {
		name string
		uc   ast.UniqueConstraint
		want []string
	}{
		{
			name: "multiple unique columns",
			uc: ast.UniqueConstraint{
				Columns: []*ast.Column{
					{Name: "email"},
					{Name: "username"},
				},
			},
			want: []string{"email", "username"},
		},
		{
			name: "empty unique columns",
			uc: ast.UniqueConstraint{
				Columns: []*ast.Column{},
			},
			want: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.uc.ColumnNames())
		})
	}
}
