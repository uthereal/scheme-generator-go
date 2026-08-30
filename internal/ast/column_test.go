package ast_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/uthereal/scheme-generator-go/internal/ast"
)

func TestColumn_ToModelType(t *testing.T) {
	tests := []struct {
		name string
		col  ast.Column
		want string
	}{
		{
			name: "basic integer",
			col: ast.Column{
				Name: "id",
				Type: "integer",
			},
			want: "int32",
		},
		{
			name: "nullable uuid",
			col: ast.Column{
				Name:       "user_id",
				Type:       "uuid",
				IsNullable: true,
			},
			want: "*uuid.UUID",
		},
		{
			name: "array nullable text",
			col: ast.Column{
				Name:       "tags",
				Type:       "text",
				IsArray:    true,
				IsNullable: true,
			},
			want: "*[]string",
		},
		{
			name: "custom enum",
			col: ast.Column{
				Name:       "status",
				Type:       "status_enum",
				CustomType: ast.CustomEnum,
			},
			want: "StatusEnum",
		},
		{
			name: "custom domain",
			col: ast.Column{
				Name:       "pos_int",
				Type:       "positive_int",
				CustomType: ast.CustomDomain,
			},
			want: "PositiveInt",
		},
		{
			name: "nullable custom domain",
			col: ast.Column{
				Name:       "pos_int",
				Type:       "positive_int",
				CustomType: ast.CustomDomain,
				IsNullable: true,
			},
			want: "*PositiveInt",
		},
		{
			name: "fallback unknown type",
			col: ast.Column{
				Name:       "extra",
				Type:       "custom_type_unknown",
				CustomType: ast.CustomNone,
			},
			want: "string",
		},
		{
			name: "fallback unknown type array nullable",
			col: ast.Column{
				Name:       "extras",
				Type:       "custom_type_unknown",
				CustomType: ast.CustomNone,
				IsArray:    true,
				IsNullable: true,
			},
			want: "*[]string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.col.ToModelType())
		})
	}
}

func TestColumn_ToQueryColumnType(t *testing.T) {
	tests := []struct {
		name      string
		modelName string
		col       ast.Column
		want      string
	}{
		{
			name:      "uuid column",
			modelName: "User",
			col: ast.Column{
				Name: "id",
				Type: "uuid",
			},
			want: "column.UUIDColumn[User]",
		},
		{
			name:      "nullable numeric",
			modelName: "User",
			col: ast.Column{
				Name:       "age",
				Type:       "integer",
				IsNullable: true,
			},
			want: "column.NullableNumericColumn[User, int32]",
		},
		{
			name:      "non-nullable numeric",
			modelName: "User",
			col: ast.Column{
				Name: "count",
				Type: "bigint",
			},
			want: "column.NumericColumn[User, int64]",
		},
		{
			name:      "string column",
			modelName: "User",
			col: ast.Column{
				Name: "email",
				Type: "varchar",
			},
			want: "column.StringColumn[User, string]",
		},
		{
			name:      "nullable string column",
			modelName: "User",
			col: ast.Column{
				Name:       "bio",
				Type:       "text",
				IsNullable: true,
			},
			want: "column.NullableStringColumn[User, string]",
		},
		{
			name:      "timestamp column",
			modelName: "User",
			col: ast.Column{
				Name: "created_at",
				Type: "timestamptz",
			},
			want: "column.TimestampColumn[User, time.Time]",
		},
		{
			name:      "array column",
			modelName: "User",
			col: ast.Column{
				Name:    "tags",
				Type:    "text",
				IsArray: true,
			},
			want: "column.ArrayColumn[User, string]",
		},
		{
			name:      "nullable array column",
			modelName: "User",
			col: ast.Column{
				Name:       "tags",
				Type:       "text",
				IsArray:    true,
				IsNullable: true,
			},
			want: "column.NullableArrayColumn[User, string]",
		},
		{
			name:      "json column",
			modelName: "User",
			col: ast.Column{
				Name: "metadata",
				Type: "jsonb",
			},
			want: "column.JSONColumn[User, []byte]",
		},
		{
			name:      "nullable json column",
			modelName: "User",
			col: ast.Column{
				Name:       "preferences",
				Type:       "jsonb",
				IsNullable: true,
			},
			want: "column.NullableJSONColumn[User, []byte]",
		},
		{
			name:      "custom enum column",
			modelName: "User",
			col: ast.Column{
				Name:       "status",
				Type:       "user_status",
				CustomType: ast.CustomEnum,
			},
			want: "column.StringColumn[User, UserStatus]",
		},
		{
			name:      "nullable custom enum column",
			modelName: "User",
			col: ast.Column{
				Name:       "status",
				Type:       "user_status",
				CustomType: ast.CustomEnum,
				IsNullable: true,
			},
			want: "column.NullableStringColumn[User, UserStatus]",
		},
		{
			name:      "custom domain column",
			modelName: "User",
			col: ast.Column{
				Name:       "pos_int",
				Type:       "positive_int",
				CustomType: ast.CustomDomain,
			},
			want: "column.Column[User, PositiveInt]",
		},
		{
			name:      "nullable custom domain column",
			modelName: "User",
			col: ast.Column{
				Name:       "pos_int",
				Type:       "positive_int",
				CustomType: ast.CustomDomain,
				IsNullable: true,
			},
			want: "column.NullableColumn[User, PositiveInt]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(
				t,
				tt.want,
				tt.col.ToQueryColumnType(tt.modelName),
			)
		})
	}
}

func TestColumn_ToGoPascalName(t *testing.T) {
	tests := []struct {
		name string
		col  ast.Column
		want string
	}{
		{
			name: "snake case name",
			col: ast.Column{
				Name: "user_id",
			},
			want: "UserID",
		},
		{
			name: "name with multiple and empty segments",
			col: ast.Column{
				Name: "__user__id__",
			},
			want: "UserID",
		},
		{
			name: "single word",
			col: ast.Column{
				Name: "name",
			},
			want: "Name",
		},
		{
			name: "empty string",
			col: ast.Column{
				Name: "",
			},
			want: "",
		},
		{
			name: "unicode accented string",
			col: ast.Column{
				Name: "über_größe",
			},
			want: "ÜberGröße",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.col.ToGoPascalName())
		})
	}
}

func TestColumnNames(t *testing.T) {
	tests := []struct {
		name string
		cols []*ast.Column
		want []string
	}{
		{
			name: "multiple columns",
			cols: []*ast.Column{
				{Name: "id"},
				{Name: "name"},
			},
			want: []string{"id", "name"},
		},
		{
			name: "empty slice",
			cols: []*ast.Column{},
			want: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, ast.ColumnNames(tt.cols))
		})
	}
}
