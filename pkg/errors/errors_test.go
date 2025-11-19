package errors

import (
	"errors"
	"net/http"
	"testing"
)

func TestNewAppError(t *testing.T) {
	tests := []struct {
		name       string
		code       string
		message    string
		statusCode int
		wantCode   string
		wantMsg    string
		wantStatus int
	}{
		{
			name:       "creates validation error",
			code:       "VALIDATION_ERROR",
			message:    "invalid input",
			statusCode: http.StatusBadRequest,
			wantCode:   "VALIDATION_ERROR",
			wantMsg:    "invalid input",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "creates not found error",
			code:       "NOT_FOUND",
			message:    "resource not found",
			statusCode: http.StatusNotFound,
			wantCode:   "NOT_FOUND",
			wantMsg:    "resource not found",
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewAppError(tt.code, tt.message, tt.statusCode)

			if err.Code != tt.wantCode {
				t.Errorf("Code = %v, want %v", err.Code, tt.wantCode)
			}
			if err.Message != tt.wantMsg {
				t.Errorf("Message = %v, want %v", err.Message, tt.wantMsg)
			}
			if err.StatusCode != tt.wantStatus {
				t.Errorf("StatusCode = %v, want %v", err.StatusCode, tt.wantStatus)
			}
		})
	}
}

func TestAppError_Error(t *testing.T) {
	err := NewAppError("TEST_ERROR", "test message", http.StatusBadRequest)
	want := "TEST_ERROR: test message"

	if got := err.Error(); got != want {
		t.Errorf("Error() = %v, want %v", got, want)
	}
}

func TestAppError_WithMetadata(t *testing.T) {
	err := NewAppError("TEST_ERROR", "test", http.StatusBadRequest)
	metadata := map[string]interface{}{
		"field": "email",
		"value": "invalid",
	}

	err = err.WithMetadata(metadata)

	if err.Metadata == nil {
		t.Fatal("Metadata should not be nil")
	}
	if err.Metadata["field"] != "email" {
		t.Errorf("Metadata[field] = %v, want email", err.Metadata["field"])
	}
	if err.Metadata["value"] != "invalid" {
		t.Errorf("Metadata[value] = %v, want invalid", err.Metadata["value"])
	}
}

func TestAppError_Wrap(t *testing.T) {
	originalErr := errors.New("database connection failed")
	appErr := NewAppError("DB_ERROR", "database error", http.StatusInternalServerError)

	wrappedErr := appErr.Wrap(originalErr)

	if wrappedErr.Cause == nil {
		t.Fatal("Cause should not be nil")
	}
	if wrappedErr.Cause.Error() != originalErr.Error() {
		t.Errorf("Cause = %v, want %v", wrappedErr.Cause, originalErr)
	}

	// Test unwrapping
	if !errors.Is(wrappedErr, originalErr) {
		t.Error("Wrapped error should be identifiable with errors.Is")
	}
}

func TestIsAppError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "returns true for AppError",
			err:  NewAppError("TEST", "test", http.StatusBadRequest),
			want: true,
		},
		{
			name: "returns false for standard error",
			err:  errors.New("standard error"),
			want: false,
		},
		{
			name: "returns false for nil",
			err:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsAppError(tt.err); got != tt.want {
				t.Errorf("IsAppError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCommonErrors(t *testing.T) {
	tests := []struct {
		name       string
		fn         func(string) *AppError
		message    string
		wantStatus int
	}{
		{
			name:       "BadRequest",
			fn:         BadRequest,
			message:    "invalid input",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "NotFound",
			fn:         NotFound,
			message:    "resource not found",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "Unauthorized",
			fn:         Unauthorized,
			message:    "unauthorized access",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "Forbidden",
			fn:         Forbidden,
			message:    "forbidden",
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "Conflict",
			fn:         Conflict,
			message:    "resource conflict",
			wantStatus: http.StatusConflict,
		},
		{
			name:       "InternalServerError",
			fn:         InternalServerError,
			message:    "internal error",
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fn(tt.message)
			if err.StatusCode != tt.wantStatus {
				t.Errorf("StatusCode = %v, want %v", err.StatusCode, tt.wantStatus)
			}
			if err.Message != tt.message {
				t.Errorf("Message = %v, want %v", err.Message, tt.message)
			}
		})
	}
}
