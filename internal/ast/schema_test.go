package ast_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/uthereal/scheme-generator-go/internal/ast"
)

func TestSchema_ToGoPascalName(t *testing.T) {
	tests := []struct {
		name   string
		schema ast.Schema
		want   string
	}{
		{
			name:   "simple schema name",
			schema: ast.Schema{Name: "public"},
			want:   "Public",
		},
		{
			name:   "snake case schema name",
			schema: ast.Schema{Name: "auth_v2"},
			want:   "AuthV2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.schema.ToGoPascalName())
		})
	}
}

func TestSchema_SortedTables(t *testing.T) {
	tblA := &ast.Table{Name: "accounts"}
	tblB := &ast.Table{Name: "blogs"}
	tblU := &ast.Table{Name: "users"}

	tests := []struct {
		name   string
		schema ast.Schema
		want   []*ast.Table
	}{
		{
			name: "unsorted map returns sorted slice",
			schema: ast.Schema{
				Name: "public",
				Tables: map[string]*ast.Table{
					"users":    tblU,
					"accounts": tblA,
					"blogs":    tblB,
				},
			},
			want: []*ast.Table{tblA, tblB, tblU},
		},
		{
			name: "empty tables map",
			schema: ast.Schema{
				Name:   "empty",
				Tables: map[string]*ast.Table{},
			},
			want: []*ast.Table{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.schema.SortedTables())
		})
	}
}
