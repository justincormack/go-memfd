## Capnproto Arena based on memfd

[![GoDoc](https://godoc.org/github.com/justincormack/go-memfd/memproto?status.svg)](https://godoc.org/github.com/justincormack/go-memfd/memproto)

This provides a single memfd backed data segment that can be written and read
without copying and which can be grown to any size. It can then be shared as
a memfd, anonymous, lockable shared memory, without being marshaled
and unmarshaled - the transport header is not really useful for this use
case; the file descriptor contains all the information about size.
    
The primary advantage here over alternative serialization methods is that
there is no parsing stage, unlike for example protobuf, so the process that
receives the memfd does not have to allocate any memory, and can read the
data directly out of the passed shared memory segment. The code to read
the data is therefore also small and easy to read and audit. This makes
it a good choice for applications that need to be secure, as well as
those that need to send or receive large quantities of data, as the data
is not copied, just mapped, unlike if you send it via a socket.
Note that the Capnproto RPC protocol is not implemented at present,
although it could be implemented using streams of file descriptors over
Unix sockets.
