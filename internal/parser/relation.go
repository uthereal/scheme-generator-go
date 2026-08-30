package parser

import (
	"fmt"
	"slices"

	"github.com/ettle/strcase"
	"github.com/uthereal/scheme-generator-go/internal/ast"
	"github.com/uthereal/scheme-generator-go/internal/inflection"
)

// buildRelations analyzes foreign keys and constraints across all schemas and
// tables to populate BelongsTo, BelongsToMany, HasOne, and HasMany relations.
func (p *PostgresAccumulator) buildRelations() {
	var tables []*ast.Table
	mapQualifiedTableNameToTable := make(map[string]*ast.Table)
	mapModelNameToTable := make(map[string]*ast.Table)

	for _, sc := range p.mapSchemaNameToSchema {
		for _, tbl := range sc.Tables {
			tables = append(tables, tbl)
			qualifiedName := fmt.Sprintf("%s.%s", sc.Name, tbl.Name)
			mapQualifiedTableNameToTable[qualifiedName] = tbl
			_, exists := mapQualifiedTableNameToTable[tbl.Name]
			if !exists {
				mapQualifiedTableNameToTable[tbl.Name] = tbl
			}
		}
	}


	for _, tbl := range tables {
		mapModelNameToTable[tbl.ToGoModelName()] = tbl
	}

	mapQualifiedTableNameToBelongsTo := make(
		map[string][]*ast.RelationBelongsTo,
	)
	mapQualifiedTableNameToBelongsToMany := make(
		map[string][]*ast.RelationBelongsToMany,
	)
	mapQualifiedTableNameToHasOne := make(
		map[string][]*ast.RelationHasOne,
	)
	mapQualifiedTableNameToHasMany := make(
		map[string][]*ast.RelationHasMany,
	)
	mapQualifiedTableNameToIsPivot := make(map[string]bool)

	for _, tbl := range tables {
		qualifiedName := fmt.Sprintf("%s.%s", tbl.Schema.Name, tbl.Name)
		if len(tbl.ForeignKeys) == 2 {
			hasCompositePK := len(tbl.PrimaryKey) >= 2 ||
				len(tbl.UniqueConstraints) > 0
			if hasCompositePK {
				mapQualifiedTableNameToIsPivot[qualifiedName] = true
			}
		}
	}


	for _, tbl := range tables {
		tblQualifiedName := fmt.Sprintf(
			"%s.%s",
			tbl.Schema.Name,
			tbl.Name,
		)
		if mapQualifiedTableNameToIsPivot[tblQualifiedName] {
			fk1 := tbl.ForeignKeys[0]
			fk2 := tbl.ForeignKeys[1]

			t1 := ""
			sc1 := defaultSchemaName
			if fk1.RefTable != nil {
				t1 = fk1.RefTable.Name
				if fk1.RefTable.Schema != nil {
					sc1 = fk1.RefTable.Schema.Name
				}
			}
			t2 := ""
			sc2 := defaultSchemaName
			if fk2.RefTable != nil {
				t2 = fk2.RefTable.Name
				if fk2.RefTable.Schema != nil {
					sc2 = fk2.RefTable.Schema.Name
				}
			}

			col1 := ""
			if len(fk1.FkColumns) > 0 && fk1.FkColumns[0] != nil {
				col1 = fk1.FkColumns[0].Name
			}
			col2 := ""
			if len(fk2.FkColumns) > 0 && fk2.FkColumns[0] != nil {
				col2 = fk2.FkColumns[0].Name
			}

			m1 := strcase.ToGoPascal(inflection.Singular(t1))
			m2 := strcase.ToGoPascal(inflection.Singular(t2))

			q1 := fmt.Sprintf("%s.%s", sc1, t1)
			q2 := fmt.Sprintf("%s.%s", sc2, t2)

			refT1, exists := mapQualifiedTableNameToTable[q1]
			if !exists {
				refT1, exists = mapQualifiedTableNameToTable[t1]
			}
			if exists {
				m1 = refT1.ToGoModelName()
			}

			refT2, exists := mapQualifiedTableNameToTable[q2]
			if !exists {
				refT2, exists = mapQualifiedTableNameToTable[t2]
			}
			if exists {
				m2 = refT2.ToGoModelName()
			}

			var childCol1 *ast.Column
			if len(fk2.RefColumns) > 0 {
				childCol1 = fk2.RefColumns[0]
			} else if refT2 != nil && len(refT2.PrimaryKey) > 0 {
				childCol1 = refT2.PrimaryKey[0]
			} else if refT2 != nil && len(refT2.Columns) > 0 {
				childCol1 = refT2.Columns[0]
			}

			var parentCol1 *ast.Column
			if len(fk1.RefColumns) > 0 {
				parentCol1 = fk1.RefColumns[0]
			} else if refT1 != nil && len(refT1.PrimaryKey) > 0 {
				parentCol1 = refT1.PrimaryKey[0]
			} else if refT1 != nil && len(refT1.Columns) > 0 {
				parentCol1 = refT1.Columns[0]
			}

			r1 := &ast.RelationBelongsToMany{
				Name:               strcase.ToGoPascal(t2),
				FieldName:          strcase.ToGoPascal(t2),
				ParentModel:        m1,
				ChildModel:         m2,
				ChildMutator:       m2 + "Mutator",
				ChildTable:         mapModelNameToTable[m2],
				JoinTableSchema:    tbl.Schema.Name,
				JoinTable:          tbl.Name,
				JoinTableFK1:       col1,
				JoinTableFK2:       col2,
				ChildForeignKeyCol: childCol1,
				ParentKeyCol:       parentCol1,
			}
			mapQualifiedTableNameToBelongsToMany[q1] = append(
				mapQualifiedTableNameToBelongsToMany[q1],
				r1,
			)

			r2 := &ast.RelationBelongsToMany{
				Name:               strcase.ToGoPascal(t1),
				FieldName:          strcase.ToGoPascal(t1),
				ParentModel:        m2,
				ChildModel:         m1,
				ChildMutator:       m1 + "Mutator",
				ChildTable:         mapModelNameToTable[m1],
				JoinTableSchema:    tbl.Schema.Name,
				JoinTable:          tbl.Name,
				JoinTableFK1:       col2,
				JoinTableFK2:       col1,
				ChildForeignKeyCol: parentCol1,
				ParentKeyCol:       childCol1,
			}
			mapQualifiedTableNameToBelongsToMany[q2] = append(
				mapQualifiedTableNameToBelongsToMany[q2],
				r2,
			)

			continue
		}

		for _, fk := range tbl.ForeignKeys {
			if fk.RefTable == nil {
				continue
			}

			refScName := defaultSchemaName
			if fk.RefSchema != nil && fk.RefSchema.Name != "" {
				refScName = fk.RefSchema.Name
			}
			if fk.RefSchema == nil &&
				fk.RefTable.Schema != nil &&
				fk.RefTable.Schema.Name != "" {
				refScName = fk.RefTable.Schema.Name
			}

			refQualifiedName := fmt.Sprintf(
				"%s.%s",
				refScName,
				fk.RefTable.Name,
			)
			refTable, exists := mapQualifiedTableNameToTable[refQualifiedName]
			if !exists {
				refTable, exists = mapQualifiedTableNameToTable[fk.RefTable.Name]
				if !exists {
					continue
				}
			}

			bModel := tbl.ToGoModelName()
			aModel := refTable.ToGoModelName()

			isUnique := false
			for _, unique := range tbl.UniqueConstraints {
				if slices.Equal(unique.ColumnNames(), fk.FkColumnNames()) {
					isUnique = true
					break
				}
			}
			if slices.Equal(
				ast.ColumnNames(tbl.PrimaryKey),
				fk.FkColumnNames(),
			) {
				isUnique = true
			}

			var localKeys []*ast.Column
			if len(fk.RefColumns) > 0 {
				for _, col := range fk.RefColumns {
					for _, c := range refTable.Columns {
						if c.Name == col.Name {
							localKeys = append(localKeys, c)
							break
						}
					}
				}
			}
			if len(fk.RefColumns) == 0 {
				localKeys = refTable.PrimaryKey
			}
			if len(localKeys) == 0 && len(refTable.Columns) > 0 {
				localKeys = []*ast.Column{refTable.Columns[0]}
			}

			bRel := &ast.RelationBelongsTo{
				Name:           aModel,
				FieldName:      aModel,
				ParentModel:    bModel,
				ChildModel:     aModel,
				ChildMutator:   aModel + "Mutator",
				ChildTable:     refTable,
				ForeignKeyCols: fk.FkColumns,
				LocalKeyCols:   localKeys,
			}
			mapQualifiedTableNameToBelongsTo[tblQualifiedName] = append(
				mapQualifiedTableNameToBelongsTo[tblQualifiedName],
				bRel,
			)

			if isUnique {
				aRel := &ast.RelationHasOne{
					Name:           bModel,
					FieldName:      bModel,
					ParentModel:    aModel,
					ChildModel:     bModel,
					ChildMutator:   bModel + "Mutator",
					ChildTable:     tbl,
					ForeignKeyCols: fk.FkColumns,
					LocalKeyCols:   localKeys,
				}
				mapQualifiedTableNameToHasOne[refQualifiedName] = append(
					mapQualifiedTableNameToHasOne[refQualifiedName],
					aRel,
				)
				continue
			}

			aRel := &ast.RelationHasMany{
				Name: strcase.ToGoPascal(
					inflection.Plural(tbl.Name),
				),
				FieldName: strcase.ToGoPascal(
					inflection.Plural(tbl.Name),
				),
				ParentModel:    aModel,
				ChildModel:     bModel,
				ChildMutator:   bModel + "Mutator",
				ChildTable:     tbl,
				ForeignKeyCols: fk.FkColumns,
				LocalKeyCols:   localKeys,
			}
			mapQualifiedTableNameToHasMany[refQualifiedName] = append(
				mapQualifiedTableNameToHasMany[refQualifiedName],
				aRel,
			)
		}

	}

	for _, tbl := range tables {
		tblQualifiedName := fmt.Sprintf(
			"%s.%s",
			tbl.Schema.Name,
			tbl.Name,
		)
		tbl.BelongsTo = mapQualifiedTableNameToBelongsTo[tblQualifiedName]
		tbl.BelongsToMany = mapQualifiedTableNameToBelongsToMany[tblQualifiedName]
		tbl.HasOne = mapQualifiedTableNameToHasOne[tblQualifiedName]
		tbl.HasMany = mapQualifiedTableNameToHasMany[tblQualifiedName]
	}
}
