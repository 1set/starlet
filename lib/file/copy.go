package file

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func copyfile(src, dst string, override bool) error {
	// Open the source file.
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open source file: %w", err)
	}
	defer srcFile.Close()

	// Stat the source file to get its mode, times, and owner.
	srcStat, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("stat source file: %w", err)
	}
	if !srcStat.Mode().IsRegular() {
		// HACK, not sure if this is the best way to check if the file is a regular file
		return errors.New("source file is not a regular file")
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
		return err
	}

	// for destination file exists
	if err == nil {
		// If override is false, return an error if the destination file exists.
		if !override {
			return &os.PathError{Op: "copy", Path: dst, Err: os.ErrExist}
		}
		// If the source and destination files are the same, return an error.
		if os.SameFile(srcStat, dstStat) {
			return fmt.Errorf("source and destination are the same file: %s", src)
		}
	}

	// Create the destination file.
	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("create destination file: %w", err)
	}
	defer dstFile.Close()

	// Copy the source file to the destination file.
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("copy file: %w", err)
	}

	// Set the mode, times file to match the source file. HACK: ignore the error maybe
	if err := os.Chmod(dst, srcStat.Mode()); err != nil {
		return fmt.Errorf("chmod destination file: %w", err)
	}
	if err := os.Chtimes(dst, srcStat.ModTime(), srcStat.ModTime()); err != nil {
		return fmt.Errorf("chtimes destination file: %w", err)
	}
	return nil
}
