package daemon

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
)

// PIDManager handles process ID management for daemon mode
type PIDManager struct {
	pidFile string
}

// NewPIDManager creates a new PID manager
func NewPIDManager() *PIDManager {
	var pidDir string

	switch runtime.GOOS {
	case "windows":
		pidDir = os.Getenv("TEMP")
		if pidDir == "" {
			pidDir = "C:\\Windows\\Temp"
		}
	default:
		pidDir = "/var/run"
		// Fallback to /tmp if /var/run is not writable
		if _, err := os.Stat(pidDir); os.IsNotExist(err) || !isWritable(pidDir) {
			pidDir = "/tmp"
		}
	}

	pidFile := filepath.Join(pidDir, "guardian.pid")

	return &PIDManager{
		pidFile: pidFile,
	}
}

// WritePID writes the current process PID to the PID file
func (pm *PIDManager) WritePID() error {
	pid := os.Getpid()
	return os.WriteFile(pm.pidFile, []byte(strconv.Itoa(pid)), 0644)
}

// ReadPID reads the PID from the PID file
func (pm *PIDManager) ReadPID() (int, error) {
	data, err := os.ReadFile(pm.pidFile)
	if err != nil {
		return 0, err
	}

	return strconv.Atoi(string(data))
}

// IsRunning checks if the process with the stored PID is running
func (pm *PIDManager) IsRunning() bool {
	pid, err := pm.ReadPID()
	if err != nil {
		return false
	}

	return pm.isProcessRunning(pid)
}

// GetRunningPID returns the PID if a Guardian daemon is running
func (pm *PIDManager) GetRunningPID() (int, bool) {
	pid, err := pm.ReadPID()
	if err != nil {
		return 0, false
	}

	if pm.isProcessRunning(pid) {
		return pid, true
	}

	// Clean up stale PID file
	pm.RemovePID()
	return 0, false
}

// RemovePID removes the PID file
func (pm *PIDManager) RemovePID() error {
	return os.Remove(pm.pidFile)
}

// GetPIDFilePath returns the path to the PID file
func (pm *PIDManager) GetPIDFilePath() string {
	return pm.pidFile
}

// isProcessRunning checks if a process with the given PID is running
func (pm *PIDManager) isProcessRunning(pid int) bool {
	if runtime.GOOS == "windows" {
		return pm.isProcessRunningWindows(pid)
	}
	return pm.isProcessRunningUnix(pid)
}

// isProcessRunningWindows checks if a process is running on Windows
func (pm *PIDManager) isProcessRunningWindows(pid int) bool {
	// On Windows, use tasklist to check if process exists
	// This is more reliable than os.FindProcess which can find zombie processes
	cmd := exec.Command("tasklist", "/FI", fmt.Sprintf("PID eq %d", pid), "/FO", "CSV", "/NH")
	output, err := cmd.Output()
	if err != nil {
		// If tasklist fails, fallback to simple process check
		_, err := os.FindProcess(pid)
		if err != nil {
			return false
		}
		// On Windows, if FindProcess succeeds, the process exists
		// We can't reliably use Signal(0) on Windows, so we assume it's running
		return true
	}

	// If output is empty or doesn't contain the PID, process is not running
	outputStr := strings.TrimSpace(string(output))
	if outputStr == "" {
		return false
	}

	// Check if the output contains our PID (more robust check)
	return strings.Contains(outputStr, fmt.Sprintf(`"%d"`, pid)) ||
		strings.Contains(outputStr, fmt.Sprintf(",%d,", pid))
}

// isProcessRunningUnix checks if a process is running on Unix-like systems
func (pm *PIDManager) isProcessRunningUnix(pid int) bool {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	// Send signal 0 to check if process exists
	err = proc.Signal(syscall.Signal(0))
	return err == nil
}

// isWritable checks if a directory is writable
func isWritable(dir string) bool {
	tempFile := filepath.Join(dir, ".guardian_test")
	err := os.WriteFile(tempFile, []byte("test"), 0644)
	if err != nil {
		return false
	}
	os.Remove(tempFile)
	return true
}

// Status represents the daemon status information
type Status struct {
	Running   bool
	PID       int
	PIDFile   string
	Startable bool
	Error     error
}

// GetStatus returns the current daemon status
func (pm *PIDManager) GetStatus() *Status {
	status := &Status{
		PIDFile:   pm.pidFile,
		Startable: true,
	}

	pid, running := pm.GetRunningPID()
	status.Running = running
	status.PID = pid

	if !running && pid != 0 {
		status.Error = fmt.Errorf("stale PID file found (process %d not running)", pid)
	}

	return status
}
