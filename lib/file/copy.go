package file

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"go.starlark.net/starlark"
)

// copyFile is a wrapper around copyFileGo for Starlark scripts.
func copyFile(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var (
		src       string
		dst       string
		overwrite = false
	)
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "src", &src, "dst", &dst, "overwrite?", &overwrite); err != nil {
		return starlark.None, err
	}
	dp, err := copyFileGo(src, dst, overwrite)
	if err != nil {
		return nil, err
	}
	return starlark.String(dp), nil
}

// copyFileGo copies the contents of the source file to the destination file or directory with the same mode and access and modification times.
// If the destination file exists and overwrite is false, an error is returned.
// Symbolic links are followed on both source and destination.
// Errors occurred while setting the mode or access and modification times are ignored.
func copyFileGo(src, dst string, overwrite bool) (string, error) {
	// No empty input
	if src == emptyStr {
		return emptyStr, errors.New("source path is empty")
	}
	if dst == emptyStr {
		return emptyStr, errors.New("destination path is empty")
	}

	// Open the source file.
	srcFile, err := os.Open(src)
	if err != nil {
		return emptyStr, fmt.Errorf("open source file: %w", err)
	}
	defer srcFile.Close()

	// Stat the source file to get its mode, times, and owner.
	srcStat, err := srcFile.Stat()
	if err != nil {
		return emptyStr, fmt.Errorf("stat source file: %w", err)
	}
	if !srcStat.Mode().IsRegular() {
		// HACK, not sure if this is the best way to check if the file is a regular file
		return emptyStr, errors.New("source file is not a regular file")
	}

	// Check if dst is a directory, and adjust the destination path if it is
	dstStat, err := os.Stat(dst)
	if err == nil {
		if dstStat.IsDir() {
			dst = filepath.Join(dst, filepath.Base(src))
			// Check adjusted destination path
			dstStat, err = os.Stat(dst)
		}
	}
	if err != nil && !os.IsNotExist(err) {
		// for errors other than file not exists
		return emptyStr, err
	}

	// for destination file exists
	if err == nil {
		// If the source and destination files are the same, return an error.
		if os.SameFile(srcStat, dstStat) {
			return emptyStr, fmt.Errorf("source and destination are the same file: %s", src)
		}
		// If overwrite is false, return an error if the destination file exists.
		if !overwrite {
			return emptyStr, &os.PathError{Op: "copy", Path: dst, Err: os.ErrExist}
		}
	}

	// Create the destination file.
	dstFile, err := os.Create(dst)
	if err != nil {
		return emptyStr, fmt.Errorf("cannot create file: %w", err)
	}
	defer dstFile.Close()

	// Copy the source file to the destination file.
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return emptyStr, fmt.Errorf("cannot copy file: %w", err)
	}

	// Attempt to set the mode, times file to match the source file, i.e. ignore the errors
	_ = os.Chmod(dst, srcStat.Mode())
	_ = os.Chtimes(dst, srcStat.ModTime(), srcStat.ModTime())
	return dst, nil
}
