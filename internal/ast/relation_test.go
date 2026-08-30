package ast_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/uthereal/scheme-generator-go/internal/ast"
)

func TestRelation_Subtypes(t *testing.T) {
	childSchema := &ast.Schema{Name: "auth"}
	childTable := &ast.Table{
		Name:   "users",
		Schema: childSchema,
		Columns: []*ast.Column{
			{Name: "id", Type: "uuid"},
			{Name: "username", Type: "text"},
		},
	}

	fkCol := &ast.Column{Name: "user_id", Type: "uuid"}
	localCol := &ast.Column{Name: "id", Type: "uuid"}

	t.Run("belongs to", func(t *testing.T) {
		rel := ast.RelationBelongsTo{
			Name:           "User",
			FieldName:      "User",
			ParentModel:    "Post",
			ChildModel:     "AuthUser",
			ChildMutator:   "AuthUserMutator",
			ChildTable:     childTable,
			ForeignKeyCols: []*ast.Column{fkCol},
			LocalKeyCols:   []*ast.Column{localCol},
		}

		expected := "relation.BelongsTo[\n\t\t\t\tPost,\n\t\t\t\tAuthUser," +
			"\n\t\t\t\tAuthUserMutator,\n\t\t\t]"
		assert.Equal(
			t,
			expected,
			rel.ToQueryRelationType(),
		)
		assert.Equal(t, []*ast.Column{fkCol}, rel.ForeignKeyCols)
		assert.Equal(t, []*ast.Column{localCol}, rel.LocalKeyCols)
	})

	t.Run("has one", func(t *testing.T) {
		rel := ast.RelationHasOne{
			Name:           "Profile",
			FieldName:      "Profile",
			ParentModel:    "User",
			ChildModel:     "AuthProfile",
			ChildMutator:   "AuthProfileMutator",
			ChildTable:     childTable,
			ForeignKeyCols: []*ast.Column{fkCol},
			LocalKeyCols:   []*ast.Column{localCol},
		}

		expected := "relation.HasOne[\n\t\t\t\tUser,\n\t\t\t\tAuthProfile," +
			"\n\t\t\t\tAuthProfileMutator,\n\t\t\t]"
		assert.Equal(
			t,
			expected,
			rel.ToQueryRelationType(),
		)
		assert.Equal(t, []*ast.Column{fkCol}, rel.ForeignKeyCols)
		assert.Equal(t, []*ast.Column{localCol}, rel.LocalKeyCols)
	})

	t.Run("has many", func(t *testing.T) {
		rel := ast.RelationHasMany{
			Name:           "Users",
			FieldName:      "Users",
			ParentModel:    "Tenant",
			ChildModel:     "AuthUser",
			ChildMutator:   "AuthUserMutator",
			ChildTable:     childTable,
			ForeignKeyCols: []*ast.Column{fkCol},
			LocalKeyCols:   []*ast.Column{localCol},
		}

		expected := "relation.HasMany[\n\t\t\t\tTenant,\n\t\t\t\tAuthUser," +
			"\n\t\t\t\tAuthUserMutator,\n\t\t\t]"
		assert.Equal(
			t,
			expected,
			rel.ToQueryRelationType(),
		)
		assert.Equal(t, []*ast.Column{fkCol}, rel.ForeignKeyCols)
		assert.Equal(t, []*ast.Column{localCol}, rel.LocalKeyCols)
	})

	t.Run("belongs to many", func(t *testing.T) {
		rel := ast.RelationBelongsToMany{
			Name:         "Users",
			FieldName:    "Users",
			ParentModel:  "Group",
			ChildModel:   "AuthUser",
			ChildMutator: "AuthUserMutator",
			ChildTable:   childTable,
			JoinTable:    "user_groups",
			JoinTableFK1: "group_id",
			JoinTableFK2: "user_id",
		}

		expected := "relation.BelongsToMany[\n\t\t\t\tGroup,\n\t\t\t\tAuthUser," +
			"\n\t\t\t\tAuthUserMutator,\n\t\t\t]"
		assert.Equal(
			t,
			expected,
			rel.ToQueryRelationType(),
		)
		assert.Equal(t, "user_groups", rel.JoinTable)
		assert.Equal(t, "group_id", rel.JoinTableFK1)
		assert.Equal(t, "user_id", rel.JoinTableFK2)
	})
}
