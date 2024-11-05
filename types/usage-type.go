package types

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type UsageType struct {
	ID                 pgtype.UUID `json:"id" db:"id"`
	Name               string      `json:"name" db:"name"`
	Description        *string     `json:"description,omitempty" db:"description"`
	ExternalIdentifier string      `json:"externalIdentifier" db:"external_identifier"`
}
