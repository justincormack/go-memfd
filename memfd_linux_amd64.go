package memfd

import "syscall"

const (
	SYS_FCNTL        = syscall.SYS_FCNTL
	SYS_MEMFD_CREATE = 319
)
