package memfd

import "syscall"

const (
	SYS_FCNTL        = syscall.SYS_FCNTL64
	SYS_MEMFD_CREATE = 356
)
