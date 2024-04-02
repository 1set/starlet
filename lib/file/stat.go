package file

import (
	"fmt"
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
			modeStr = "char_device"
		} else {
			modeStr = "block_device"
		}
	case mode&os.ModeIrregular != 0:
		modeStr = "irregular"
	default:
		modeStr = "unknown"
	}
	// create struct
	fileName := f.Name()
	fields := starlark.StringDict{
		"name":     starlark.String(fileName),
		"path":     starlark.String(f.fullPath),
		"ext":      starlark.String(filepath.Ext(fileName)),
		"size":     starlark.MakeInt64(f.Size()),
		"type":     starlark.String(modeStr),
		"modified": stdtime.Time(f.ModTime()),
	}
	return starlarkstruct.FromStringDict(starlark.String("file_stat"), fields)
}

func getFileStat(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var inputPath string
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "name", &inputPath); err != nil {
		return starlark.None, err
	}
	// get file stat
	stat, err := os.Stat(inputPath)
	if err != nil {
		return none, fmt.Errorf("%s: %v", b.Name(), err)
	}
	// get file abs path
	absPath, err := filepath.Abs(inputPath)
	if err != nil {
		return none, fmt.Errorf("%s: %v", b.Name(), err)
	}
	// return file stat
	fs := &FileStat{stat, absPath}
	return fs.Struct(), nil
}
