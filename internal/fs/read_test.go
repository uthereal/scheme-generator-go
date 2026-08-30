package fs_test

import (
	"errors"
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	internalfs "github.com/uthereal/scheme-generator-go/internal/fs"
)

// errFS is a mock fs.FS implementation that triggers errors during Open.
type errFS struct{}

// Open implements the fs.FS interface for errFS.
func (e *errFS) Open(name string) (fs.File, error) {
	return nil, errors.New("open error")
}

// brokenFileFS delegates directory traversal to an underlying FS,
// but fails when opening actual files.
type brokenFileFS struct {
	fs.FS
}

// Open intercepts file opens to return an error for non-directory paths.
func (b *brokenFileFS) Open(name string) (fs.File, error) {
	if name == "." {
		return b.FS.Open(name)
	}
	return nil, errors.New("open failed for file")
}

func TestReadSQLFiles(t *testing.T) {
	tests := []struct {
		name        string
		fSys        fs.FS
		wantContent string
		wantErr     bool
	}{
		{
			name: "multiple sql files sorted alphabetically",
			fSys: fstest.MapFS{
				"b_users.sql": &fstest.MapFile{
					Data: []byte("CREATE TABLE users ();"),
				},
				"a_auth.sql": &fstest.MapFile{
					Data: []byte("CREATE SCHEMA auth;"),
				},
				"c_sub/nested.sql": &fstest.MapFile{
					Data: []byte("CREATE TABLE nested ();"),
				},
			},
			wantContent: "CREATE SCHEMA auth;\n" +
				"CREATE TABLE users ();\n" +
				"CREATE TABLE nested ();\n",
			wantErr: false,
		},
		{
			name: "ignores non-sql files and matches uppercase sql",
			fSys: fstest.MapFS{
				"schema.SQL": &fstest.MapFile{
					Data: []byte("CREATE SCHEMA main;"),
				},
				"readme.md": &fstest.MapFile{
					Data: []byte("# Schema Documentation"),
				},
				"config.json": &fstest.MapFile{
					Data: []byte(`{"version": 1}`),
				},
			},
			wantContent: "CREATE SCHEMA main;\n",
			wantErr:     false,
		},
		{
			name:        "empty filesystem",
			fSys:        fstest.MapFS{},
			wantContent: "",
			wantErr:     false,
		},
		{
			name: "filesystem with directories only",
			fSys: fstest.MapFS{
				"dir1": &fstest.MapFile{
					Mode: fs.ModeDir,
				},
				"dir2/subdir": &fstest.MapFile{
					Mode: fs.ModeDir,
				},
			},
			wantContent: "",
			wantErr:     false,
		},
		{
			name:        "filesystem walk error",
			fSys:        &errFS{},
			wantContent: "",
			wantErr:     true,
		},
		{
			name: "file content read error",
			fSys: &brokenFileFS{
				FS: fstest.MapFS{
					"test.sql": &fstest.MapFile{
						Data: []byte("SELECT 1;"),
					},
				},
			},
			wantContent: "",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, err := internalfs.ReadSQLFiles(tt.fSys)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantContent, content)
		})
	}
}

