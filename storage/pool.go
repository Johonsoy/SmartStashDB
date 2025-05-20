package storage

import (
	"bytes"
	"sync"
)

type bufferPool struct {
	buffer sync.Pool
}

func (p *bufferPool) Get() *bytes.Buffer {
	return p.buffer.Get().(*bytes.Buffer)
}

func (p *bufferPool) Put(buffer *bytes.Buffer) {
	p.buffer.Put(buffer)
}

var DefaultBuffer = newBufferPool()

func newBufferPool() *bufferPool {
	return &bufferPool{
		buffer: sync.Pool{
			New: func() any { return new(bytes.Buffer) },
		},
	}
}
