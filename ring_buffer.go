package utp

import (
	"errors"
	"fmt"
	"sync"
)

type ringBuffer struct {
	b     [][]byte
	begin int
	seq   uint16
	cond  *sync.Cond
	m     sync.RWMutex
}

func NewRingBuffer(n, seq uint16) *ringBuffer {
	r := &ringBuffer{
		b:     make([][]byte, n),
		begin: 0,
		seq:   seq,
	}
	r.cond = sync.NewCond(&r.m)
	return r
}

func (r *ringBuffer) Pop() ([][]byte, error) {
	r.m.Lock()
	defer r.m.Unlock()
	fmt.Println("readable", r.readable(), r.begin, r.seq)
	for r.readable() == 0 {
		r.cond.Wait()
	}
	b := make([][]byte, 0, 1)
	for i := 0; i < len(r.b); i++ {
		c := r.b[r.begin]
		if c != nil {
			b = append(b, c)
		} else {
			break
		}
		r.begin = (r.begin + 1) % len(r.b)
		r.seq = uint16((int(r.seq) + 1) % 65536)
	}
	r.cond.Signal()
	return b, nil
}

func (r *ringBuffer) Put(b []byte, seq uint16) error {
	r.m.Lock()
	defer r.m.Unlock()
	i := r.getIndex(seq)
	if i < 0 {
		return errors.New("out of bounds")
	}
	r.b[i] = b
	r.cond.Signal()
	return nil
}

func (r *ringBuffer) getIndex(seq uint16) int {
	i := int(seq) - int(r.seq)
	if i < 0 {
		i += 65536
	}
	if i >= len(r.b) {
		return -1
	} else {
		return (i + r.begin) % len(r.b)
	}
}

func (r *ringBuffer) readable() int {
	i := 0
	for ; i < len(r.b); i++ {
		if r.b[(r.begin+i)%len(r.b)] == nil {
			break
		}
	}
	return i
}

func (r *ringBuffer) writable() int {
	i := 0
	for ; i < len(r.b); i++ {
		if r.b[len(r.b)-1-(r.begin+i)%len(r.b)] != nil {
			break
		}
	}
	return i
}

func (r *ringBuffer) Close() error {
	return nil
}