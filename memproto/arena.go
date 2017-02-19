// Package memproto provides a Capnproto Arena interface, which provides a
// single memfd backed data segment that can be written and read without
// copying and which can be grown to any size. It can then be shared as
// a memfd, anonymous, lockable shared memory, without being marshaled
// and unmarshaled - the transport header is not really useful for this use
// case; the file descriptor contains all the information about size.
//
// For more information about Capnproto see https://capnproto.org/encoding.html
// The primary advantage here over alternative serialization methods is that
// there is no parsing stage, unlike for example protobuf, so the process that
// receives the memfd does not have to allocate any memory, and can read the
// data directly out of the passed shared memory segment. The code to read
// the data is therefore also small and easy to read and audit. This makes
// it a good choice for applications that need to be secure, as well as
// those that need to send or receive large quantities of data, as the data
// is not copied, just mapped, unlike if you send it via a socket.
// Note that the Capnproto RPC protocol is not implemented at present,
// although it could be implemented using streams of file descriptors over
// Unix sockets.
package memproto

import (
	"errors"
	"os"

	"github.com/justincormack/go-memfd"
	"zombiezen.com/go/capnproto2"
)

// MemfdArena is the base type, that wraps a memfd
type MemfdArena struct {
	*memfd.Memfd
}

// Create creates an memfd arena
func Create() (*MemfdArena, error) {
	mfd, err := memfd.Create()
	if err != nil {
		return nil, err
	}
	mfa := MemfdArena{mfd}
	return &mfa, nil
}

// CreateNameFlags creates an memfd arena specifying name and flags is needed
func CreateNameFlags(name string, flags uint) (*MemfdArena, error) {
	mfd, err := memfd.CreateNameFlags(name, flags)
	if err != nil {
		return nil, err
	}
	mfa := MemfdArena{mfd}
	return &mfa, nil
}

// New creates an memfd arena from a file descriptor
func New(fd uintptr) (*MemfdArena, error) {
	mfd, err := memfd.New(fd)
	if err != nil {
		return nil, err
	}
	mfa := MemfdArena{mfd}
	return &mfa, nil
}

// NumSegments returns the number of segments, which is zero if empty, then 1
func (mfa *MemfdArena) NumSegments() int64 {
	if mfa.Size() == 0 {
		return 0
	}
	return 1
}

// errSegmentOutOfBounds is returned if an invalid segment is used
var errSegmentOutOfBounds = errors.New("segment ID out of bounds")

// Data returns the byte slice of data from a segment.
// This is a Map of the underlying memfd.
func (mfa *MemfdArena) Data(id capnp.SegmentID) ([]byte, error) {
	if mfa.NumSegments() == 0 {
		return nil, errSegmentOutOfBounds
	}
	if id != 0 {
		return nil, errSegmentOutOfBounds
	}
	return mfa.Map()
}

// buffer must grow by at least a page; more might be desirable.
var minBufferGrowth = capnp.Size(os.Getpagesize())

// Allocate expands the segment so it has space for at least sz more data, by calling Remap on the memfd.
func (mfa *MemfdArena) Allocate(sz capnp.Size, segs map[capnp.SegmentID]*capnp.Segment) (capnp.SegmentID, []byte, error) {
	var data []byte
	var err error
	if segs[0] != nil {
		data = segs[0].Data()
	} else {
		data, err = mfa.Map()
		if err != nil {
			return 0, []byte{}, err
		}
	}
	if hasCapacity(data, sz) {
		return 0, data, nil
	}
	if sz%minBufferGrowth != 0 {
		sz = sz + minBufferGrowth - (sz % minBufferGrowth)
	}
	currentSize := len(data)
	mfa.SetSize(int64(cap(data)) + int64(sz))
	r, err := mfa.Remap()
	if err != nil {
		return 0, []byte{}, err
	}
	return 0, r[:currentSize], nil
}

func hasCapacity(b []byte, sz capnp.Size) bool {
	return sz <= capnp.Size(cap(b)-len(b))
}

// Message returns a capnproto Message from an (non empty) arena.
func (mfa *MemfdArena) Message() *capnp.Message {
	return &capnp.Message{Arena: mfa}
}
