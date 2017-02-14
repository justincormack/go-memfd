package memproto_test

import (
	"fmt"
	"log"

	"github.com/justincormack/go-memfd/memproto"
	air "github.com/justincormack/go-memfd/memproto/internal/aircraftlib"
	capnp "zombiezen.com/go/capnproto2"
)

func writer() uintptr {
	arena, err := memproto.Create()
	if err != nil {
		log.Fatalf("failed to create memfd arena: %v\n", err)
	}
	// do not Close as we want to keep fd to pass to reader, but
	// unmap as required before we can make immutable.
	defer arena.Unmap()

	_, s, err := capnp.NewMessage(arena)
	if err != nil {
		log.Fatalf("allocation error: %v\n", err)
	}
	d, err := air.NewRootZdate(s)
	if err != nil {
		log.Fatalf("root error %v\n", err)
	}
	d.SetYear(2004)
	d.SetMonth(12)
	d.SetDay(7)

	return arena.Fd()
}

func reader(fd uintptr) {
	arena, err := memproto.New(fd)
	if err != nil {
		log.Fatalf("File descriptor did not appear to be a memfd: %v", err)
	}
	defer arena.Close()

	err = arena.SetImmutable()
	if err != nil {
		log.Fatalf("make immutable failed")
	}

	msg := arena.Message()

	d, err := air.ReadRootZdate(msg)
	if err != nil {
		log.Fatalf("read root error %v\n", err)
	}
	fmt.Printf("year %d, month %d, day %d\n", d.Year(), d.Month(), d.Day())
}

// In reality reader and writer would be in different processes.
func Example() {
	reader(writer())
	// Output: year 2004, month 12, day 7
}
