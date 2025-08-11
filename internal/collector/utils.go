package collector

import (
	"strings"
	"time"
)

// BoolToFloat converts a boolean to float64 (1.0 or 0.0).
func BoolToFloat(flag bool) float64 {
	if flag {
		return 1.0
	}
	return 0.0
}

// SafeString safely dereferences a *string.
// Returns an empty string if the pointer is nil.
func SafeString(ptr *string) string {
	if ptr != nil {
		return *ptr
	}
	return ""
}

// SafeBool safely dereferences a *bool.
// Returns false if the pointer is nil.
func SafeBool(ptr *bool) bool {
	if ptr != nil {
		return *ptr
	}
	return false
}

// SafeTime formats a *time.Time to RFC3339.
// Returns an empty string if the pointer is nil.
func SafeTime(t *time.Time) string {
	if t != nil {
		return t.Format(time.RFC3339)
	}
	return ""
}

// SafeUnixTime returns a Unix timestamp (seconds since epoch) for *time.Time.
// Returns 0 if the input is nil.
func SafeUnixTime(t *time.Time) int64 {
	if t != nil {
		return t.UTC().Unix()
	}
	return 0
}

// SafeTimeUnix formats a *time.Time into float64 UNIX seconds for Prometheus metrics.
// Returns 0.0 if the pointer is nil.
func SafeTimeUnix(t *time.Time) float64 {
	return float64(SafeUnixTime(t))
}

// SafeSlice safeguards iteration over possibly nil slices.
// Returns an empty slice if the input is nil.
func SafeSlice[T any](s *[]T) []T {
	if s != nil {
		return *s
	}
	return []T{}
}

// SafeJoin converts a []*string to comma-separated string safely
func SafeJoin(values *[]string) string {
	if values == nil || len(*values) == 0 {
		return ""
	}
	return strings.Join(*values, ",")
}
