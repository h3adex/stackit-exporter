package utils

import "time"

func BoolToFloat(flag bool) float64 {
	if flag {
		return 1
	}
	return 0
}

func SafeString(ptr *string) string {
	if ptr != nil {
		return *ptr
	}
	return ""
}

func SafeTime(t *time.Time) string {
	if t != nil {
		return t.Format(time.RFC3339)
	}
	return ""
}
