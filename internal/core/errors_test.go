package core

import (
	"errors"
	"testing"
)

func TestGuardianError(t *testing.T) {
	tests := []struct {
		name        string
		code        ErrorCode
		message     string
		cause       error
		expectedMsg string
	}{
		{
			name:        "error without cause",
			code:        ErrPlatformNotSupported,
			message:     "platform not supported",
			cause:       nil,
			expectedMsg: "[PLATFORM_NOT_SUPPORTED] platform not supported",
		},
		{
			name:        "error with cause",
			code:        ErrFirewallOperation,
			message:     "failed to block IP",
			cause:       errors.New("iptables command failed"),
			expectedMsg: "[FIREWALL_OPERATION] failed to block IP: iptables command failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewError(tt.code, tt.message, tt.cause)

			if err.Code != tt.code {
				t.Errorf("expected code %s, got %s", tt.code, err.Code)
			}

			if err.Message != tt.message {
				t.Errorf("expected message %s, got %s", tt.message, err.Message)
			}

			if err.Cause != tt.cause {
				t.Errorf("expected cause %v, got %v", tt.cause, err.Cause)
			}

			if err.Error() != tt.expectedMsg {
				t.Errorf("expected error message %s, got %s", tt.expectedMsg, err.Error())
			}
		})
	}
}

func TestNewErrorf(t *testing.T) {
	cause := errors.New("underlying error")
	err := NewErrorf(ErrConfigInvalid, cause, "invalid configuration at line %d", 42)

	expected := "[CONFIG_INVALID] invalid configuration at line 42: underlying error"
	if err.Error() != expected {
		t.Errorf("expected %s, got %s", expected, err.Error())
	}
}

func TestErrorUnwrap(t *testing.T) {
	cause := errors.New("original error")
	err := NewError(ErrStorageConnection, "database connection failed", cause)

	unwrapped := err.Unwrap()
	if unwrapped != cause {
		t.Errorf("expected unwrapped error to be %v, got %v", cause, unwrapped)
	}

	// Test with no cause
	errNoCause := NewError(ErrConfigNotFound, "config file not found", nil)
	if errNoCause.Unwrap() != nil {
		t.Error("expected unwrapped error to be nil when no cause is present")
	}
}

func TestIsErrorCode(t *testing.T) {
	guardianErr := NewError(ErrPlatformNotSupported, "test error", nil)
	regularErr := errors.New("regular error")

	// Test with GuardianError
	if !IsErrorCode(guardianErr, ErrPlatformNotSupported) {
		t.Error("expected IsErrorCode to return true for matching error code")
	}

	if IsErrorCode(guardianErr, ErrFirewallOperation) {
		t.Error("expected IsErrorCode to return false for non-matching error code")
	}

	// Test with regular error
	if IsErrorCode(regularErr, ErrPlatformNotSupported) {
		t.Error("expected IsErrorCode to return false for non-GuardianError")
	}

	// Test with nil error
	if IsErrorCode(nil, ErrPlatformNotSupported) {
		t.Error("expected IsErrorCode to return false for nil error")
	}
}

func TestErrorCodes(t *testing.T) {
	// Test that all error codes are defined and have values
	errorCodes := []ErrorCode{
		ErrPlatformNotSupported,
		ErrPlatformRequirements,
		ErrFirewallAccess,
		ErrFirewallOperation,
		ErrIPAlreadyBlocked,
		ErrIPNotBlocked,
		ErrInvalidIP,
		ErrStorageConnection,
		ErrStorageOperation,
		ErrRecordNotFound,
		ErrConfigInvalid,
		ErrConfigNotFound,
		ErrConfigPermission,
		ErrLogFileNotFound,
		ErrLogFilePermission,
		ErrLogParseError,
		ErrServiceNotRunning,
		ErrServicePermission,
		ErrServiceInstall,
		ErrServiceUninstall,
	}

	for _, code := range errorCodes {
		if string(code) == "" {
			t.Errorf("error code should not be empty: %v", code)
		}

		// Test that error code can be used to create an error
		err := NewError(code, "test message", nil)
		if err.Code != code {
			t.Errorf("expected code %s, got %s", code, err.Code)
		}
	}
}

func TestErrorWrapping(t *testing.T) {
	// Test Go 1.13+ error wrapping compatibility
	originalErr := errors.New("original error")
	guardianErr := NewError(ErrStorageOperation, "operation failed", originalErr)

	// Test errors.Is
	if !errors.Is(guardianErr, originalErr) {
		t.Error("GuardianError should wrap the original error properly")
	}

	// Test errors.As
	var target *GuardianError
	if !errors.As(guardianErr, &target) {
		t.Error("should be able to extract GuardianError with errors.As")
	}

	if target.Code != ErrStorageOperation {
		t.Errorf("expected code %s, got %s", ErrStorageOperation, target.Code)
	}
}
