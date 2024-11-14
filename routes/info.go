package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/wisdom-oss/service-usage-forecasts/globals"
	"github.com/wisdom-oss/service-usage-forecasts/helpers"
	"github.com/wisdom-oss/service-usage-forecasts/types"

	wisdomMiddlware "github.com/wisdom-oss/microservice-middlewares/v4"
)

// InformationRoute allows users to check the capabilities and available scripts
// and identifiers for the different algorithms
func InformationRoute(w http.ResponseWriter, r *http.Request) {
	// access the error handlers
	errorHandler := r.Context().Value(wisdomMiddlware.ErrorChannelName).(chan<- interface{})
	statusChannel := r.Context().Value(wisdomMiddlware.StatusChannelName).(<-chan bool)

	entries, err := os.ReadDir(globals.Environment["INTERNAL_ALGORITHM_LOCATION"])
	if err != nil {
		errorHandler <- err
		<-statusChannel
		return
	}
	var algorithms []types.AlgorithmInformation
	// now iterate over the entries
	for _, entry := range entries {
		// skip every entry that is a directory
		if entry.IsDir() {
			continue
		}

		// now check if the file name starts with the algorithm name and skip
		// every entry that does not start with the algorithm name
		if !strings.HasSuffix(entry.Name(), ".py") && !strings.HasSuffix(entry.Name(), ".rscript") {
			continue
		}

		// now create an emtpy algorithm information object
		var algorithmInformation types.AlgorithmInformation
		algorithmInformation.Filename = entry.Name()
		algorithmInformation.Identifier = strings.SplitN(entry.Name(), ".", 2)[0]
		metaFilePath := fmt.Sprintf("%s/%s.yaml", globals.Environment["INTERNAL_ALGORITHM_LOCATION"], algorithmInformation.Identifier)
		metadata, err := helpers.GetAlgorithmMetadata(metaFilePath)
		if err != nil {
			errorHandler <- err
			<-statusChannel
			return
		}
		algorithmInformation.DisplayName = metadata.DisplayName
		algorithmInformation.Parameter = metadata.Parameters
		algorithmInformation.Description = metadata.Description
		algorithmInformation.BucketConfiguration.UseBuckets = metadata.UseBuckets
		algorithmInformation.BucketConfiguration.BucketSize = metadata.BucketSize
		algorithms = append(algorithms, algorithmInformation)
	}

	// now respond with the algorithm information
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(algorithms)
	if err != nil {
		errorHandler <- fmt.Errorf("unable encode response: %w", err)
		<-statusChannel
		return
	}

}
