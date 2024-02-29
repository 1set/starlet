package file

import (
	"bufio"
	"os"
)

var (
	emptyStr       string
	filePerm       os.FileMode = 0644
	createFileFlag             = os.O_RDWR | os.O_CREATE | os.O_TRUNC
	appendFileFlag             = os.O_APPEND | os.O_CREATE | os.O_WRONLY
)

// TrimUTF8BOM removes the leading UTF-8 byte order mark from bytes.
func TrimUTF8BOM(b []byte) []byte {
	if len(b) >= 3 && b[0] == 0xef && b[1] == 0xbb && b[2] == 0xbf {
		return b[3:]
	}
	return b
}

// ReadFileBytes reads the whole named file and returns the contents.
// It's a sugar actually, simply calls os.ReadFile like ioutil.ReadFile does since Go 1.16.
func ReadFileBytes(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// ReadFileString reads the whole named file and returns the contents as a string.
func ReadFileString(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return emptyStr, err
	}
	return string(b), nil
}

// WriteFileBytes writes the given data into a file.
func WriteFileBytes(path string, data []byte) error {
	return openFileWriteBytes(path, createFileFlag, data)
}

// WriteFileString writes the given content string into a file.
func WriteFileString(path string, content string) error {
	return openFileWriteString(path, createFileFlag, content)
}

// AppendFileBytes writes the given data to the end of a file.
func AppendFileBytes(path string, data []byte) error {
	return openFileWriteBytes(path, appendFileFlag, data)
}

// AppendFileString appends the given content string to the end of a file.
func AppendFileString(path string, content string) error {
	return openFileWriteString(path, appendFileFlag, content)
}

func openFileWriteBytes(path string, flag int, data []byte) error {
	file, err := os.OpenFile(path, flag, filePerm)
	if err != nil {
		return err
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	defer w.Flush()
	_, err = w.Write(data)
	return err
}

func openFileWriteString(path string, flag int, content string) error {
	file, err := os.OpenFile(path, flag, filePerm)
	if err != nil {
		return err
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	defer w.Flush()
	_, err = w.WriteString(content)
	return err
}
