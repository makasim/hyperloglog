package hyperloglog

import (
	"fmt"
	"sync"
)

type SketchPool struct {
	p uint8
	s bool

	mu sync.Mutex
	sp *sync.Pool
}

func NewSketchPool(precision uint8, sparse bool) *SketchPool {
	return &SketchPool{
		p: precision,
		s: sparse,

		sp: &sync.Pool{},
	}
}

func (skp *SketchPool) Get() (*Sketch, error) {
	skp.mu.Lock()
	defer skp.mu.Unlock()

	sk := skp.sp.Get()
	if sk == nil {
		return NewSketch(skp.p, skp.s)
	}

	return sk.(*Sketch), nil
}

func (skp *SketchPool) MustGet() *Sketch {
	sk, err := skp.Get()
	if err != nil {
		panic(err)
	}

	return sk
}

func (skp *SketchPool) Put(sk *Sketch) error {
	if sk == nil {
		return nil
	}
	if sk.p != skp.p {
		return fmt.Errorf("sketch with precision %d could not be placed into pool with precision %d", sk.p, skp.p)
	}

	sk.Reset()

	skp.mu.Lock()
	skp.sp.Put(sk)
	skp.mu.Unlock()
	return nil
}

func (skp *SketchPool) MustPut(sk *Sketch) {
	if err := skp.Put(sk); err != nil {
		panic(err)
	}
}
