package emitter

import (
	"bytes"
	"embed"
	"fmt"
	"go/format"
	"io/fs"
	"os"
	"text/template"

	"golang.org/x/tools/imports"
)

// templateFolder specifies template directory in the embedded filesystem.
const templateFolder = "template"

// dehydrateTemplateFile is template filename for generating dehydration logic.
const dehydrateTemplateFile = "dehydrate.go.tmpl"

// enumTemplateFile is template filename for generating enum definitions.
const enumTemplateFile = "enum.go.tmpl"

// hydrateTemplateFile is template filename for generating hydration logic.
const hydrateTemplateFile = "hydrate.go.tmpl"

// modelTemplateFile is template filename for generating model structs.
const modelTemplateFile = "model.go.tmpl"

// mutatorTemplateFile is template filename for generating mutator methods.
const mutatorTemplateFile = "mutator.go.tmpl"

// queryTemplateFile is template filename for generating query builders.
const queryTemplateFile = "query.go.tmpl"

// relationTemplateFile is template filename for generating relations.
const relationTemplateFile = "relations.go.tmpl"

// schemaTemplateFile is template filename for generating schema definitions.
const schemaTemplateFile = "schema.go.tmpl"

// tableTemplateFile is template filename for generating table mappings.
const tableTemplateFile = "table.go.tmpl"

// outFileUnixPermission defines default file permissions for generated files.
const outFileUnixPermission fs.FileMode = 0644

// templateFS embeds template directory containing code generation templates.
//
//go:embed template/*.tmpl
var templateFS embed.FS

// renderTemplate parses, executes, formats, and writes a template.
func (e *Emitter) renderTemplate(
	tmplFile string,
	data any,
	targetPath string,
) error {
	tmplPath := templateFolder + "/" + tmplFile
	tmpl, err := template.New(tmplFile).ParseFS(templateFS, tmplPath)
	if err != nil {
		return fmt.Errorf(
			"failed to parse template '%s' -> %w",
			tmplFile,
			err,
		)
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return fmt.Errorf(
			"failed to execute template '%s' -> %w",
			tmplFile,
			err,
		)
	}

	formatted, err := imports.Process(targetPath, buf.Bytes(), nil)
	if err != nil {
		formatted, err = format.Source(buf.Bytes())
		if err != nil {
			formatted = buf.Bytes()
		}
	}

	return os.WriteFile(targetPath, formatted, outFileUnixPermission)
}

