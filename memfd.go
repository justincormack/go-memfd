// +build linux

// Package memfd provides a Go library for working with Linux memfd memory file descriptors.
// This provides shareable anonymous memory which can be locked.
package memfd

import (
	"errors"
	"os"
	"syscall"

	"github.com/justincormack/go-memfd/msyscall"
)

var (
	// ErrTooBig is returned if you try to map a memfd over 2GB on a 32 bit platform
	ErrTooBig = errors.New("memfd too large for slice")
)

const (
	// Cloexec sets the cloexec flag on the memfd when opened
	Cloexec = msyscall.MFD_CLOEXEC
	// AllowSealing allows seal operations to be performed
	AllowSealing = msyscall.MFD_ALLOW_SEALING

	// SealSeal means no more seal operations can be performed
	SealSeal = msyscall.F_SEAL_SEAL
	// SealShrink means the memfd may no longer shrink
	SealShrink = msyscall.F_SEAL_SHRINK
	// SealGrow means the memfd may no longer grow
	SealGrow = msyscall.F_SEAL_GROW
	// SealWrite means the memfd may no longer be written to
	SealWrite = msyscall.F_SEAL_WRITE
	// SealAll means the memfd is now immutable
	SealAll = SealSeal | SealShrink | SealGrow | SealWrite
)

// Memfd is the type for an memory fd, an os.File with extra methods
type Memfd struct {
	*os.File
	b []byte
}

// Create creates a memfd and sets the flags to the most common options, Cloexec and AllowSealing.
func Create() (*Memfd, error) {
	return CreateNameFlags("", Cloexec|AllowSealing)
}

// CreateNameFlags creates a memfd, and allows setting name and flags if required.
// Name is not required; it is recorded as the symlink name in /proc/self/fd.
func CreateNameFlags(name string, flags uint) (*Memfd, error) {
	fd, err := msyscall.MemfdCreate(name, flags)
	if err != nil {
		return nil, err
	}
	memfd := Memfd{os.NewFile(uintptr(fd), name), []byte{}}
	return &memfd, nil
}

// New creates a memfd object from a file descriptor, eg passed via a pipe or to an exec.
// Will return an error if the file was not a memfd, ie cannot have seals.
func New(fd uintptr) (*Memfd, error) {
	_, err := msyscall.FcntlSeals(fd)
	if err != nil {
		return nil, err
	}
	// TODO(justin) read name with readlink /proc/self/fd
	mfd := Memfd{os.NewFile(uintptr(fd), ""), []byte{}}
	return &mfd, nil
}

// Size returns the current size of the memfd.
// It could return an error if the memfd is closed; we return zero in that case.
func (mfd *Memfd) Size() int64 {
	fi, err := mfd.Stat()
	if err != nil {
		return 0
	}
	return fi.Size()
}

// SetSize sets the size of the memfd. It is just Truncate but a more understandable name.
func (mfd *Memfd) SetSize(size int64) error {
	return mfd.Truncate(size)
}

// ClearCloexec clears the (default) Cloexec flag. Useful just before you exec another process
// if you are not using the Go fork exec which can set this.
// While this can fail in theory, it only will if the file is closed, so we ignore error.
func (mfd *Memfd) ClearCloexec() {
	_ = msyscall.FcntlCloexec(mfd.Fd(), 0)
}

// SetCloexec sets the Cloexec flag. Useful if you inherited an fd with this clear.
// While this can fail in theory, it only will if the file is closed, so we ignore error.
func (mfd *Memfd) SetCloexec() {
	_ = msyscall.FcntlCloexec(mfd.Fd(), 1)
}

// seals is an internal function that returns seals or an error
func (mfd *Memfd) seals() (int, error) {
	return msyscall.FcntlSeals(mfd.Fd())
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
	return msyscall.FcntlSetSeals(mfd.Fd(), seals)
}

// IsImmutable returns true if the memfd is fully immutable, all seals set
func (mfd *Memfd) IsImmutable() bool {
	seals, err := msyscall.FcntlSeals(mfd.Fd())
	if err != nil {
		return false
	}
	return seals == SealAll
}

// SetImmutable fully seals the memfd if it is not already.
func (mfd *Memfd) SetImmutable() error {
	err := mfd.SetSeals(SealAll)
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
// Multiple requests return the same slice.
func (mfd *Memfd) Map() ([]byte, error) {
	if cap(mfd.b) > 0 {
		return mfd.b, nil
	}
	seals, err := mfd.seals()
	if err != nil {
		return []byte{}, err
	}
	var prot, flags int
	if seals&SealWrite == SealWrite {
		prot = syscall.PROT_READ
		flags = syscall.MAP_PRIVATE
	} else {
		prot = syscall.PROT_READ | syscall.PROT_WRITE
		flags = syscall.MAP_SHARED
	}
	size := mfd.Size()
	if size > maxint {
		return []byte{}, ErrTooBig
	}
	if size == 0 {
		return []byte{}, nil
	}
	b, err := syscall.Mmap(int(mfd.Fd()), 0, int(size), prot, flags)
	if err != nil {
		return []byte{}, err
	}
	mfd.b = b
	return b, nil
}

// Unmap clears a mapping. Note that Close does not Unmap, it is fine to use the mapping after close,
// so you should usually defer Close and defer Unmap to avoid leaking resources.
func (mfd *Memfd) Unmap() error {
	if cap(mfd.b) == 0 {
		return nil
	}
	err := syscall.Munmap(mfd.b)
	mfd.b = []byte{}
	return err
}

// Remap adjusts the mapping to the current size of the memfd. It may be a new slice.
// Implementation is currently inefficient, will rework to use mremap syscall if this is useful.
func (mfd *Memfd) Remap() ([]byte, error) {
	if cap(mfd.b) == 0 {
		return mfd.Map()
	}
	err := mfd.Unmap()
	if err != nil {
		return []byte{}, err
	}
	return mfd.Map()
}
