package utils

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"sync"
	"time"

	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/log"
	"github.com/kardianos/service"
)

type SystemService struct {
	run     func()
	app     contractsfoundation.Application // For graceful shutdown
	mu      sync.Mutex
	done    chan struct{}
	timeout time.Duration
	service.Service
}

func (ss *SystemService) Start(s service.Service) error {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	if ss.done != nil {
		return nil // already started
	}

	done := make(chan struct{})
	ss.done = done

	go func() {
		defer close(done)
		ss.iRun()
	}()

	return nil
}

func (ss *SystemService) Stop(s service.Service) error {
	ss.mu.Lock()
	app := ss.app
	done := ss.done
	timeout := ss.timeout
	ss.mu.Unlock()

	// Actively call Shutdown if Application is available
	var shutdownErr error
	if app != nil {
		shutdownErr = app.Shutdown()
	}

	// Wait for graceful shutdown with timeout
	if done != nil {
		if timeout == 0 {
			timeout = 15 * time.Second // default timeout
		}

		select {
		case <-done:
			// Clean shutdown
		case <-time.After(timeout):
			return fmt.Errorf("timeout waiting for service to stop after %v", timeout)
		}
	}

	return shutdownErr
}

func (ss *SystemService) iRun() {
	defer func() {
		if service.Interactive() {
			ss.Stop(ss.Service)
		} else {
			ss.Service.Stop()
		}
	}()
	ss.run()
}

func CreateSystemServiceWithApp(svcName, displayName, description string, app contractsfoundation.Application, workDir string, options service.KeyValue) (service.Service, error) {
	currentPath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("get current path failed, err: %v", err)
	}

	// Use provided workDir or default to executable directory
	if workDir == "" {
		workDir = path.Dir(currentPath)
	}

	svcConfig := &service.Config{
		Name:             svcName,
		DisplayName:      displayName,
		Description:      description,
		Arguments:        nil, // No arguments needed when using Application
		WorkingDirectory: workDir,
		Option:           options,
	}

	ss := &SystemService{
		run: func() {
			if app != nil {
				app.Start()
			}
		},
		app:     app, // Store app for graceful shutdown
		timeout: 15 * time.Second,
	}

	s, err := service.New(ss, svcConfig)
	if err != nil {
		return nil, fmt.Errorf("service New failed, err: %v", err)
	}
	ss.Service = s
	return s, nil
}

func CreateSystemServiceWithOptions(svcName, displayName, description string, args []string, run func(), workDir string, options service.KeyValue) (service.Service, error) {
	currentPath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("get current path failed, err: %v", err)
	}

	// Use provided workDir or default to executable directory
	if workDir == "" {
		workDir = path.Dir(currentPath)
	}

	svcConfig := &service.Config{
		Name:             svcName,
		DisplayName:      displayName,
		Description:      description,
		Arguments:        args,
		WorkingDirectory: workDir,
		Option:           options,
	}

	ss := &SystemService{
		run:     run,
		timeout: 15 * time.Second,
	}

	s, err := service.New(ss, svcConfig)
	if err != nil {
		return nil, fmt.Errorf("service New failed, err: %v", err)
	}
	ss.Service = s
	return s, nil
}

func ControlSystemService(svcName, displayName, description string, args []string, action string, run func(), workDir string, logger log.Log) error {
	return ControlSystemServiceWithOptions(svcName, displayName, description, args, action, run, workDir, nil, logger)
}

// ControlSystemServiceSimple controls an already-installed service without needing DisplayName/Description.
// Use this for start, stop, restart, status, and uninstall commands where the service is already configured.
func ControlSystemServiceSimple(svcName string, action string, logger log.Log) error {
	// For control operations on existing services, DisplayName and Description don't matter
	// The service is already installed with its configuration
	return ControlSystemServiceWithOptions(svcName, "", "", nil, action, func() {}, "", nil, logger)
}

func ControlSystemServiceWithOptions(svcName, displayName, description string, args []string, action string, run func(), workDir string, options service.KeyValue, logger log.Log) error {
	ctx := context.Background()

	if logger != nil {
		logger.Infof("try to %s service, args: %v, workDir: %s", action, args, workDir)
	}

	s, err := CreateSystemServiceWithOptions(svcName, displayName, description, args, run, workDir, options)
	if err != nil {
		if logger != nil {
			logger.WithContext(ctx).Error("create service controller failed: " + err.Error())
		}
		return err
	}

	if err := service.Control(s, action); err != nil {
		if logger != nil {
			logger.WithContext(ctx).Errorf("controller %v service failed: %v", action, err)
		}
		return err
	}

	if logger != nil {
		logger.Infof("controller %v service success", action)
	}
	return nil
}

// InstallToSystemPath copies the current executable to a system path (e.g., /usr/local/bin).
// NOTE: Currently unused. The service installation uses the original binary location,
// which is simpler and avoids permission issues. kardianos/service records the full
// path to the current executable, so copying to a system path is not necessary.
// Users can manually create symlinks if global access is needed:
//
//	ln -s /path/to/frpps /usr/local/bin/frpps
func InstallToSystemPath(installPath string) error {
	currentPath, err := os.Executable()
	if err != nil {
		return err
	}

	targetPath := path.Join(installPath, filepath.Base(currentPath))

	src, err := os.Open(currentPath)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(targetPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	if err != nil {
		return err
	}

	err = os.Chmod(targetPath, 0755)
	if err != nil {
		return err
	}

	return nil
}
