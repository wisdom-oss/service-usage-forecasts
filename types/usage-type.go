package types

import "github.com/google/uuid"

type UsageType struct {
	ID                 uuid.UUID `json:"id" db:"id"`
	Name               string    `json:"name" db:"name"`
	Description        *string   `json:"description,omitempty" db:"description"`
	ExternalIdentifier string    `json:"externalIdentifier" db:"external_identifier"`
}
