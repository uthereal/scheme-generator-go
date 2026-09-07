package ast_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/uthereal/scheme-generator-go/internal/ast"
)

func TestTable_ToGoModelName(t *testing.T) {
	tests := []struct {
		name string
		tbl  ast.Table
		want string
	}{
		{
			name: "plural schema and table",
			tbl: ast.Table{
				Name:   "users",
				Schema: &ast.Schema{Name: "publics"},
			},
			want: "PublicUser",
		},
		{
			name: "snake case schema and table",
			tbl: ast.Table{
				Name:   "blog_posts",
				Schema: &ast.Schema{Name: "content_store"},
			},
			want: "ContentStoreBlogPost",
		},
		{
			name: "nexus schema and users table",
			tbl: ast.Table{
				Name:   "users",
				Schema: &ast.Schema{Name: "nexus"},
			},
			want: "NexusUser",
		},
		{
			name: "public schema and nexus table",
			tbl: ast.Table{
				Name:   "nexus",
				Schema: &ast.Schema{Name: "public"},
			},
			want: "PublicNexus",
		},
		{
			name: "nexus schema and nexus_cache table",
			tbl: ast.Table{
				Name:   "nexus_cache",
				Schema: &ast.Schema{Name: "nexus"},
			},
			want: "NexusCache",
		},
		{
			name: "auth schema and auth_users table",
			tbl: ast.Table{
				Name:   "auth_users",
				Schema: &ast.Schema{Name: "auth"},
			},
			want: "AuthUser",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.tbl.ToGoModelName())
		})
	}
}

func TestTable_ToGoTableName(t *testing.T) {
	tests := []struct {
		name string
		tbl  ast.Table
		want string
	}{
		{
			name: "plural table",
			tbl: ast.Table{
				Name:   "users",
				Schema: &ast.Schema{Name: "public"},
			},
			want: "User",
		},
		{
			name: "snake case table",
			tbl: ast.Table{
				Name:   "blog_posts",
				Schema: &ast.Schema{Name: "content_store"},
			},
			want: "BlogPost",
		},
		{
			name: "prefixed table with schema name",
			tbl: ast.Table{
				Name:   "nexus_cache",
				Schema: &ast.Schema{Name: "nexus"},
			},
			want: "Cache",
		},
		{
			name: "prefixed table with auth schema",
			tbl: ast.Table{
				Name:   "auth_users",
				Schema: &ast.Schema{Name: "auth"},
			},
			want: "User",
		},
		{
			name: "nil schema",
			tbl: ast.Table{
				Name: "devices",
			},
			want: "Device",
		},
		{
			name: "empty schema name",
			tbl: ast.Table{
				Name:   "devices",
				Schema: &ast.Schema{Name: ""},
			},
			want: "Device",
		},
		{
			name: "collision with existing stripped table in schema",
			tbl: ast.Table{
				Name: "nexus_cache",
				Schema: &ast.Schema{
					Name: "nexus",
					Tables: map[string]*ast.Table{
						"cache": {Name: "cache"},
					},
				},
			},
			want: "NexusCache",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.tbl.ToGoTableName())
		})
	}
}

func TestTable_ColumnNames(t *testing.T) {
	tests := []struct {
		name string
		tbl  ast.Table
		want []string
	}{
		{
			name: "multiple columns",
			tbl: ast.Table{
				Name: "users",
				Columns: []*ast.Column{
					{Name: "id"},
					{Name: "username"},
				},
			},
			want: []string{"id", "username"},
		},
		{
			name: "no columns",
			tbl: ast.Table{
				Name:    "empty",
				Columns: []*ast.Column{},
			},
			want: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.tbl.ColumnNames())
		})
	}
}

func TestTable_PrimaryKeyColumnNames(t *testing.T) {
	tests := []struct {
		name string
		tbl  ast.Table
		want []string
	}{
		{
			name: "composite primary key",
			tbl: ast.Table{
				Name: "user_roles",
				PrimaryKey: []*ast.Column{
					{Name: "user_id"},
					{Name: "role_id"},
				},
			},
			want: []string{"user_id", "role_id"},
		},
		{
			name: "no primary key",
			tbl: ast.Table{
				Name:       "logs",
				PrimaryKey: []*ast.Column{},
			},
			want: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.tbl.PrimaryKeyColumnNames())
		})
	}
}
