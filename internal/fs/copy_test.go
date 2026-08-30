package fs_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	internalfs "github.com/uthereal/scheme-generator-go/internal/fs"
)

func TestCopyFile(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(t *testing.T, tmpDir string) (string, string)
		wantErr     bool
		errContains string
	}{
		{
			name: "successful file copy",
			setup: func(t *testing.T, tmpDir string) (string, string) {
				src := filepath.Join(tmpDir, "source.txt")
				dst := filepath.Join(tmpDir, "dest.txt")
				err := os.WriteFile(src, []byte("hello world"), 0644)
				require.NoError(t, err)
				return src, dst
			},
			wantErr: false,
		},
		{
			name: "source file does not exist",
			setup: func(t *testing.T, tmpDir string) (string, string) {
				src := filepath.Join(tmpDir, "non_existent.txt")
				dst := filepath.Join(tmpDir, "dest.txt")
				return src, dst
			},
			wantErr:     true,
			errContains: "failed opening source file",
		},
		{
			name: "destination path is invalid",
			setup: func(t *testing.T, tmpDir string) (string, string) {
				src := filepath.Join(tmpDir, "source_valid.txt")
				err := os.WriteFile(src, []byte("content"), 0644)
				require.NoError(t, err)
				dst := filepath.Join(tmpDir, "missing_dir", "dest.txt")
				return src, dst
			},
			wantErr:     true,
			errContains: "failed creating target file",
		},
		{
			name: "reading from source directory fails during copy",
			setup: func(t *testing.T, tmpDir string) (string, string) {
				srcDir := filepath.Join(tmpDir, "src_dir")
				err := os.MkdirAll(srcDir, 0755)
				require.NoError(t, err)
				dst := filepath.Join(tmpDir, "dest.txt")
				return srcDir, dst
			},
			wantErr:     true,
			errContains: "failed writing file content",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			src, dst := tt.setup(t, tmpDir)
			err := internalfs.CopyFile(src, dst)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			srcContent, err := os.ReadFile(src)
			require.NoError(t, err)
			dstContent, err := os.ReadFile(dst)
			require.NoError(t, err)
			assert.Equal(t, srcContent, dstContent)
		})
	}
}

func TestCopyDir(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(t *testing.T, tmpDir string) (string, string)
		verify      func(t *testing.T, src string, dst string)
		wantErr     bool
		errContains string
	}{
		{
			name: "successful nested directory copy",
			setup: func(t *testing.T, tmpDir string) (string, string) {
				srcRoot := filepath.Join(tmpDir, "src")
				dstParent := filepath.Join(tmpDir, "dst")

				err := os.MkdirAll(filepath.Join(srcRoot, "dirA"), 0755)
				require.NoError(t, err)
				err = os.MkdirAll(filepath.Join(srcRoot, "dirB"), 0755)
				require.NoError(t, err)

				err = os.WriteFile(
					filepath.Join(srcRoot, "root.txt"),
					[]byte("root content"),
					0644,
				)
				require.NoError(t, err)
				err = os.WriteFile(
					filepath.Join(srcRoot, "dirA", "a.txt"),
					[]byte("a content"),
					0644,
				)
				require.NoError(t, err)
				err = os.WriteFile(
					filepath.Join(srcRoot, "dirB", "b.txt"),
					[]byte("b content"),
					0644,
				)
				require.NoError(t, err)

				return srcRoot, dstParent
			},
			verify: func(t *testing.T, src string, dst string) {
				assert.FileExists(t, filepath.Join(dst, "root.txt"))
				assert.FileExists(t, filepath.Join(dst, "dirA", "a.txt"))
				assert.FileExists(t, filepath.Join(dst, "dirB", "b.txt"))

				b, err := os.ReadFile(filepath.Join(dst, "dirA", "a.txt"))
				require.NoError(t, err)
				assert.Equal(t, "a content", string(b))
			},
			wantErr: false,
		},
		{
			name: "empty directory copy",
			setup: func(t *testing.T, tmpDir string) (string, string) {
				src := filepath.Join(tmpDir, "empty_src")
				dst := filepath.Join(tmpDir, "empty_dst")
				err := os.MkdirAll(src, 0755)
				require.NoError(t, err)
				return src, dst
			},
			verify: func(t *testing.T, src string, dst string) {
				assert.DirExists(t, dst)
			},
			wantErr: false,
		},
		{
			name: "non-existent source directory",
			setup: func(t *testing.T, tmpDir string) (string, string) {
				src := filepath.Join(tmpDir, "non_existent")
				dst := filepath.Join(tmpDir, "dst")
				return src, dst
			},
			wantErr:     true,
			errContains: "failed to read directory",
		},
		{
			name: "destination directory creation fails",
			setup: func(t *testing.T, tmpDir string) (string, string) {
				src := filepath.Join(tmpDir, "src")
				err := os.MkdirAll(src, 0755)
				require.NoError(t, err)

				// Create a file where the destination directory should be
				dst := filepath.Join(tmpDir, "blocked_dst")
				err = os.WriteFile(dst, []byte("blocker"), 0644)
				require.NoError(t, err)

				return src, dst
			},
			wantErr:     true,
			errContains: "failed to create destination directory",
		},
		{
			name: "subdirectory creation fails during copy",
			setup: func(t *testing.T, tmpDir string) (string, string) {
				srcRoot := filepath.Join(tmpDir, "src_with_sub")
				err := os.MkdirAll(filepath.Join(srcRoot, "sub"), 0755)
				require.NoError(t, err)

				dstDir := filepath.Join(tmpDir, "dst_dir")
				err = os.MkdirAll(dstDir, 0755)
				require.NoError(t, err)

				// Create a file where the destination subdirectory should be
				blockingSub := filepath.Join(dstDir, "sub")
				err = os.WriteFile(blockingSub, []byte("block"), 0644)
				require.NoError(t, err)

				return srcRoot, dstDir
			},
			wantErr:     true,
			errContains: "failed to create destination subdirectory",
		},
		{
			name: "file copy fails during directory traversal",
			setup: func(t *testing.T, tmpDir string) (string, string) {
				srcRoot := filepath.Join(tmpDir, "src_with_file")
				err := os.MkdirAll(srcRoot, 0755)
				require.NoError(t, err)
				err = os.WriteFile(
					filepath.Join(srcRoot, "file.txt"),
					[]byte("test"),
					0644,
				)
				require.NoError(t, err)

				dstDir := filepath.Join(tmpDir, "dst_target_dir")
				err = os.MkdirAll(dstDir, 0755)
				require.NoError(t, err)

				// Create a directory where the destination file should be
				blockingFile := filepath.Join(dstDir, "file.txt")
				err = os.MkdirAll(blockingFile, 0755)
				require.NoError(t, err)

				return srcRoot, dstDir
			},
			wantErr:     true,
			errContains: "failed creating target file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			src, dst := tt.setup(t, tmpDir)
			err := internalfs.CopyDir(src, dst)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			if tt.verify != nil {
				tt.verify(t, src, dst)
			}
		})
	}
}
