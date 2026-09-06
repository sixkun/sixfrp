package selfupgrade

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// UpgradeType selects how a downloaded release is applied. It mirrors
// pb.FrppUpgradeType so callers can convert the wire enum directly.
type UpgradeType int32

const (
	// UpgradeTypeReplaceOnly downloads and replaces the on-disk binary but does not
	// restart the process. The new code takes effect on the next restart. This
	// is the default / backward-compatible behavior.
	UpgradeTypeReplaceOnly UpgradeType = 0
	// UpgradeTypeReplaceExec replaces the binary then re-execs the current process in
	// place (same PID) via unix.Exec. Not supported on Windows.
	UpgradeTypeReplaceExec UpgradeType = 1
	// UpgradeTypeReplaceRespawn replaces the binary, spawns a new detached process,
	// then exits the current one. Not supported on Windows.
	UpgradeTypeReplaceRespawn UpgradeType = 2
	// UpgradeTypeCommand runs an external upgrade command with {{url}} substituted,
	// detached, frpps-style. The command owns download and restart.
	UpgradeTypeCommand UpgradeType = 3
)

// Options describes a single upgrade request.
type Options struct {
	// DownloadURL is the release .tar.gz to fetch. Required for every type
	// except UpgradeTypeCommand, where it is substituted into Command's {{url}}.
	DownloadURL string
	// TargetPath is the binary to overwrite; defaults to the running executable
	// (symlinks resolved). Unused by UpgradeTypeCommand.
	TargetPath string
	// Backup keeps a .bak copy of the previous binary before overwrite.
	Backup bool
	// Type selects the restart strategy (see the Type* constants).
	Type UpgradeType
	// Command is the external upgrade command template for UpgradeTypeCommand. It must
	// contain the {{url}} placeholder.
	Command string
}

// restartDelay gives the caller's RPC response time to flush before the
// process re-execs or exits.
const restartDelay = 2 * time.Second

// Upgrade applies a release according to opt.Type.
//
//   - UpgradeTypeReplaceOnly: download + replace, return (restart handled elsewhere).
//   - UpgradeTypeReplaceExec: download + replace, then re-exec self after a short delay.
//   - UpgradeTypeReplaceRespawn: download + replace, then spawn detached + exit self.
//   - UpgradeTypeCommand: launch the configured command detached with {{url}} filled in.
//
// For the replace types the download/replace happens synchronously so failures
// surface to the caller; only the restart is deferred.
func Upgrade(opt Options) error {
	if opt.Type == UpgradeTypeCommand {
		return runUpgradeCommand(opt.Command, opt.DownloadURL)
	}

	target, err := replaceBinary(opt.DownloadURL, opt.TargetPath, opt.Backup)
	if err != nil {
		return err
	}

	switch opt.Type {
	case UpgradeTypeReplaceOnly:
		return nil
	case UpgradeTypeReplaceExec:
		scheduleRestart(func() { _ = reExecSelf(target) })
		return nil
	case UpgradeTypeReplaceRespawn:
		scheduleRestart(func() { respawnDetachedAndExit(target) })
		return nil
	default:
		return fmt.Errorf("unsupported upgrade type %d", opt.Type)
	}
}

// scheduleRestart runs fn after restartDelay so the triggering RPC can respond
// before the process is replaced.
func scheduleRestart(fn func()) {
	go func() {
		time.Sleep(restartDelay)
		fn()
	}()
}

// runUpgradeCommand launches command detached with the {{url}} placeholder
// replaced by downloadURL. The command owns download + restart, so this returns
// as soon as the process is started. Detaching is platform-specific
// (detachSysProcAttr): setsid on unix, DETACHED_PROCESS on Windows.
func runUpgradeCommand(command, downloadURL string) error {
	command = strings.TrimSpace(command)
	if command == "" {
		return fmt.Errorf("frppc upgrade command is not configured (FRPPC_UPGRADE_COMMAND)")
	}
	if !strings.Contains(command, "{{url}}") {
		return fmt.Errorf("frppc upgrade command must contain the {{url}} placeholder")
	}
	if strings.TrimSpace(downloadURL) == "" {
		return fmt.Errorf("download_url is required")
	}
	return RunDetachedCommand(strings.ReplaceAll(command, "{{url}}", downloadURL))
}

