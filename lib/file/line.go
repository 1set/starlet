package file

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
)

// LineFunc stands for a handler for each line string.
type LineFunc func(line string) (err error)

//revive:disable:error-naming It's not a real error
var (
	// QuitRead indicates the arbitrary error means to quit from reading.
	QuitRead = errors.New("file: quit read by line")
)

// ReadFileLines reads all lines from the given file (the line ending chars are not included).
func ReadFileLines(path string) (lines []string, err error) {
	err = readFileByLine(path, func(l string) error {
		lines = append(lines, l)
		return nil
	})
	return
}

// CountFileLines counts all lines from the given file (the line ending chars are not included).
func CountFileLines(path string) (count int, err error) {
	err = readFileByLine(path, func(l string) error {
		count++
		return nil
	})
	return
}

// WriteFileLines writes the given lines as a text file.
func WriteFileLines(path string, lines []string) error {
	return openFileWriteLines(path, createFileFlag, lines)
}

// AppendFileLines appends the given lines to the end of a text file.
func AppendFileLines(path string, lines []string) error {
	return openFileWriteLines(path, appendFileFlag, lines)
}

// ReadFirstLines reads the top n lines from the given file (the line ending chars are not included), or lesser lines if the given file doesn't contain enough line ending chars.
func ReadFirstLines(path string, n int) (lines []string, err error) {
	var f *os.File
	if f, err = os.Open(path); err != nil {
		return
	}
	defer f.Close()
	return extractIOTopLines(f, n)
}

// ReadLastLines reads the bottom n lines from the given file (the line ending chars are not included), or lesser lines if the given file doesn't contain enough line ending chars.
func ReadLastLines(path string, n int) (lines []string, err error) {
	var f *os.File
	if f, err = os.Open(path); err != nil {
		return
	}
	defer f.Close()
	return extractIOBottomLines(f, n)
}

// extractIOTopLines extracts the top n lines from the given stream (the line ending chars are not included), or lesser lines if the given stream doesn't contain enough line ending chars.
func extractIOTopLines(rd io.Reader, n int) ([]string, error) {
	if n <= 0 {
		return nil, errors.New("amoy: n should be greater than 0")
	}
	result := make([]string, 0)
	if err := readIOByLine(rd, func(line string) error {
		result = append(result, line)
		n--
		if n <= 0 {
			return QuitRead
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return result, nil
}

// extractIOBottomLines extracts the bottom n lines from the given stream (the line ending chars are not included), or lesser lines if the given stream doesn't contain enough line ending chars.
func extractIOBottomLines(rd io.Reader, n int) ([]string, error) {
	if n <= 0 {
		return nil, errors.New("amoy: n should be greater than 0")
	}
	var (
		result = make([]string, n, n)
		cnt    int
	)
	if err := readIOByLine(rd, func(line string) error {
		result[cnt%n] = line
		cnt++
		return nil
	}); err != nil {
		return nil, err
	}
	if cnt <= n {
		return result[0:cnt], nil
	}
	pos := cnt % n
	return append(result[pos:], result[0:pos]...), nil
}

// readFileByLine iterates the given file by lines (the line ending chars are not included).
func readFileByLine(path string, callback LineFunc) (err error) {
	var file *os.File
	if file, err = os.Open(path); err != nil {
		return
	}
	defer file.Close()
	return readIOByLine(file, callback)
}

func openFileWriteLines(path string, flag int, lines []string) error {
	file, err := os.OpenFile(path, flag, filePerm)
	if err != nil {
		return err
	}
	defer file.Close()
	return writeIOLines(file, lines)
}

// writeIOLines writes the given lines to a Writer.
func writeIOLines(wr io.Writer, lines []string) error {
	w := bufio.NewWriter(wr)
	defer w.Flush()
	for _, line := range lines {
		if _, err := fmt.Fprintln(w, line); err != nil {
			return err
		}
	}
	return nil
}

// readIOByLine iterates the given Reader by lines (the line ending chars are not included).
func readIOByLine(rd io.Reader, callback LineFunc) (err error) {
	readLine := func(r *bufio.Reader) (string, error) {
		var (
			err      error
			line, ln []byte
			isPrefix = true
		)
		for isPrefix && err == nil {
			line, isPrefix, err = r.ReadLine()
			ln = append(ln, line...)
		}
		return string(ln), err
	}
	r := bufio.NewReader(rd)
	s, e := readLine(r)
	for e == nil {
		if err = callback(s); err != nil {
			break
		}
		s, e = readLine(r)
	}

	if err == QuitRead {
		err = nil
	}
	return
}
