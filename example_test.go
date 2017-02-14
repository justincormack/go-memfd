package memfd_test

import (
	"fmt"

	"github.com/justincormack/go-memfd"
)

func Example() {
	mfd, err := memfd.Create()
	if err != nil {
		panic(err)
	}
	defer mfd.Close()
	_, _ = mfd.WriteString("add some contents")
	err = mfd.SetImmutable()
	if err != nil {
		panic(err)
	}
	b, err := mfd.Map()
	if err != nil {
		panic(err)
	}
	defer mfd.Unmap()
	fmt.Println(string(b))
	// Output: add some contents
}
