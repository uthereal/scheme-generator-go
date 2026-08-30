package fs

import (
	"fmt"
	"io/fs"
	"sort"
	"strings"
)

// sqlSuffix defines standard suffix for SQL files matched in filesystem.
const sqlSuffix = ".sql"

// ReadSQLFiles walks the given filesystem, reads and concatenates
// all .sql files inside sorted alphabetically.
func ReadSQLFiles(fSys fs.FS) (string, error) {
	var sqlFiles []string
	err := fs.WalkDir(
		fSys,
		".",
		func(p string, d fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return fmt.Errorf(
					"failed walking sql directory -> %w",
					walkErr,
				)
			}
			if d.IsDir() {
				return nil
			}

			nameLower := strings.ToLower(p)
			if !strings.HasSuffix(nameLower, sqlSuffix) {
				return nil
			}

			sqlFiles = append(sqlFiles, p)
			return nil
		},
	)
	if err != nil {
		return "", fmt.Errorf("failed walking sql directory -> %w", err)
	}

	sort.Strings(sqlFiles)

	var sb strings.Builder
	for _, f := range sqlFiles {
		var bytes []byte
		bytes, err = fs.ReadFile(fSys, f)
		if err != nil {
			return "", fmt.Errorf("failed reading sql file -> %w", err)
		}
		sb.Write(bytes)
		sb.WriteString("\n")
	}

	return sb.String(), nil
}
