package types

type AlgorithmMetadata struct {
	// Description contains a description of the algorithm
	Description string `json:"description,omitempty"`

	// Parameters is an array that contains all parameters that may be passed to
	// the algorithm using the request body
	Parameters map[string]Parameter `json:"parameters"`

	// UseBuckets specifies if the data query uses buckets of time to
	// pre-aggregate the data it receives
	UseBuckets bool `json:"useBuckets" yaml:"useBuckets"`

	// BucketSize specifies the size of each bucket as a postgres interval
	BucketSize string `json:"bucketSize" yaml:"bucketSize"`
}
