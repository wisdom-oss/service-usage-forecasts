package types

import "time"

type UsageDataPoint struct {
	Municipal string    `json:"municipal" db:"municipality"`
	UsageType string    `json:"usageType" db:"usage_type"`
	Date      time.Time `json:"date" db:"date"`
	Amount    float64   `json:"amount" db:"amount"`
}
