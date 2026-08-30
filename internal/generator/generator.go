package generator

import (
	"fmt"
	"io/fs"
	"os"

	"github.com/uthereal/scheme-generator-go/internal/emitter"
	internalfs "github.com/uthereal/scheme-generator-go/internal/fs"
	"github.com/uthereal/scheme-generator-go/internal/parser"
)

// Run programmatically parses SQL schema files from a virtual filesystem
// and emits Go code atomically.
func Run(
	fSys fs.FS,
	pkgName string,
	outDir string,
) error {
	ddl, err := internalfs.ReadSQLFiles(fSys)
	if err != nil {
		return fmt.Errorf("failed reading SQL files -> %w", err)
	}

	state := parser.NewPostgresAccumulator()
	err = state.ParseDDL(ddl)
	if err != nil {
		return fmt.Errorf("failed parsing DDL -> %w", err)
	}

	tempDir, err := os.MkdirTemp("", "scheme-gen-*")
	if err != nil {
		return fmt.Errorf("failed to create temp staging directory -> %w", err)
	}
	defer func() {
		_ = os.RemoveAll(tempDir)
	}()

	err = emitter.NewEmitter(tempDir, pkgName).
		EmitSchemas(state.Schemas)
	if err != nil {
		return fmt.Errorf("failed emitting schemas -> %w", err)
	}

	err = os.MkdirAll(outDir, 0755)
	if err != nil {
		return fmt.Errorf(
			"failed creating destination directory -> %w",
			err,
		)
	}

	err = internalfs.CopyDir(tempDir, outDir)
	if err != nil {
		return fmt.Errorf(
			"failed copying staged files to output directory -> %w",
			err,
		)
	}

	return nil
}
