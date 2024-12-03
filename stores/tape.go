package stores

import (
	"io"
	"os"
)

// A tape wraps an os.File and writes from the start of the file on each time.
type tape struct {
	file *os.File
}

func (t *tape) Write(b []byte) (int, error) {
	err := t.file.Truncate(0)

	if err != nil {
		return 0, err
	}

	_, err = t.file.Seek(0, io.SeekStart)

	if err != nil {
		return 0, err
	}

	return t.file.Write(b)
}
