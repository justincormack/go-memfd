// +build linux

// Package memfd provides a Go library for working with Linux memfd memory file descriptors.
// This provides shareable anonymous memory which can be locked.
package memfd

import (
	"errors"
	"os"
	"syscall"
)

var (
	// ErrTooBig is returned if you try to map a memfd over 2GB on a 32 bit platform
	ErrTooBig = errors.New("memfd too large for slice")
)

// Memfd is the type for an memory fd, an os.File with extra methods
type Memfd struct {
	*os.File
	b []byte
}

// Create exposes the high level interface to memfd. It sets the flags to the
// options that you probably want. Name can be empty, it is just for reference.
func Create(name string) (*Memfd, error) {
	return CreateFlags(name, MFD_CLOEXEC|MFD_ALLOW_SEALING)
}

// CreateFlags exposes the high level interface to memfd, and allows setting
// flags if required. Name can be empty, it is just for reference.
func CreateFlags(name string, flags uint) (*Memfd, error) {
	fd, err := SyscallMemfdCreate(name, flags)
	if err != nil {
		return nil, err
	}
	memfd := Memfd{os.NewFile(uintptr(fd), name), []byte{}}
	return &memfd, nil
}

// NewMemfd creates a high level memfd object from a file descriptor, eg
// passed via a pipe or to an exec. Will return an error if the file was
// not a memfd, ie cannot have seals.
func NewMemfd(fd uintptr) (*Memfd, error) {
	_, err := SyscallFcntlSeals(fd)
	if err != nil {
		return nil, err
	}
	// TODO(justin) read name with readlink /proc/self/fd
	mfd := Memfd{os.NewFile(uintptr(fd), ""), []byte{}}
	return &mfd, nil
}

// Size returns the current size of the memfd
// It could return an error if the memfd is closed, so
// we return zero in that case.
func (mfd *Memfd) Size() int64 {
	fi, err := mfd.Stat()
	if err != nil {
		return 0
	}
	return fi.Size()
}

// SetSize sets the size of the memfd. It is just
// Truncate but a more understandable name.
func (mfd *Memfd) SetSize(size int64) error {
	return mfd.Truncate(size)
}

// seals is an internal function that returns seals or an error
func (mfd *Memfd) seals() (int, error) {
	return SyscallFcntlSeals(mfd.Fd())
}

// Seals returns the current seals. It can only error if something
// out of the ordinary has happened, eg the file is closed, so we
// just return 0 (no seals) in this case.
func (mfd *Memfd) Seals() int {
	seals, err := mfd.seals()
	if err != nil {
		return 0
	}
	return seals
}

// SetSeals sets the current seals. It can error if the item is sealed.
func (mfd *Memfd) SetSeals(seals int) error {
	return SyscallFcntlSetSeals(mfd.Fd(), seals)
}

// IsImmutable returns true if the memfd is fully immutable, all seals set
func (mfd *Memfd) IsImmutable() bool {
	seals, err := SyscallFcntlSeals(mfd.Fd())
	if err != nil {
		return false
	}
	return seals == ALL_SEALS
}

// SetImmutable fully seals the memfd if it is not already.
func (mfd *Memfd) SetImmutable() error {
	err := mfd.SetSeals(ALL_SEALS)
	if err == nil {
		return nil
	}
	// already immutable, will return EPERM
	if mfd.IsImmutable() {
		return nil
	}
	return err
}

const maxint int64 = int64(^uint(0) >> 1)

// Map returns a byte slice with the memfd contents in.
// This is a private read only mapping if the memfd has a write
// seal, and a shared writeable mapping if it is not sealed.
// The slice must be under 2GB on a 32 bit platform.
// Writeable mappings must be unmapped before sealing.
func (mfd *Memfd) Map() ([]byte, error) {
	seals, err := mfd.seals()
	if err != nil {
		return []byte{}, err
	}
	prot := syscall.PROT_READ | syscall.PROT_WRITE
	flags := syscall.MAP_SHARED
	if seals&F_SEAL_WRITE == F_SEAL_WRITE {
		prot = syscall.PROT_READ
		flags = syscall.MAP_PRIVATE
	}
	size := mfd.Size()
	if size > maxint {
		return []byte{}, ErrTooBig
	}
	b, err := syscall.Mmap(int(mfd.Fd()), 0, int(size), prot, flags)
	if err != nil {
		return []byte{}, err
	}
	mfd.b = b
	return b, nil
}

// Unmap clears a mapping. Note that Close does not Unmap, it is fine to use the mapping after close.
func (mfd *Memfd) Unmap() error {
	err := syscall.Munmap(mfd.b)
	mfd.b = []byte{}
	return err
}
