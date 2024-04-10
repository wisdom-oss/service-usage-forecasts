package helpers

import (
	"os"

	"gopkg.in/yaml.v3"

	"github.com/wisdom-oss/service-usage-forecasts/types"
)

// GetAlgorithmMetadata reads the yaml metadata file supplied by the filepath
func GetAlgorithmMetadata(metadataFilePath string) (types.AlgorithmMetadata, error) {
	var metadata types.AlgorithmMetadata
	file, err := os.Open(metadataFilePath)
	if err != nil {
		return types.AlgorithmMetadata{}, err
	}
	err = yaml.NewDecoder(file).Decode(&metadata)
	if err != nil {
		return types.AlgorithmMetadata{}, err
	}
	return metadata, nil
}
