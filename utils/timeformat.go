package utils

import (
	"fmt"
	"time"
)

func ConvertRFC3339ToDatetime(rfc3339 string) (string, error) {
	t, err := time.Parse(time.RFC3339Nano, rfc3339)
	if err != nil {
		return "", fmt.Errorf("failed to parse time: %w", err)
	}
	return t.Format("2006-01-02 15:04:05"), nil
}
