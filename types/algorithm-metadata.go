package types

type AlgorithmMetadata struct {
	// Description contains a description of the algorithm
	Description string `json:"description,omitempty"`

	// Parameters is an array that contains all parameters that may be passed to
	// the algorithm using the request body
	Parameters map[string]Parameter `json:"parameters"`
}
