package memproto

import (
	"os"
	"strings"
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

func TestLargeArena(t *testing.T) {
	arena, err := Create()
	if err != nil {
		t.Errorf("failed to create memfd arena: %v", err)
	}
	defer arena.Close()

	_, s, err := capnp.NewMessage(arena)
	if err != nil {
		t.Errorf("allocation error: %v", err)
	}
	d, err := air.NewRootPlaneBase(s)
	if err != nil {
		t.Errorf("root error %v", err)
	}
	long := strings.Repeat("fly ", os.Getpagesize())
	d.SetName(long)
	d.SetCanFly(false)

	arena.Unmap()

	err = arena.SetImmutable()
	if err != nil {
		t.Errorf("make Immutable failed")
	}

	msg := arena.Message()

	d, err = air.ReadRootPlaneBase(msg)
	if err != nil {
		t.Errorf("read root error %v", err)
	}
	name, err := d.Name()
	if err != nil {
		t.Errorf("Expected a name")
	}
	if name != long {
		t.Errorf("Incorrect name: %s", name)
	}
	if d.CanFly() != false {
		t.Errorf("Incorrectly claims it can fly")
	}
}
