package fs

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// copyEntry holds source and destination paths for directory traversal.
type copyEntry struct {
	src string
	dst string
}

// CopyDir recursively copies files and directories from src into dst
// using iterative DFS to avoid recursion limits.
func CopyDir(src string, dst string) error {
	err := os.MkdirAll(dst, 0755)
	if err != nil {
		return fmt.Errorf("failed to create destination directory -> %w", err)
	}

	stack := make([]copyEntry, 0)
	stack = append(stack, copyEntry{
		src: src,
		dst: dst,
	})

	for len(stack) > 0 {
		current := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		entries, errRead := os.ReadDir(current.src)
		if errRead != nil {
			return fmt.Errorf("failed to read directory -> %w", errRead)
		}

		for _, entry := range entries {
			srcPath := filepath.Join(current.src, entry.Name())
			dstPath := filepath.Join(current.dst, entry.Name())

			if entry.IsDir() {
				err = os.MkdirAll(dstPath, 0755)
				if err != nil {
					return fmt.Errorf(
						"failed to create destination subdirectory -> %w",
						err,
					)
				}
				stack = append(stack, copyEntry{
					src: srcPath,
					dst: dstPath,
				})
				continue
			}

			err = CopyFile(srcPath, dstPath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// CopyFile copies a single file from src to dst.
func CopyFile(src string, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed opening source file -> %w", err)
	}
	defer func() {
		_ = in.Close()
	}()

	out, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed creating target file -> %w", err)
	}
	defer func() {
		_ = out.Close()
	}()

	_, err = io.Copy(out, in)
	if err != nil {
		return fmt.Errorf("failed writing file content -> %w", err)
	}

	return nil
}
