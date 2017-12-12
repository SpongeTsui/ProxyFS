package refcntpool

// refcntpool provides functions and interfaces to implement pools of reference
// counted objects, where the object is returned to the pool when its reference
// count drops to zero.
//
// Users of refcountpool must supply an object implementing the RefCntItemPool
// interface which makes objects of the desired type and maintains a pool
// for holding inactive objects of that type.
//
// An implementation of pools of reference counted memory buffers is also
// provided, and also serves as an example.  Use RefCntBufPoolMake(bufSz uint64)
// to create a pool of reference counted memory buffers of size bufSz.

import (
	"fmt"
	"sync"
	"sync/atomic"
)

// RefCntItemPool interface defines the Get() and put() methods for RefCntItem
// objects.
//
// While Get() is called to get a new object, put() should only be called via
// the object's Release() method and not called directly.
//
type RefCntItemPool interface {
	// Return an object of the type held by the pool which also supports
	// the RefCntItem methods (Hold() and Release())
	Get() interface{}

	// Put an object of the type held by the pool back in the pool.
	put(*RefCntItem)
}

// A RefCntItem is acquired from a RefCntItemPool. It has Hold() and Release()
// methods for the object.
//
// An object returned by Get() starts with one hold.  When all the holds are
// released the object should no longer be accessed.
//
type RefCntItem struct {
	pool   RefCntItemPool
	refCnt uint32 // updated atomically
}

func (item *RefCntItem) Hold() {
	newCnt := atomic.AddUint32(&item.refCnt, 1)
	if newCnt < 2 {
		panic(fmt.Sprintf("RefCntItem.Hold(): item was not held when called: newCnt %d", newCnt))
	}
}

func (item *RefCntItem) Release() {
	// Decrement cnt by 1.  Even if two threads do this concurrently,
	// only one will have newcnt == 0
	newCnt := atomic.AddUint32(&item.refCnt, ^uint32(0))

	if newCnt == 0 {
		item.pool.put(item)
	} else if int32(newCnt) < 0 {
		panic(fmt.Sprintf("RefCntItem.Release(): item was not held when called: newCnt %d", newCnt))
	}
}

// A reference counted memory buffer implementing Hold() and Release()
//
type RefCntBuf struct {
	RefCntItem        // track reference count; provides Hold() and Release()
	origBuf    []byte // original buffer allocation
	Buf        []byte // current buffer
}

// A pool of reference counted memory buffers, where bufers are acquired using
// Get() and returned on the final Relase() on the memory buffer.
//
type RefCntBufPool struct {
	bufPool sync.Pool // buffer pool
	bufSz   uint64    // all buffers in this pool are bufSz bytes
}

// Get a RefCntBuf from the pool.
//
// The caller must use a type assertion like (*refCntBufPool).Get().(*RefCntBuf)
// to get a pointer to the memory buffer.
//
func (poolp *RefCntBufPool) Get() (item interface{}) {

	// get a buffer
	item = poolp.bufPool.Get()

	// reinitialize the buffer
	bufp := item.(*RefCntBuf)
	newCnt := atomic.AddUint32(&bufp.refCnt, 1)
	bufp.Buf = bufp.origBuf

	if newCnt != 1 {
		panic(fmt.Sprintf("RefCntBufPool.Get(): item %p in pool was not free: newCnt %d", bufp, newCnt))
	}

	return
}

func (poolp *RefCntBufPool) put(itemp *RefCntItem) {
	poolp.bufPool.Put(itemp)
	return
}

// Create and return a pool of reference counted buffers with the defined size.
//
func RefCntBufPoolMake(size uint64) (poolp *RefCntBufPool) {
	poolp = &RefCntBufPool{}

	poolp.bufPool.New = func() interface{} {

		// Make a new RefCntBuf and initialize it
		bufp := &RefCntBuf{
			RefCntItem: RefCntItem{
				pool: poolp,
			},
			origBuf: make([]byte, 0, size),
		}
		return bufp
	}

	poolp.bufSz = size
	return
}
