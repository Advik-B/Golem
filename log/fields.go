package log

import "go.uber.org/zap"

// Error is a helper to create a zap.Field for an error, matching the common "error" key.
func Error(err error) zap.Field {
	return zap.Error(err)
}
