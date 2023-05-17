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

func (c *Counter) add(amount int32) uint32 {
	f, err := lockedfile.OpenFile(c.file, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	var data = make([]byte, 4)
	var current uint32
	f.Read(data)

	// For simplicitly, no checks on the bounds.
	current = binary.BigEndian.Uint32(data)
	if amount < 0 {
		current -= uint32(amount)
	} else {
		current += uint32(amount)
	}

	binary.BigEndian.PutUint32(data, current)
	_, err = f.WriteAt(data, 0)
	if err != nil {
		panic(err)
	}
	return current
}

// Inc increases the counter value by 1.
func (c *Counter) Inc() uint32 {
	return c.add(1)
}

// Dec decreases the counter value by 1.
func (c *Counter) Dec() uint32 {
	return c.add(-1)
}
