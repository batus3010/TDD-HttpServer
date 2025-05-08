package poker

import (
	"io"
	"os"
)

// Tape encapsulates the process of "when we write data we go from the beginning"
type Tape struct {
	File *os.File
}

func (t *Tape) Write(p []byte) (n int, err error) {
	t.File.Truncate(0)
	t.File.Seek(0, io.SeekStart)
	return t.File.Write(p)
}
