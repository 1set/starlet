package file

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"os"
	"path/filepath"

	stdtime "go.starlark.net/lib/time"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

// FileStat represents the file information.
type FileStat struct {
	os.FileInfo
	fullPath string
}

// Struct returns a starlark struct with file information.
func (f *FileStat) Struct() *starlarkstruct.Struct {
	// get file type
	var modeStr string
	switch mode := f.Mode(); {
	case mode.IsRegular():
		modeStr = "file"
	case mode.IsDir():
		modeStr = "dir"
	case mode&os.ModeSymlink != 0:
		modeStr = "symlink"
	case mode&os.ModeNamedPipe != 0:
		modeStr = "fifo"
	case mode&os.ModeSocket != 0:
		modeStr = "socket"
	case mode&os.ModeDevice != 0:
		if mode&os.ModeCharDevice != 0 {
			modeStr = "char"
		} else {
			modeStr = "block"
		}
	case mode&os.ModeIrregular != 0:
		modeStr = "irregular"
	default:
		modeStr = "unknown"
	}
	// create struct
	fileName := f.Name()
	fields := starlark.StringDict{
		"name":       starlark.String(fileName),
		"path":       starlark.String(f.fullPath),
		"ext":        starlark.String(filepath.Ext(fileName)),
		"size":       starlark.MakeInt64(f.Size()),
		"type":       starlark.String(modeStr),
		"modified":   stdtime.Time(f.ModTime()),
		"get_md5":    starlark.NewBuiltin("get_md5", genFileHashFunc(f.fullPath, md5.New)),
		"get_sha1":   starlark.NewBuiltin("get_sha1", genFileHashFunc(f.fullPath, sha1.New)),
		"get_sha256": starlark.NewBuiltin("get_sha256", genFileHashFunc(f.fullPath, sha256.New)),
		"get_sha512": starlark.NewBuiltin("get_sha512", genFileHashFunc(f.fullPath, sha512.New)),
	}
	return starlarkstruct.FromStringDict(starlark.String("file_stat"), fields)
}

func getFileStat(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var (
		inputPath     string
		followSymlink = false
	)
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "name", &inputPath, "follow?", &followSymlink); err != nil {
		return starlark.None, err
	}
	// get file stat
	var statFn func(string) (os.FileInfo, error)
	if followSymlink {
		statFn = os.Stat
	} else {
		statFn = os.Lstat
	}
	stat, err := statFn(inputPath)
	if err != nil {
		return none, fmt.Errorf("%s: %w", b.Name(), err)
	}
	// get file abs path
	absPath, err := filepath.Abs(inputPath)
	if err != nil {
		return none, fmt.Errorf("%s: %w", b.Name(), err)
	}
	// return file stat
	fs := &FileStat{stat, absPath}
	return fs.Struct(), nil
}

func genFileHashFunc(fp string, algo func() hash.Hash) func(*starlark.Thread, *starlark.Builtin, starlark.Tuple, []starlark.Tuple) (starlark.Value, error) {
	return func(t *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		// open file
		file, err := os.Open(fp)
		if err != nil {
			return none, fmt.Errorf("%s: %w", fn.Name(), err)
		}
		defer file.Close()

		// get hash
		h := algo()
		if _, err := io.Copy(h, file); err != nil {
			return none, fmt.Errorf("%s: %w", fn.Name(), err)
		}
		return starlark.String(hex.EncodeToString(h.Sum(nil))), nil
	}
}
