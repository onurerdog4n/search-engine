package errors

import (
	"errors"
	"fmt"
)

// Domain errors - business logic errors
var (
	ErrContentNotFound     = errors.New("content not found")
	ErrInvalidProvider     = errors.New("invalid provider format")
	ErrRateLimitExceeded   = errors.New("rate limit exceeded")
	ErrInvalidSearchParams = errors.New("invalid search parameters")
	ErrProviderNotActive   = errors.New("provider is not active")
	ErrDuplicateContent    = errors.New("content already exists")
)

// ValidationError represents a validation error with field-level details
type ValidationError struct {
	Field   string
	Message string
	Value   interface{}
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error on field '%s': %s (value: %v)", e.Field, e.Message, e.Value)
}

// NewValidationError creates a new validation error
func NewValidationError(field, message string, value interface{}) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
		Value:   value,
	}
}

// ProviderError represents an error from a provider
type ProviderError struct {
	ProviderName string
	Operation    string
	Err          error
}

func (e *ProviderError) Error() string {
	return fmt.Sprintf("provider '%s' failed during %s: %v", e.ProviderName, e.Operation, e.Err)
}

func (e *ProviderError) Unwrap() error {
	return e.Err
}

// NewProviderError creates a new provider error
func NewProviderError(providerName, operation string, err error) *ProviderError {
	return &ProviderError{
		ProviderName: providerName,
		Operation:    operation,
		Err:          err,
	}
}

// DatabaseError represents a database operation error
type DatabaseError struct {
	Operation string
	Table     string
	Err       error
}

func (e *DatabaseError) Error() string {
	return fmt.Sprintf("database error during %s on table '%s': %v", e.Operation, e.Table, e.Err)
}

func (e *DatabaseError) Unwrap() error {
	return e.Err
}

// NewDatabaseError creates a new database error
func NewDatabaseError(operation, table string, err error) *DatabaseError {
	return &DatabaseError{
		Operation: operation,
		Table:     table,
		Err:       err,
	}
}

// CacheError represents a cache operation error
type CacheError struct {
	Operation string
	Key       string
	Err       error
}

func (e *CacheError) Error() string {
	return fmt.Sprintf("cache error during %s for key '%s': %v", e.Operation, e.Key, e.Err)
}

func (e *CacheError) Unwrap() error {
	return e.Err
}

// NewCacheError creates a new cache error
func NewCacheError(operation, key string, err error) *CacheError {
	return &CacheError{
		Operation: operation,
		Key:       key,
		Err:       err,
	}
}
