// Package errors provides standardized exit codes for the feat CLI.
package errors

// Exit codes for the feat CLI.
const (
	ExitSuccess          = 0 // Successful completion
	ExitGeneralError     = 1 // General error
	ExitInvalidConfig    = 2 // Invalid configuration
	ExitContextLimit     = 3 // Context limit exceeded
	ExitFeatureNotFound  = 4 // Feature not found
	ExitCircularReference = 5 // Circular reference detected
)