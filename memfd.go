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
	// ErrTooBig is returned if you try to map a memfd over 2GB, due to current Go limitations
	ErrTooBig = errors.New("memfd too large for slice")
	// ErrAlreadyMapped is returned if you try to map a writeable memfd more than once
	ErrAlreadyMapped = errors.New("memfd already mapped")
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
	memfd := Memfd{os.NewFile(uintptr(fd), ""), []byte{}}
	return &memfd, nil
}

// Seals returns the current seals. It can only error if something
// out of the ordinary has happened, eg the file is closed, so we
// just return 0 (no seals) in this case.
func (memfd *Memfd) Seals() int {
	seals, err := SyscallFcntlSeals(memfd.Fd())
	if err != nil {
		return 0
	}
	return seals
}

// IsImmutable returns true if the memfd is fully immutable, all seals set
func (memfd *Memfd) IsImmutable() bool {
	seals, err := SyscallFcntlSeals(memfd.Fd())
	if err != nil {
		return false
	}
	return seals == ALL_SEALS
}

// SetSeals sets the current seals. It can error if the item is sealed.
func (memfd *Memfd) SetSeals(seals int) error {
	return SyscallFcntlSetSeals(memfd.Fd(), seals)
}

// SetImmutable fully seals the memfd if it is not already.
func (memfd *Memfd) SetImmutable() error {
	err := memfd.SetSeals(ALL_SEALS)
	if err == nil {
		return nil
	}
	// already immutable, will return EPERM
	if memfd.IsImmutable() {
		return nil
	}
	return err
}

// Size returns the current size of the memfd
// It could return an error if the memfd is closed, so
// we return zero in that case.
func (memfd *Memfd) Size() int64 {
	fi, err := memfd.Stat()
	if err != nil {
		return 0
	}
	return fi.Size()
}

// SetSize sets the size of the memfd. It is just
// Truncate but a more understandable name.
func (memfd *Memfd) SetSize(size int64) error {
	return memfd.Truncate(size)
}

const maxint int64 = int64(^uint(0) >> 1)

// MapImmutable makes sure the memfd is immutable and returns a byte slice
// with the contents in. You will not be able to modify this.
// In general this should not normally fail if the memfd is already immutable.
// Go does not allow slices over 2GB at present.
// Multiple immutable maps return the same slice.
func (memfd *Memfd) MapImmutable() ([]byte, error) {
	if cap(memfd.b) != 0 {
		return memfd.b, nil
	}
	err := memfd.SetImmutable()
	if err != nil {
		return []byte{}, err
	}
	size := memfd.Size()
	if size > maxint {
		return []byte{}, ErrTooBig
	}
	memfd.b, err = syscall.Mmap(int(memfd.Fd()), 0, int(size), syscall.PROT_READ, syscall.MAP_PRIVATE)
	return memfd.b, err
}

// MapRW returns a read write byte slice from mmaping the underlying memfd.
// You need to specify a capacity; currently as there is no remap support you
// should specify something larger than what you might need and truncate later
// A zero size uses the current size.
// Go does not allow slices over 2GB at present.
// Sealing will fail while a writeable mapping exists, it must be unmapped first.
// Mapping multiple times returns an error
func (memfd *Memfd) MapRW() ([]byte, error) {
	if cap(memfd.b) != 0 {
		return []byte{}, ErrAlreadyMapped
	}
	size := memfd.Size()
	if size > maxint {
		return []byte{}, ErrTooBig
	}
	var err error
	memfd.b, err = syscall.Mmap(int(memfd.Fd()), 0, int(size), syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	return memfd.b, err
}

// Unmap clears a mapping. Note that Close does not Unmap, it is fine to use the mapping after close.
// Writeable maps must be unmapped before sealing.
func (memfd *Memfd) Unmap() error {
	return syscall.Munmap(memfd.b)
}
