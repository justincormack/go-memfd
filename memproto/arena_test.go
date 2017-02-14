package memproto

import (
	"testing"

	air "github.com/justincormack/go-memfd/memproto/internal/aircraftlib"
	capnp "zombiezen.com/go/capnproto2"
)

func TestArena(t *testing.T) {
	arena, err := Create()
	if err != nil {
		t.Errorf("failed to create memfd arena: %v", err)
	}
	defer arena.Close()

	_, s, err := capnp.NewMessage(arena)
	if err != nil {
		t.Errorf("allocation error: %v", err)
	}
	d, err := air.NewRootZdate(s)
	if err != nil {
		t.Errorf("root error %v", err)
	}
	d.SetYear(2004)
	d.SetMonth(12)
	d.SetDay(7)

	arena.Unmap()

	err = arena.SetImmutable()
	if err != nil {
		t.Errorf("make Immutable failed")
	}

	msg := arena.Message()

	d, err = air.ReadRootZdate(msg)
	if err != nil {
		t.Errorf("read root error %v", err)
	}
	if d.Year() != 2004 {
		t.Errorf("Incorrect year: %d", d.Year())
	}
	if d.Month() != 12 {
		t.Errorf("Incorrect month: %d", d.Month())
	}
	if d.Day() != 7 {
		t.Errorf("Incorrect day: %d", d.Day())
	}
}
