package memfd

import (
	"os"
	"testing"
)

func TestCreate(t *testing.T) {
	mfd, err := Create("test")
	if err != nil {
		t.Errorf("Create failed: %v", err)
	}
	mfd.Close()
}

func TestCreateFlags(t *testing.T) {
	mfd, err := CreateFlags("test", MFD_CLOEXEC)
	if err != nil {
		t.Errorf("CreateFlags failed: %v", err)
	}
	mfd.Close()
}

func TestSealing(t *testing.T) {
	mfd, err := Create("test")
	if err != nil {
		t.Errorf("Create failed: %v", err)
	}
	seals := mfd.Seals()
	if seals != 0 {
		t.Errorf("Expected no seals initially, got %d", seals)
	}
	if mfd.IsImmutable() {
		t.Errorf("Expected not fully sealed initially")
	}
	err = mfd.SetSeals(F_SEAL_WRITE)
	if err != nil {
		t.Errorf("SetSeals failed: %v", err)
	}
	seals = mfd.Seals()
	if seals != F_SEAL_WRITE {
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
	mfd.Close()
}

func TestResize(t *testing.T) {
	mfd, err := Create("test")
	if err != nil {
		t.Errorf("Create failed: %v", err)
	}
	err = mfd.SetSize(1024)
	if err != nil {
		t.Errorf("Grow failed: %v", err)
	}
	err = mfd.SetSeals(F_SEAL_SHRINK)
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
	err = mfd.SetSeals(F_SEAL_GROW)
	if err != nil {
		t.Errorf("SetSeals failed: %v", err)
	}
	err = mfd.SetSize(4096)
	if err == nil {
		t.Errorf("Grow succeeded after seal")
	}
	mfd.Close()
}

func TestImmutable(t *testing.T) {
	mfd, err := Create("test")
	if err != nil {
		t.Errorf("Create failed: %v", err)
	}
	err = mfd.SetImmutable()
	if err != nil {
		t.Errorf("SetImmutable failed: %v", err)
	}
	err = mfd.SetSize(1024)
	if err == nil {
		t.Errorf("Resize succeeded after seal")
	}
	mfd.Close()
}

func TestMap(t *testing.T) {
	mfd, err := Create("test")
	if err != nil {
		t.Errorf("Create failed: %v", err)
	}
	text := "Putting something in the memfd"
	n, err := mfd.WriteString(text)
	if err != nil {
		t.Errorf("Write error: %v", err)
	}
	if n != len(text) {
		t.Errorf("Short write")
	}
	b, err := mfd.MapRW()
	if err != nil {
		t.Errorf("MapRW error: %v", err)
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
	mfd.Close()
}

func TestNewMemfd(t *testing.T) {
	fd, err := SyscallMemfdCreate("test", MFD_CLOEXEC|MFD_ALLOW_SEALING)
	if err != nil {
		t.Errorf("SyscallMemfdCreate failed: %v", err)
	}
	mfd, err := NewMemfd(uintptr(fd))
	if err != nil {
		t.Errorf("NewMemfd failed: %v", err)
	}
	mfd.Close()
}

func TestNotMemfdNewMemfd(t *testing.T) {
	file, err := os.Open("/dev/null")
	if err != nil {
		t.Errorf("Cannot open /dev/null")
	}
	fd := file.Fd()
	_, err = NewMemfd(fd)
	if err == nil {
		t.Errorf("Expected an error calling NewMemfd with /dev/null")
	}
}
