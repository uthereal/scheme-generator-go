package ast_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/uthereal/scheme-generator-go/internal/ast"
)

func TestForeignKey_FkColumnNames(t *testing.T) {
	tests := []struct {
		name string
		fk   ast.ForeignKey
		want []string
	}{
		{
			name: "single column",
			fk: ast.ForeignKey{
				FkColumns: []*ast.Column{{Name: "user_id"}},
			},
			want: []string{"user_id"},
		},
		{
			name: "composite columns",
			fk: ast.ForeignKey{
				FkColumns: []*ast.Column{
					{Name: "tenant_id"},
					{Name: "user_id"},
				},
			},
			want: []string{"tenant_id", "user_id"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.fk.FkColumnNames())
		})
	}
}

func TestForeignKey_RefColumnNames(t *testing.T) {
	tests := []struct {
		name string
		fk   ast.ForeignKey
		want []string
	}{
		{
			name: "single column",
			fk: ast.ForeignKey{
				RefColumns: []*ast.Column{{Name: "id"}},
			},
			want: []string{"id"},
		},
		{
			name: "composite columns",
			fk: ast.ForeignKey{
				RefColumns: []*ast.Column{
					{Name: "tenant_id"},
					{Name: "id"},
				},
			},
			want: []string{"tenant_id", "id"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.fk.RefColumnNames())
		})
	}
}
