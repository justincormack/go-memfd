## Go Memfd library

[![GoDoc](https://godoc.org/github.com/justincormack/go-memfd?status.svg)](https://godoc.org/github.com/justincormack/go-memfd)
[![Build Status](https://travis-ci.org/justincormack/go-memfd.svg?branch=master)](https://travis-ci.org/justincormack/go-memfd)

This is a Go library for working with Linux memfd, memory file descriptors.

These provide shareable anonymous memory, which can be passed around via file descriptors,
and also locked from write or resize. They are designed to let programs that do not trust each
other communicate via shared memory without issues of naming, truncation, or race conditions due
to modifications.

For more information about the underlying syscalls see [`memfd_create`](http://man7.org/linux/man-pages/man2/memfd_create.2.html)
and the file sealing section of [`fcntl`](http://man7.org/linux/man-pages/man2/fcntl.2.html). There is also a
[sealed files overview](https://lwn.net/Articles/593918/) from LWN, but written slightly before the final design was merged,
so the details are not quite correct for the final version.

The functionality was added in Linux 3.17, in February 2015. It was also added to the Debian Jessie 3.16 kernels, and is in the
Ubuntu 14.04 updates, as well as being backported to the Centos 7.3/RHEL 7.3 series, so it is available in all non ancient Linux
distros, so should be generally usable.. Currently there is no support in the BSDs or other non Linux systems, I hope this can
be added as it is a useful interface.

A Capnproto Arena library is also included, for sending structured data between processes.
