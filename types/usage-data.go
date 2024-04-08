package types

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type UsageDataPoint struct {
	Municipal string             `json:"municipal" db:"municipality"`
	UsageType pgtype.UUID        `json:"usageType" db:"usage_type"`
	Date      pgtype.Timestamptz `json:"date" db:"time"`
	Amount    float64            `json:"amount" db:"amount"`
}
