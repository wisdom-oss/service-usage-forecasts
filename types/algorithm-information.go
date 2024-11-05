package types

type AlgorithmInformation struct {
	// Filename contains the filename in which the algorithm is stored
	Filename string `json:"-"`

	// Description contains a description of the algorithm
	Description string `json:"description"`

	// Identifier contains the identification of the algorithm for logs,
	//  requests, and other purposes
	Identifier string `json:"identifier"`

	// BucketConfiguration
	BucketConfiguration struct {
		UseBuckets bool   `json:"useBuckets"`
		BucketSize string `json:"bucketSize,omitempty"`
	} `json:"-"`

	// Parameter describes the parameters and is directly read from the
	// file containing the metadata
	Parameter map[string]Parameter `json:"parameter"`
}
