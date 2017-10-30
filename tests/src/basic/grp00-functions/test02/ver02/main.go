// Version 02:
package main

import "fmt"

type Buffer struct {
	data   []byte
	reader func(*Buffer, int) (byte, error)
	writer func(*Buffer, int, byte) error
}

type ReaderFunc func(*Buffer, int) (byte, error)
type WriterFunc func(*Buffer, int, byte) error

func (b *Buffer) Length() (l int, e error) {
	if b.data == nil {
		return 0, fmt.Errorf("Uninitialized buffer!")
	}
	return len(b.data), nil
}

func (b *Buffer) Capacity() (l int, e error) {
	if b.data == nil {
		return 0, fmt.Errorf("Uninitialized buffer!")
	}
	return cap(b.data), nil
}

func GenericReader(b *Buffer, pos int) (v byte, e error) {
	l, e := b.Length()
	if e != nil {
		return 0, e
	}
	if pos < 0 || pos >= l {
		return 0, fmt.Errorf("%v is out of range!", pos)
	}
	return b.data[pos], nil
}

func GenericWriter(b *Buffer, pos int, v byte) error {
	l, e := b.Length()
	if e != nil {
		return e
	}
	if pos < 0 || pos >= l {
		return fmt.Errorf("%v is out of range!", pos)
	}
	b.data[pos] = v
	return nil
}

func NewBuffer(sz int, reader ReaderFunc, writer WriterFunc) *Buffer {
	if reader == nil {
		reader = GenericReader
	}
	if writer == nil {
		writer = GenericWriter
	}
	return &Buffer{
		data: make([]byte, sz),
		reader: reader, writer: writer,
	}
}

var unit_buffer = NewBuffer(1, nil, nil)

func main() {
	b := NewBuffer(16, nil, nil)
	ub := unit_buffer
        var (
		r ReaderFunc
		w WriterFunc
	)
	if os.Args[0] == "x" {
		r = ub.reader
		w = ub.writer
	} else {
		r = b.reader
		w = b.writer
	}
	_ := w(b, 0, 42)
	v, _ := r(b, 0)
	fmt.Printf("%v\n", v)
}
/*
  Sequences:

    (TODO)
*/
