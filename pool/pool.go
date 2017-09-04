package pool

import (
	"fmt"
	"io"
	"log"
	"sync"
)

type Pool struct {
	m         *sync.RWMutex
	resources chan io.Closer
	closed    bool
	factory   func() (io.Closer, error)
}

func New(fn func() (io.Closer, error), size uint) (*Pool, error) {
	if size <= 0 {
		return nil, fmt.Errorf("pool size too small")
	}
	return &Pool{
		m:         &sync.RWMutex{},
		resources: make(chan io.Closer, size),
		closed:    true,
		factory:   fn,
	}, nil
}

func (p *Pool) Acquire() (io.Closer, error) {
	p.m.RLock()
	p.m.RUnlock()

	select {
	case r, ok := <-p.resources:
		if !ok {
			return nil, fmt.Errorf("pool has been closed")
		}
		return r, nil

	default:
		return p.factory()
	}
}

func (p *Pool) Release(r io.Closer) {
	p.m.Lock()
	defer p.m.Unlock()

	if p.closed {
		r.Close()
		return
	}

	select {
	case p.resources <- r:
		log.Println("release resource: ", r)

	default:
		log.Println("closing resource")
		r.Close()
	}
}

func (p *Pool) Close() {
	p.m.Lock()
	defer p.m.Unlock()

	if p.closed {
		return
	}

	p.closed = true

	close(p.resources)

	for r := range p.resources {
		r.Close()
	}
}