// RunDetachedCommand launches command as a detached process (setsid on unix,
// DETACHED_PROCESS on Windows) and returns as soon as it has started. It is
// used for the restart command, which is expected to restart the agent's own
// service out-of-band. Unlike the upgrade command it takes no {{url}}
// placeholder.
func RunDetachedCommand(command string) error {
	fields := strings.Fields(strings.TrimSpace(command))
	if len(fields) == 0 {
		return fmt.Errorf("empty command")
	}
	attr := &os.ProcAttr{
		Env:   os.Environ(),
		Files: []*os.File{os.Stdin, os.Stdout, os.Stderr},
		Sys:   detachSysProcAttr(),
	}
	p, err := os.StartProcess(fields[0], fields, attr)
	if err != nil {
		return fmt.Errorf("launch command: %w", err)
	}
	return p.Release()
}

// replaceBinary downloads the single executable contained in a release .tar.gz
// and atomically replaces targetPath. Unix permits replacing an in-use
// executable. It returns the resolved target path.
func replaceBinary(downloadURL, targetPath string, backup bool) (string, error) {
	if strings.TrimSpace(downloadURL) == "" {
		return "", fmt.Errorf("download_url is required")
	}
	if targetPath == "" {
		var err error
		targetPath, err = os.Executable()
		if err != nil {
			return "", fmt.Errorf("get executable path: %w", err)
		}
	}
	targetPath, err := filepath.EvalSymlinks(targetPath)
	if err != nil {
		return "", fmt.Errorf("resolve executable path: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, nil)
	if err != nil {
		return "", err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("download release: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("download release returned %s", resp.Status)
	}

	gz, err := gzip.NewReader(resp.Body)
	if err != nil {
		return "", fmt.Errorf("open release archive: %w", err)
	}
	defer gz.Close()

	tmp, err := os.CreateTemp(filepath.Dir(targetPath), ".frpp-upgrade-*")
	if err != nil {
		return "", fmt.Errorf("create upgrade file: %w", err)
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	found := false
	tr := tar.NewReader(gz)
	for {
		h, nextErr := tr.Next()
		if nextErr == io.EOF {
			break
		}
		if nextErr != nil {
			tmp.Close()
			return "", fmt.Errorf("read release archive: %w", nextErr)
		}
		if h.Typeflag != tar.TypeReg || strings.HasPrefix(filepath.Base(h.Name), ".") {
			continue
		}
		if found {
			tmp.Close()
			return "", fmt.Errorf("release archive contains multiple files")
		}
		if _, err := io.Copy(tmp, tr); err != nil {
			tmp.Close()
			return "", fmt.Errorf("extract release binary: %w", err)
		}
		found = true
	}
	if err := tmp.Close(); err != nil {
		return "", err
	}
	if !found {
		return "", fmt.Errorf("release archive contains no binary")
	}
	if err := os.Chmod(tmpPath, 0o755); err != nil {
		return "", err
	}

	backupPath := targetPath + ".bak"
	if backup {
		_ = os.Remove(backupPath)
		if err := copyFile(targetPath, backupPath); err != nil {
			return "", fmt.Errorf("backup executable: %w", err)
		}
	}
	if err := os.Rename(tmpPath, targetPath); err != nil {
		return "", fmt.Errorf("replace executable: %w", err)
	}
	return targetPath, nil
}

func copyFile(srcPath, dstPath string) error {
	src, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer src.Close()
	dst, err := os.OpenFile(dstPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o755)
	if err != nil {
		return err
	}
	if _, err := io.Copy(dst, src); err != nil {
		dst.Close()
		return err
	}
	return dst.Close()
}
