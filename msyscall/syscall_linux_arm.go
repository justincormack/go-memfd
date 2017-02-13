package msyscall

import "syscall"

const (
	// SYS_FCNTL is the fcntl syscall to use for this architecture
	SYS_FCNTL = syscall.SYS_FCNTL64
	// SYS_MEMFD_CREATE is the syscall number for this architecture
	SYS_MEMFD_CREATE = 385
)
