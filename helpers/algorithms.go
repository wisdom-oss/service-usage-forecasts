package helpers

import (
	"os"

	"gopkg.in/yaml.v3"

	"github.com/wisdom-oss/service-usage-forecasts/types"
)

// GetAlgorithmMetadata reads the yaml metadata file supplied by the filepath
func GetAlgorithmMetadata(metadataFilePath string) (map[string]types.Parameter, string, error) {
	var metadata types.AlgorithmMetadata
	file, err := os.Open(metadataFilePath)
	if err != nil {
		return nil, "", nil
	}
	err = yaml.NewDecoder(file).Decode(&metadata)
	if err != nil {
		return nil, "", err
	}
	return metadata.Parameters, metadata.Description, nil
}
