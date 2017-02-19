package memfd

import (
	"os"
	"testing"
)

func TestCreate(t *testing.T) {
	mfd, err := Create()
	if err != nil {
		t.Errorf("Create failed: %v", err)
	}
	mfd.Close()
}

func TestCreateFlags(t *testing.T) {
	mfd, err := CreateNameFlags("test", Cloexec)
	if err != nil {
		t.Errorf("CreateFlags failed: %v", err)
	}
	mfd.Close()
}

func TestSealing(t *testing.T) {
	mfd, err := Create()
	if err != nil {
		t.Errorf("Create failed: %v", err)
	}
	defer mfd.Close()
	seals := mfd.Seals()
	if seals != 0 {
		t.Errorf("Expected no seals initially, got %d", seals)
	}
	if mfd.IsImmutable() {
		t.Errorf("Expected not fully sealed initially")
	}
	err = mfd.SetSeals(SealWrite)
	if err != nil {
		t.Errorf("SetSeals failed: %v", err)
	}
	seals = mfd.Seals()
	if seals != SealWrite {
		t.Errorf("Expected write seal, got %d", seals)
	}
	if mfd.IsImmutable() {
		t.Errorf("Expected not fully immutable after setting write seal")
	}
	err = mfd.SetImmutable()
	if err != nil {
		t.Errorf("SetImmutable failed: %v", err)
	}
	if !mfd.IsImmutable() {
		t.Errorf("Expected fully immutable after setting immutable")
	}
	seals = mfd.Seals()
	if seals != SealAll {
		t.Errorf("Expected all seals set after setting immutable: got %d", seals)
	}
}

func TestResize(t *testing.T) {
	mfd, err := Create()
	if err != nil {
		t.Errorf("Create failed: %v", err)
	}
	defer mfd.Close()
	err = mfd.SetSize(1024)
	if err != nil {
		t.Errorf("Grow failed: %v", err)
	}
	err = mfd.SetSeals(SealShrink)
	if err != nil {
		t.Errorf("SetSeals failed: %v", err)
	}
	err = mfd.SetSize(2048)
	if err != nil {
		t.Errorf("Grow failed: %v", err)
	}
	err = mfd.SetSize(0)
	if err == nil {
		t.Errorf("Shrink succeeded after seal")
	}
	err = mfd.SetSeals(SealGrow)
	if err != nil {
		t.Errorf("SetSeals failed: %v", err)
	}
	err = mfd.SetSize(4096)
	if err == nil {
		t.Errorf("Grow succeeded after seal")
	}
}

func TestImmutable(t *testing.T) {
	mfd, err := Create()
	if err != nil {
		t.Errorf("Create failed: %v", err)
	}
	defer mfd.Close()
	err = mfd.SetImmutable()
	if err != nil {
		t.Errorf("SetImmutable failed: %v", err)
	}
	err = mfd.SetSize(1024)
	if err == nil {
		t.Errorf("Resize succeeded after seal")
	}
}

func TestMap(t *testing.T) {
	mfd, err := Create()
	if err != nil {
		t.Errorf("Create failed: %v", err)
	}
	defer mfd.Close()
	text := "Putting something in the memfd"
	n, err := mfd.WriteString(text)
	if err != nil {
		t.Errorf("Write error: %v", err)
	}
	if n != len(text) {
		t.Errorf("Short write")
	}
	b, err := mfd.Map()
	if err != nil {
		t.Errorf("Map read write error: %v", err)
	}
	if string(b) != text {
		t.Errorf("Did not read previous write: %s", string(b))
	}
	err = mfd.SetImmutable()
	if err == nil {
		t.Errorf("Should not be able to set immutable while there is a writeable mapping")
	}
	err = mfd.Unmap()
	if err != nil {
		t.Errorf("Unmap error: %v", err)
	}
	err = mfd.SetImmutable()
	if err != nil {
		t.Errorf("SetImmutable failed: %v", err)
	}
	b, err = mfd.Map()
	if err != nil {
		t.Errorf("Map read only error: %v", err)
	}
	err = mfd.Unmap()
	if err != nil {
		t.Errorf("Unmap error: %v", err)
	}
}

func TestMapZero(t *testing.T) {
	mfd, err := Create()
	if err != nil {
		t.Errorf("Create failed: %v", err)
	}
	defer mfd.Close()
	b, err := mfd.Map()
	if err != nil {
		t.Errorf("Map error: %v", err)
	}
	if cap(b) != 0 {
		t.Errorf("Expected zero capacity map, got %d", cap(b))
	}
	if len(b) != 0 {
		t.Errorf("Expected zero length map, got %d", len(b))
	}
	err = mfd.Unmap()
	if err != nil {
		t.Errorf("Unmap error: %v", err)
	}
}

func TestRemap(t *testing.T) {
	mfd, err := Create()
	if err != nil {
		t.Errorf("Create failed: %v", err)
	}
	defer mfd.Close()
	err = mfd.SetSize(100)
	if err != nil {
		t.Errorf("Resize failed: %v", err)
	}
	b, err := mfd.Map()
	if err != nil {
		t.Errorf("Map error: %v", err)
	}
	if cap(b) != 100 {
		t.Errorf("Expected 100 capacity map, got %d", cap(b))
	}
	if len(b) != 100 {
		t.Errorf("Expected 100 length map, got %d", len(b))
	}
	mfd.SetSize(200)
	if cap(b) != 100 {
		t.Errorf("Expected 100 capacity map, got %d", cap(b))
	}
	if len(b) != 100 {
		t.Errorf("Expected 100 length map, got %d", len(b))
	}
	b, err = mfd.Remap()
	if err != nil {
		t.Errorf("Remap error: %v", err)
	}
	if cap(b) != 200 {
		t.Errorf("Expected 200 capacity map, got %d", cap(b))
	}
	if len(b) != 200 {
		t.Errorf("Expected 200 length map, got %d", len(b))
	}
	err = mfd.Unmap()
	if err != nil {
		t.Errorf("Unmap error: %v", err)
	}
}

func TestNew(t *testing.T) {
	mfd0, err := Create()
	if err != nil {
		t.Errorf("Create failed: %v", err)
	}
	defer mfd0.Close()
	fd := mfd0.Fd()
	mfd, err := New(uintptr(fd))
	if err != nil {
		t.Errorf("New failed: %v", err)
	}
	mfd.Close()
}

func TestNotMemfdNewMemfd(t *testing.T) {
	file, err := os.Open("/dev/null")
	if err != nil {
		t.Errorf("Cannot open /dev/null")
	}
	fd := file.Fd()
	_, err = New(fd)
	if err == nil {
		t.Errorf("Expected an error calling New with /dev/null")
	}
}

func TestCloexec(t *testing.T) {
	mfd, err := Create()
	if err != nil {
		t.Errorf("Create failed: %v", err)
	}
	defer mfd.Close()
	// does not really test it is working at present
	mfd.SetCloexec()
	mfd.ClearCloexec()
}
