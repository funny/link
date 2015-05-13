package link

import (
	"sync/atomic"
	"unsafe"
)

type MemPool struct {
	classes []*memClass
	min     int
	max     int
	off     int
}

func NewMemPool(total /* M */, min /* K */, max /* K */ int) *MemPool {
	total *= 1024 * 1024
	num := max - min + 1
	each := total / num
	classes := make([]*memClass, num)

	for i, j := min, 0; i <= max; i, j = i+1, j+1 {
		size := i * 1024
		items := each / size
		c := &memClass{maxlen: int32(items)}
		for k := 0; k < items; k++ {
			c.Push(&mem{Data: make([]byte, size)})
		}
		classes[j] = c
	}

	min *= 1024
	max *= 1024
	return &MemPool{classes, min, max, min - 1023}
}

func (pool *MemPool) Alloc(size, capacity int) *mem {
	if capacity <= pool.max {
		if capacity < pool.min {
			capacity = pool.min
		}
		m := pool.classes[(capacity-pool.off)/1024].Pop()
		if m != nil {
			return m
		}
	}
	return &mem{Data: make([]byte, size, capacity)}
}

func (pool *MemPool) Free(m *mem) {
	if cap(m.Data) >= pool.min && cap(m.Data) <= pool.max {
		pool.classes[(cap(m.Data)-pool.off)/1024].Push(m)
	}
}

type memClass struct {
	head   unsafe.Pointer
	length int32
	maxlen int32
}

type mem struct {
	Data []byte
	pool *MemPool
	next unsafe.Pointer
}

func (class *memClass) Push(item *mem) {
	if atomic.LoadInt32(&class.length) >= class.maxlen {
		return
	}
	for {
		item.next = atomic.LoadPointer(&class.head)
		if atomic.CompareAndSwapPointer(&class.head, item.next, unsafe.Pointer(item)) {
			atomic.AddInt32(&class.length, 1)
			break
		}
	}
}

func (class *memClass) Pop() *mem {
	var ptr unsafe.Pointer
	for {
		ptr = atomic.LoadPointer(&class.head)
		if ptr == nil {
			break
		}
		if atomic.CompareAndSwapPointer(&class.head, ptr, ((*mem)(ptr)).next) {
			atomic.AddInt32(&class.length, -1)
			break
		}
	}
	return (*mem)(ptr)
}
