package types

type Parameter struct {
	// Description contains a description of the parameter. This text should
	// contain important information on how to use the paramteter and what
	// effects it has.
	Description string `yaml:"description" json:"description"`

	// DefaultValue contains the default value of the parameter. Since the
	// algorithm may accept any type, this value is just an interface
	DefaultValue interface{} `yaml:"default" json:"default"`

	// Type denotes the python-specific data type that the parameter uses
	Type string `yaml:"type" json:"type"`
}
