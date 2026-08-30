package emitter

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uthereal/scheme-generator-go/internal/ast"
)

func TestEmitter_RenderTemplate(t *testing.T) {
	tempDir := t.TempDir()
	e := NewEmitter(tempDir, "testpkg")

	t.Run("render enum template successfully", func(t *testing.T) {
		targetPath := filepath.Join(tempDir, "enum.go")
		data := struct {
			PackageName string
			Schemas     []*ast.Schema
		}{
			PackageName: "testpkg",
			Schemas: []*ast.Schema{
				{
					Name: "public",
					Enums: map[string]*ast.Enum{
						"status": {
							Name:   "status",
							Values: []string{"active", "inactive"},
						},
					},
				},
			},
		}

		err := e.renderTemplate(enumTemplateFile, data, targetPath)
		require.NoError(t, err)

		content, err := os.ReadFile(targetPath)
		require.NoError(t, err)

		src := string(content)
		assert.Contains(t, src, "package testpkg")
		assert.Contains(t, src, "type Status string")
		assert.Contains(t, src, "StatusActive")
		assert.Contains(t, src, `"active"`)
		assert.Contains(t, src, "StatusInactive")
		assert.Contains(t, src, `"inactive"`)
	})

	t.Run("render model template successfully", func(t *testing.T) {
		targetPath := filepath.Join(tempDir, "model.go")
		pubSchema := &ast.Schema{Name: "public"}
		data := struct {
			PackageName string
			Tables      []*ast.Table
		}{
			PackageName: "testpkg",
			Tables: []*ast.Table{
				{
					Schema: pubSchema,
					Name:   "users",
					Columns: []*ast.Column{
						{
							Name: "id",
							Type: "integer",
						},
						{
							Name: "name",
							Type: "text",
						},
					},
				},
			},
		}

		err := e.renderTemplate(modelTemplateFile, data, targetPath)
		require.NoError(t, err)

		content, err := os.ReadFile(targetPath)
		require.NoError(t, err)

		src := string(content)
		assert.Contains(t, src, "package testpkg")
		assert.Contains(t, src, "type PublicUser struct {")
		assert.Contains(t, src, "ID")
		assert.Contains(t, src, "int32")
		assert.Contains(t, src, "Name")
		assert.Contains(t, src, "string")
	})

	t.Run("render template with invalid template file fails", func(t *testing.T) {
		targetPath := filepath.Join(tempDir, "invalid.go")
		err := e.renderTemplate("non_existent.tmpl", nil, targetPath)
		assert.Error(t, err)
	})
}


