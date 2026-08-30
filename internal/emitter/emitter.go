package emitter

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/uthereal/scheme-generator-go/internal/ast"
)

// Emitter generates Go files from the intermediate database state.
type Emitter struct {
	OutputDir   string
	PackageName string
}

// outDirectoryUnixPermission defines default Unix file permission for output
// directory creation.
var outDirectoryUnixPermission fs.FileMode = 0755

// NewEmitter initializes a new Emitter.
func NewEmitter(outDir string, pkgName string) *Emitter {
	return &Emitter{
		OutputDir:   outDir,
		PackageName: pkgName,
	}
}

// EmitSchemas generates all Go files for the provided schemas.
func (e *Emitter) EmitSchemas(schemas []*ast.Schema) error {
	targetDir := e.OutputDir
	err := os.MkdirAll(targetDir, outDirectoryUnixPermission)
	if err != nil {
		return fmt.Errorf("failed to create directory -> %w", err)
	}

	var tables []*ast.Table
	for _, sc := range schemas {
		tables = append(tables, sc.SortedTables()...)
	}

	err = e.emitModels(targetDir, tables)
	if err != nil {
		return err
	}

	err = e.emitMutator(targetDir, tables)
	if err != nil {
		return err
	}

	err = e.emitSchema(targetDir, schemas)
	if err != nil {
		return err
	}

	err = e.emitTable(targetDir, schemas)
	if err != nil {
		return err
	}

	err = e.emitHydrate(targetDir, tables)
	if err != nil {
		return err
	}

	err = e.emitDehydrate(targetDir, tables)
	if err != nil {
		return err
	}

	err = e.emitRelations(targetDir, tables)
	if err != nil {
		return err
	}

	err = e.emitQuery(targetDir, tables)
	if err != nil {
		return err
	}

	err = e.emitEnums(targetDir, schemas)
	if err != nil {
		return err
	}

	return nil
}

// emitModels renders model struct definitions to models.go.
func (e *Emitter) emitModels(
	targetDir string,
	tables []*ast.Table,
) error {
	data := struct {
		PackageName string
		Tables      []*ast.Table
	}{
		PackageName: e.PackageName,
		Tables:      tables,
	}
	return e.renderTemplate(
		modelTemplateFile,
		data,
		filepath.Join(targetDir, "models.go"),
	)
}

// emitMutator renders mutator struct definitions to mutator.go.
func (e *Emitter) emitMutator(
	targetDir string,
	tables []*ast.Table,
) error {
	data := struct {
		PackageName string
		Tables      []*ast.Table
	}{
		PackageName: e.PackageName,
		Tables:      tables,
	}
	return e.renderTemplate(
		mutatorTemplateFile,
		data,
		filepath.Join(targetDir, "mutator.go"),
	)
}

// emitSchema renders schema definitions to schema.go.
func (e *Emitter) emitSchema(
	targetDir string,
	schemas []*ast.Schema,
) error {
	data := struct {
		PackageName string
		Schemas     []*ast.Schema
	}{
		PackageName: e.PackageName,
		Schemas:     schemas,
	}
	return e.renderTemplate(
		schemaTemplateFile,
		data,
		filepath.Join(targetDir, "schema.go"),
	)
}

// emitTable renders table column definitions to table.go.
func (e *Emitter) emitTable(
	targetDir string,
	schemas []*ast.Schema,
) error {
	data := struct {
		PackageName string
		Schemas     []*ast.Schema
	}{
		PackageName: e.PackageName,
		Schemas:     schemas,
	}
	return e.renderTemplate(
		tableTemplateFile,
		data,
		filepath.Join(targetDir, "table.go"),
	)
}

// emitHydrate renders hydrator methods to hydrate.go.
func (e *Emitter) emitHydrate(
	targetDir string,
	tables []*ast.Table,
) error {
	data := struct {
		PackageName string
		Tables      []*ast.Table
	}{
		PackageName: e.PackageName,
		Tables:      tables,
	}
	return e.renderTemplate(
		hydrateTemplateFile,
		data,
		filepath.Join(targetDir, "hydrate.go"),
	)
}

// emitDehydrate renders dehydrator methods to dehydrate.go.
func (e *Emitter) emitDehydrate(
	targetDir string,
	tables []*ast.Table,
) error {
	data := struct {
		PackageName string
		Tables      []*ast.Table
	}{
		PackageName: e.PackageName,
		Tables:      tables,
	}
	return e.renderTemplate(
		dehydrateTemplateFile,
		data,
		filepath.Join(targetDir, "dehydrate.go"),
	)
}

// emitRelations renders relation definitions to relations.go.
func (e *Emitter) emitRelations(
	targetDir string,
	tables []*ast.Table,
) error {
	data := struct {
		PackageName string
		Tables      []*ast.Table
	}{
		PackageName: e.PackageName,
		Tables:      tables,
	}
	return e.renderTemplate(
		relationTemplateFile,
		data,
		filepath.Join(targetDir, "relations.go"),
	)
}

// emitQuery renders query builder methods to query.go.
func (e *Emitter) emitQuery(
	targetDir string,
	tables []*ast.Table,
) error {
	data := struct {
		PackageName string
		Tables      []*ast.Table
	}{
		PackageName: e.PackageName,
		Tables:      tables,
	}
	return e.renderTemplate(
		queryTemplateFile,
		data,
		filepath.Join(targetDir, "query.go"),
	)
}

// emitEnums renders enum constants to enums.go if any exist in the schemas.
func (e *Emitter) emitEnums(
	targetDir string,
	schemas []*ast.Schema,
) error {
	hasEnums := false
	for _, sc := range schemas {
		if len(sc.Enums) > 0 {
			hasEnums = true
			break
		}
	}

	if !hasEnums {
		return nil
	}

	data := struct {
		PackageName string
		Schemas     []*ast.Schema
	}{
		PackageName: e.PackageName,
		Schemas:     schemas,
	}
	return e.renderTemplate(
		enumTemplateFile,
		data,
		filepath.Join(targetDir, "enums.go"),
	)
}

