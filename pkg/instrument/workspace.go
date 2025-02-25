package instrument

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// PrepareWorkspace copies the target Go project from projectDir into a new temporary workspace.
// It returns the absolute path to the workspace and an error if the operation fails.
//
// Parameters:
//   - projectDir (string): the path to the target Go project directory.
//
// Returns:
//   - string: the path to the temporary workspace directory.
//   - error: an error object if any error occurs during workspace preparation.
func PrepareWorkspace(projectDir string) (string, error) {
	tempDir, err := os.MkdirTemp("", "tracewrap_workspace_*")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary directory: %v", err)
	}
	err = copyDir(projectDir, tempDir)
	if err != nil {
		return "", fmt.Errorf("failed to copy project: %v", err)
	}
	return tempDir, nil
}

// copyDir recursively copies the directory tree from src to dst.
// It skips the subdirectory named "tracewrap" in the source.
//
// Parameters:
//   - src (string): the source directory path.
//   - dst (string): the destination directory path.
//
// Returns:
//   - error: an error object if any error occurs during the copy process.
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		if info.IsDir() && relPath == "tracewrap" {
			return filepath.SkipDir
		}
		targetPath := filepath.Join(dst, relPath)
		if info.IsDir() {
			return os.MkdirAll(targetPath, info.Mode())
		}
		return copyFile(path, targetPath, info)
	})
}

// copyFile copies a single file from srcFile to dstFile using the provided file information.
// It returns an error if the file cannot be copied.
//
// Parameters:
//   - srcFile (string): the path to the source file.
//   - dstFile (string): the path to the destination file.
//   - info (os.FileInfo): the file information used to preserve file mode.
//
// Returns:
//   - error: an error object if any error occurs during the file copy.
func copyFile(srcFile, dstFile string, info os.FileInfo) error {
	srcF, err := os.Open(srcFile)
	if err != nil {
		return err
	}
	defer srcF.Close()
	dstF, err := os.OpenFile(dstFile, os.O_CREATE|os.O_WRONLY, info.Mode())
	if err != nil {
		return err
	}
	defer dstF.Close()
	_, err = io.Copy(dstF, srcF)
	return err
}
