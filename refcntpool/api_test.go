package refcntpool

import (
	"fmt"
	"testing"
	"unsafe"
)

func TestMain(t *testing.T) {
	fmt.Println("Hello, playground")

	var (
		bufPool *RefCntBufPool = RefCntBufPoolMake(65536)
		bufp    *RefCntBuf     = bufPool.Get().(*RefCntBuf)
		arr64k  [65536]byte
	)
	fmt.Printf("sizeof(arr64k) = %d\n", unsafe.Sizeof(arr64k))
	fmt.Printf("sizeof(*bufp) = %d\n", unsafe.Sizeof(*bufp))

	fmt.Printf("type(bufp) = %T\n", bufp)
	fmt.Printf("val(bufp) = %v\n", bufp)
	fmt.Printf("type(*bufp) = %T\n", *bufp)
	fmt.Printf("val(*bufp) = %v\n", *bufp)

	bufp.Hold()
	bufp.Release()
	bufp.Release()
}
