package counter

import (
	"encoding/binary"
	"os"

	"github.com/rogpeppe/go-internal/lockedfile"
)

type Counter struct {
	file string
}

func NewCounter(file string) *Counter {
	return &Counter{
		file: file,
	}
}

func (c *Counter) Inc() uint32 {
	f, err := lockedfile.OpenFile(c.file, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	var data = make([]byte, 4)
	var current uint32
	f.Read(data)

	current = binary.BigEndian.Uint32(data)
	current++

	binary.BigEndian.PutUint32(data, current)
	_, err = f.WriteAt(data, 0)
	if err != nil {
		panic(err)
	}
	return current
}
