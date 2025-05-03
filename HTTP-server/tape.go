package poker

import (
	"io"
	"os"
)

// tape encapsulates the process of "when we write data we go from the beginning"
type tape struct {
	file *os.File
}

func (t *tape) Write(p []byte) (n int, err error) {
	t.file.Truncate(0)
	t.file.Seek(0, io.SeekStart)
	return t.file.Write(p)
}
