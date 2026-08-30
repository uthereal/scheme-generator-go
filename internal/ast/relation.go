package ast

import (
	"fmt"
)

// RelationBelongsTo represents an auto-detected belongs-to relationship.
type RelationBelongsTo struct {
	Name           string
	FieldName      string
	ParentModel    string
	ChildModel     string
	ChildMutator   string
	ChildTable     *Table
	ForeignKeyCols []*Column
	LocalKeyCols   []*Column
}

// RelationHasOne represents an auto-detected has-one relationship.
type RelationHasOne struct {
	Name           string
	FieldName      string
	ParentModel    string
	ChildModel     string
	ChildMutator   string
	ChildTable     *Table
	ForeignKeyCols []*Column
	LocalKeyCols   []*Column
}

// RelationHasMany represents an auto-detected has-many relationship.
type RelationHasMany struct {
	Name           string
	FieldName      string
	ParentModel    string
	ChildModel     string
	ChildMutator   string
	ChildTable     *Table
	ForeignKeyCols []*Column
	LocalKeyCols   []*Column
}

// RelationBelongsToMany represents an auto-detected belongs-to-many relation.
type RelationBelongsToMany struct {
	Name               string
	FieldName          string
	ParentModel        string
	ChildModel         string
	ChildMutator       string
	ChildTable         *Table
	JoinTableSchema    string
	JoinTable          string
	JoinTableFK1       string
	JoinTableFK2       string
	ChildForeignKeyCol *Column
	ParentKeyCol       *Column
}

// ToQueryRelationType returns the type-safe schema relation type.
func (r *RelationBelongsTo) ToQueryRelationType() string {
	return fmt.Sprintf(
		"relation.BelongsTo[\n\t\t\t\t%s,\n\t\t\t\t%s,\n\t\t\t\t%s,\n\t\t\t]",
		r.ParentModel,
		r.ChildModel,
		r.ChildMutator,
	)
}

// ToQueryRelationType returns the type-safe schema relation type.
func (r *RelationHasOne) ToQueryRelationType() string {
	return fmt.Sprintf(
		"relation.HasOne[\n\t\t\t\t%s,\n\t\t\t\t%s,\n\t\t\t\t%s,\n\t\t\t]",
		r.ParentModel,
		r.ChildModel,
		r.ChildMutator,
	)
}

// ToQueryRelationType returns the type-safe schema relation type.
func (r *RelationHasMany) ToQueryRelationType() string {
	return fmt.Sprintf(
		"relation.HasMany[\n\t\t\t\t%s,\n\t\t\t\t%s,\n\t\t\t\t%s,\n\t\t\t]",
		r.ParentModel,
		r.ChildModel,
		r.ChildMutator,
	)
}

// ToQueryRelationType returns the type-safe schema relation type.
func (r *RelationBelongsToMany) ToQueryRelationType() string {
	return fmt.Sprintf(
		"relation.BelongsToMany[\n\t\t\t\t%s,\n\t\t\t\t%s,\n\t\t\t\t%s,\n\t\t\t]",
		r.ParentModel,
		r.ChildModel,
		r.ChildMutator,
	)
}
