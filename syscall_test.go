package memfd

import (
	"os"
	"syscall"
	"testing"
)

func TestSyscalls(t *testing.T) {
	fd, err := SyscallMemfdCreate("test", MFD_CLOEXEC|MFD_ALLOW_SEALING)
	if err != nil {
		t.Errorf("SyscallMemfdCreate failed: %v", err)
	}
	seals, err := SyscallFcntlSeals(fd)
	if err != nil {
		t.Errorf("SyscallFcntlSeals failed: %v", err)
	}
	if seals != 0 {
		t.Errorf("Expected no seals initially, got %d", seals)
	}
	err = SyscallFcntlSetSeals(fd, F_SEAL_SHRINK)
	if err != nil {
		t.Errorf("SyscallFcntlSetSeals failed: %v", err)
	}
	seals, err = SyscallFcntlSeals(fd)
	if err != nil {
		t.Errorf("SyscallFcntlSeals failed: %v", err)
	}
	if seals != F_SEAL_SHRINK {
		t.Errorf("Expected shrink seal, got %d", seals)
	}
	syscall.Close(int(fd))
}

func TestNotMemfdGetSeals(t *testing.T) {
	file, err := os.Open("/dev/null")
	if err != nil {
		t.Errorf("Cannot open /dev/null")
	}
	fd := file.Fd()
	_, err = SyscallFcntlSeals(fd)
	if err == nil {
		t.Errorf("Expected an error getting seals from /dev/null")
	}
	if err.Error() != "fcntl: invalid argument" {
		t.Errorf("Expected to get EINVAL reading seals from /dev/null got: %s", err.Error())
	}
}
