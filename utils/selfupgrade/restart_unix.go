//go:build !windows

package selfupgrade

import (
	"os"
	"syscall"

	"golang.org/x/sys/unix"
)

// detachSysProcAttr returns the SysProcAttr that detaches a spawned process
// from the current one so it survives this process being replaced or exiting.
// On unix that means starting a new session (setsid).
func detachSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{Setsid: true}
}

// reExecSelf replaces the current process image with the freshly installed
// binary at target, preserving PID, argv and environment. On success it does
// not return.
func reExecSelf(target string) error {
	argv := append([]string{target}, os.Args[1:]...)
	return unix.Exec(target, argv, os.Environ())
}

// respawnDetachedAndExit starts the new binary as a detached session leader,
// then exits the current process so a supervisor or the new process takes over.
func respawnDetachedAndExit(target string) {
	argv := append([]string{target}, os.Args[1:]...)
	// Respawn in the current working directory. The process relies on relative
	// paths (e.g. storage/...) resolved against the cwd it was launched with,
	// not against the directory the binary happens to live in. An empty Dir
	// would already inherit the parent's cwd; set it explicitly for clarity and
	// fall back to inheritance if Getwd fails.
	cwd, _ := os.Getwd()
	attr := &os.ProcAttr{
		Dir:   cwd,
		Env:   os.Environ(),
		Files: []*os.File{os.Stdin, os.Stdout, os.Stderr},
		Sys:   detachSysProcAttr(),
	}
	if p, err := os.StartProcess(target, argv, attr); err == nil {
		_ = p.Release()
	}
	os.Exit(0)
}
