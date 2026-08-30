package emitter

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uthereal/scheme-generator-go/internal/ast"
)

func TestNewEmitter(t *testing.T) {
	outDir := "/tmp/output"
	pkgName := "samplepkg"
	em := NewEmitter(outDir, pkgName)

	require.NotNil(t, em)
	assert.Equal(t, outDir, em.OutputDir)
	assert.Equal(t, pkgName, em.PackageName)
}

func TestEmitter_EmitSchemas(t *testing.T) {
	t.Run("successfully emits all files for schemas with enums", func(
		t *testing.T,
	) {
		tempDir := t.TempDir()
		outDir := filepath.Join(tempDir, "generated")
		em := NewEmitter(outDir, "testpkg")

		pubSchema := &ast.Schema{
			Name: "public",
			Enums: map[string]*ast.Enum{
				"status": {
					Name:   "status",
					Values: []string{"active", "inactive"},
				},
			},
			Tables: map[string]*ast.Table{
				"users": {
					Name: "users",
					Columns: []*ast.Column{
						{
							Name: "id",
							Type: "bigint",
						},
						{
							Name: "name",
							Type: "text",
						},
					},
				},
			},
		}
		pubSchema.Tables["users"].Schema = pubSchema

		err := em.EmitSchemas([]*ast.Schema{pubSchema})
		require.NoError(t, err)

		files := []string{
			"models.go",
			"mutator.go",
			"schema.go",
			"table.go",
			"hydrate.go",
			"dehydrate.go",
			"relations.go",
			"query.go",
			"enums.go",
		}
		for _, f := range files {
			p := filepath.Join(outDir, f)
			assert.FileExists(t, p)
		}
	})

	t.Run("does not emit enums.go when no enums present", func(
		t *testing.T,
	) {
		tempDir := t.TempDir()
		outDir := filepath.Join(tempDir, "no_enums")
		em := NewEmitter(outDir, "testpkg")

		pubSchema := &ast.Schema{
			Name:  "public",
			Enums: make(map[string]*ast.Enum),
			Tables: map[string]*ast.Table{
				"items": {
					Name: "items",
					Columns: []*ast.Column{
						{
							Name: "id",
							Type: "integer",
						},
					},
				},
			},
		}
		pubSchema.Tables["items"].Schema = pubSchema

		err := em.EmitSchemas([]*ast.Schema{pubSchema})
		require.NoError(t, err)

		enumsPath := filepath.Join(outDir, "enums.go")
		assert.NoFileExists(t, enumsPath)
		assert.FileExists(t, filepath.Join(outDir, "models.go"))
	})

	t.Run("returns error when target directory cannot be created", func(
		t *testing.T,
	) {
		tempDir := t.TempDir()
		filePath := filepath.Join(tempDir, "regular_file")
		err := os.WriteFile(filePath, []byte("data"), 0644)
		require.NoError(t, err)

		// Using regular file as directory path causes MkdirAll to fail
		invalidDir := filepath.Join(filePath, "subfolder")
		em := NewEmitter(invalidDir, "testpkg")

		err = em.EmitSchemas([]*ast.Schema{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create directory")
	})
}

func TestEmitter_EmitHelpers(t *testing.T) {
	tempDir := t.TempDir()
	em := NewEmitter(tempDir, "helpertest")

	pubSchema := &ast.Schema{
		Name: "public",
		Enums: map[string]*ast.Enum{
			"role": {
				Name:   "role",
				Values: []string{"admin", "member"},
			},
		},
		Tables: map[string]*ast.Table{
			"users": {
				Name: "users",
				Columns: []*ast.Column{
					{
						Name: "id",
						Type: "bigint",
					},
				},
			},
		},
	}
	pubSchema.Tables["users"].Schema = pubSchema
	tables := []*ast.Table{pubSchema.Tables["users"]}
	schemas := []*ast.Schema{pubSchema}

	t.Run("emitModels writes models.go", func(t *testing.T) {
		err := em.emitModels(tempDir, tables)
		require.NoError(t, err)
		assert.FileExists(t, filepath.Join(tempDir, "models.go"))
	})

	t.Run("emitMutator writes mutator.go", func(t *testing.T) {
		err := em.emitMutator(tempDir, tables)
		require.NoError(t, err)
		assert.FileExists(t, filepath.Join(tempDir, "mutator.go"))
	})

	t.Run("emitSchema writes schema.go", func(t *testing.T) {
		err := em.emitSchema(tempDir, schemas)
		require.NoError(t, err)
		assert.FileExists(t, filepath.Join(tempDir, "schema.go"))
	})

	t.Run("emitTable writes table.go", func(t *testing.T) {
		err := em.emitTable(tempDir, schemas)
		require.NoError(t, err)
		assert.FileExists(t, filepath.Join(tempDir, "table.go"))
	})

	t.Run("emitHydrate writes hydrate.go", func(t *testing.T) {
		err := em.emitHydrate(tempDir, tables)
		require.NoError(t, err)
		assert.FileExists(t, filepath.Join(tempDir, "hydrate.go"))
	})

	t.Run("emitDehydrate writes dehydrate.go", func(t *testing.T) {
		err := em.emitDehydrate(tempDir, tables)
		require.NoError(t, err)
		assert.FileExists(t, filepath.Join(tempDir, "dehydrate.go"))
	})

	t.Run("emitRelations writes relations.go", func(t *testing.T) {
		err := em.emitRelations(tempDir, tables)
		require.NoError(t, err)
		assert.FileExists(t, filepath.Join(tempDir, "relations.go"))
	})

	t.Run("emitQuery writes query.go", func(t *testing.T) {
		err := em.emitQuery(tempDir, tables)
		require.NoError(t, err)
		assert.FileExists(t, filepath.Join(tempDir, "query.go"))
	})

	t.Run("emitEnums writes enums.go when enums present", func(
		t *testing.T,
	) {
		err := em.emitEnums(tempDir, schemas)
		require.NoError(t, err)
		assert.FileExists(t, filepath.Join(tempDir, "enums.go"))
	})

	t.Run("emitEnums does not write enums.go when no enums", func(
		t *testing.T,
	) {
		emptyDir := t.TempDir()
		emptySchemas := []*ast.Schema{{Name: "public"}}
		err := em.emitEnums(emptyDir, emptySchemas)
		require.NoError(t, err)
		assert.NoFileExists(t, filepath.Join(emptyDir, "enums.go"))
	})
}
