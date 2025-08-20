package core

import "fmt"

// GuardianError represents a Guardian-specific error
type GuardianError struct {
	Code    ErrorCode
	Message string
	Cause   error
}

// Error implements the error interface
func (e *GuardianError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap implements the unwrap interface for Go 1.13+ error handling
func (e *GuardianError) Unwrap() error {
	return e.Cause
}

// ErrorCode represents different types of Guardian errors
type ErrorCode string

const (
	// Platform errors
	ErrPlatformNotSupported ErrorCode = "PLATFORM_NOT_SUPPORTED"
	ErrPlatformRequirements ErrorCode = "PLATFORM_REQUIREMENTS"

	// Firewall errors
	ErrFirewallAccess    ErrorCode = "FIREWALL_ACCESS"
	ErrFirewallOperation ErrorCode = "FIREWALL_OPERATION"
	ErrIPAlreadyBlocked  ErrorCode = "IP_ALREADY_BLOCKED"
	ErrIPNotBlocked      ErrorCode = "IP_NOT_BLOCKED"
	ErrInvalidIP         ErrorCode = "INVALID_IP"

	// Storage errors
	ErrStorageConnection ErrorCode = "STORAGE_CONNECTION"
	ErrStorageOperation  ErrorCode = "STORAGE_OPERATION"
	ErrRecordNotFound    ErrorCode = "RECORD_NOT_FOUND"

	// Configuration errors
	ErrConfigInvalid    ErrorCode = "CONFIG_INVALID"
	ErrConfigNotFound   ErrorCode = "CONFIG_NOT_FOUND"
	ErrConfigPermission ErrorCode = "CONFIG_PERMISSION"

	// Log monitoring errors
	ErrLogFileNotFound   ErrorCode = "LOG_FILE_NOT_FOUND"
	ErrLogFilePermission ErrorCode = "LOG_FILE_PERMISSION"
	ErrLogParseError     ErrorCode = "LOG_PARSE_ERROR"

	// Service errors
	ErrServiceNotRunning ErrorCode = "SERVICE_NOT_RUNNING"
	ErrServicePermission ErrorCode = "SERVICE_PERMISSION"
	ErrServiceInstall    ErrorCode = "SERVICE_INSTALL"
	ErrServiceUninstall  ErrorCode = "SERVICE_UNINSTALL"
)

// NewError creates a new GuardianError
func NewError(code ErrorCode, message string, cause error) *GuardianError {
	return &GuardianError{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}

// NewErrorf creates a new GuardianError with formatted message
func NewErrorf(code ErrorCode, cause error, format string, args ...interface{}) *GuardianError {
	return &GuardianError{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
		Cause:   cause,
	}
}

// IsErrorCode checks if an error is a GuardianError with the specified code
func IsErrorCode(err error, code ErrorCode) bool {
	if guardianErr, ok := err.(*GuardianError); ok {
		return guardianErr.Code == code
	}
	return false
}
