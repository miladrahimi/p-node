// Package paqet runs the paqet binary (https://github.com/hanselime/paqet), a
// raw-packet KCP tunnel, as a supervised child process in client or server role.
package paqet

import (
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/miladrahimi/p-node/pkg/logger"
	"github.com/miladrahimi/p-node/pkg/util"
	"go.uber.org/zap"
)

// retryDelay is the pause before an unexpectedly exited process is restarted.
const retryDelay = 5 * time.Second

// Paqet supervises one paqet process. Reconfigure starts, restarts, or stops it
// so that it always runs the last config given (or nothing).
type Paqet struct {
	l          *logger.Logger
	binaryPath string
	configPath string
	logPath    string
	mu         sync.Mutex
	config     *Config
	cmd        *exec.Cmd
	exited     chan struct{}
	killed     *atomic.Bool
}

// New creates a paqet supervisor. Nothing runs until Reconfigure is called.
func New(l *logger.Logger, binaryPath, configPath, logPath string) *Paqet {
	return &Paqet{l: l, binaryPath: binaryPath, configPath: configPath, logPath: logPath}
}

// Reconfigure applies the config: nil stops the process, an unchanged config is
// a no-op, anything else (re)starts the process with the new config.
func (p *Paqet) Reconfigure(config *Config) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if config == nil {
		return errors.WithStack(p.stopLocked())
	}
	if err := config.Validate(); err != nil {
		return errors.WithStack(err)
	}
	if p.config != nil && *p.config == *config && p.cmd != nil {
		return nil
	}

	if err := p.stopLocked(); err != nil {
		return errors.WithStack(err)
	}
	p.config = config
	return errors.WithStack(p.runLocked())
}

// Config returns the current config, or nil when stopped.
func (p *Paqet) Config() *Config {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.config
}

// Stop stops the process.
func (p *Paqet) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	return errors.WithStack(p.stopLocked())
}

func (p *Paqet) runLocked() error {
	if !util.FileExist(p.binaryPath) {
		return errors.Errorf("paqet: binary not found at %s", p.binaryPath)
	}
	if err := os.MkdirAll(filepath.Dir(p.configPath), 0o755); err != nil {
		return errors.WithStack(err)
	}
	// The file holds the shared key, so keep it private.
	if err := os.WriteFile(p.configPath, []byte(p.config.Yaml()), 0o600); err != nil {
		return errors.WithStack(err)
	}
	if err := os.MkdirAll(filepath.Dir(p.logPath), 0o755); err != nil {
		return errors.WithStack(err)
	}
	logFile, err := os.OpenFile(p.logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return errors.WithStack(err)
	}

	if p.config.Port > 0 {
		allowPort(p.l, p.config.Port)
	}

	cmd := exec.Command(p.binaryPath, "run", "-c", p.configPath)
	cmd.Stdout, cmd.Stderr = logFile, logFile
	if err := cmd.Start(); err != nil {
		_ = logFile.Close()
		return errors.WithStack(err)
	}
	p.l.Info("paqet: started", zap.String("config", p.configPath), zap.Int("pid", cmd.Process.Pid))

	exited := make(chan struct{})
	killed := &atomic.Bool{}
	p.cmd, p.exited, p.killed = cmd, exited, killed

	go func() {
		err := cmd.Wait()
		_ = logFile.Close()
		close(exited)
		if killed.Load() {
			return
		}
		p.l.Error("paqet: process exited unexpectedly, restarting", zap.Error(err), zap.String("log", p.logPath))
		time.Sleep(retryDelay)

		p.mu.Lock()
		defer p.mu.Unlock()
		// Stopped or replaced in the meantime.
		if p.cmd != cmd || p.config == nil {
			return
		}
		if err := p.runLocked(); err != nil {
			p.l.Error("paqet: restart failed", zap.Error(err))
			p.cmd = nil
		}
	}()

	return nil
}

func (p *Paqet) stopLocked() error {
	if p.cmd != nil && p.cmd.Process != nil {
		p.killed.Store(true)
		if err := p.cmd.Process.Kill(); err != nil && !errors.Is(err, os.ErrProcessDone) {
			return errors.WithStack(err)
		}
		select {
		case <-p.exited:
		case <-time.After(10 * time.Second):
			p.l.Error("paqet: the process did not exit after kill")
		}
		p.l.Info("paqet: stopped")
	}
	if p.config != nil && p.config.Port > 0 {
		disallowPort(p.l, p.config.Port)
	}
	p.cmd = nil
	p.config = nil
	return nil
}
