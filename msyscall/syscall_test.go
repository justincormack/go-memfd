package msyscall

import (
	"os"
	"syscall"
	"testing"
)

func TestSyscalls(t *testing.T) {
	fd, err := MemfdCreate("test", MFD_CLOEXEC|MFD_ALLOW_SEALING)
	if err != nil {
		t.Errorf("MemfdCreate failed: %v", err)
	}
	defer syscall.Close(int(fd))
	seals, err := FcntlSeals(fd)
	if err != nil {
		t.Errorf("FcntlSeals failed: %v", err)
	}
	if seals != 0 {
		t.Errorf("Expected no seals initially, got %d", seals)
	}
	err = FcntlSetSeals(fd, F_SEAL_SHRINK)
	if err != nil {
		t.Errorf("FcntlSetSeals failed: %v", err)
	}
	seals, err = FcntlSeals(fd)
	if err != nil {
		t.Errorf("FcntlSeals failed: %v", err)
	}
	if seals != F_SEAL_SHRINK {
		t.Errorf("Expected shrink seal, got %d", seals)
	}
}

func TestCloexec(t *testing.T) {
	fd, err := MemfdCreate("test", MFD_CLOEXEC|MFD_ALLOW_SEALING)
	if err != nil {
		t.Errorf("MemfdCreate failed: %v", err)
	}
	defer syscall.Close(int(fd))
	err = FcntlClearCloexec(fd)
	if err != nil {
		t.Errorf("FcntlClearCloexec failed: %v", err)
	}
}

func TestNotMemfdGetSeals(t *testing.T) {
	file, err := os.Open("/dev/null")
	if err != nil {
		t.Errorf("Cannot open /dev/null")
	}
	defer file.Close()
	fd := file.Fd()
	_, err = FcntlSeals(fd)
	if err == nil {
		t.Errorf("Expected an error getting seals from /dev/null")
	}
	if err.Error() != "fcntl: invalid argument" {
		t.Errorf("Expected to get EINVAL reading seals from /dev/null got: %s", err.Error())
	}
}
