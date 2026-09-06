//go:build windows

package selfupgrade

import (
	"fmt"
	"syscall"
)

// On Windows a running executable cannot be replaced in place and unix.Exec is
// unavailable, so the in-place exec/respawn restart strategies are not
// supported. frppc on Windows should use UpgradeTypeReplaceOnly (replace, then
// restart the service separately) or UpgradeTypeCommand, which runs an external
// upgrade command and works on every platform.

// detachSysProcAttr returns the SysProcAttr that detaches a spawned process
// from the current one so the upgrade command survives frppc restarting. On
// Windows that is DETACHED_PROCESS (no inherited console) plus a new process
// group so a Ctrl-Break to frppc does not reach it.
func detachSysProcAttr() *syscall.SysProcAttr {
	const detachedProcess = 0x00000008
	const createNewProcessGroup = 0x00000200
	return &syscall.SysProcAttr{CreationFlags: detachedProcess | createNewProcessGroup}
}

func reExecSelf(target string) error {
	return fmt.Errorf("in-place re-exec upgrade is not supported on Windows")
}

func respawnDetachedAndExit(target string) {
	// Cannot restart in place on Windows; leave the replaced binary for the
	// next service restart.
}
