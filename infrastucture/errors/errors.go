package errors

import "fmt"

type HTTPError struct {
	Code    int    `json:"code" example:"400"`
	Message string `json:"message" example:"status bad request"`
}

type domainError struct {
	Message string
	Error   error
}

func newDomainError(message, details string) *domainError {
	return &domainError{
		Message: message,
		Error:   fmt.Errorf(details),
	}
}

// OptimisticLockError is a custom error type for optimistic locking errors
type OptimisticLockError struct {
	*domainError
}

func NewOptimisticLockError(details string) *OptimisticLockError {
	return &OptimisticLockError{
		domainError: newDomainError("Opsie", details),
	}
}

// BusinessRuleError is a custom error type for business rule violations
type BusinessRuleError struct {
	*domainError
}

func NewBusinessRuleError(details string) *BusinessRuleError {
	return &BusinessRuleError{
		domainError: newDomainError("Business condemns!", details),
	}
}
