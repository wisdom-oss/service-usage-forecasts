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
)

// InformationRoute allows users to check the capabilities and available scripts
// and identifiers for the different algorithms
func InformationRoute(w http.ResponseWriter, r *http.Request) {
	// access the error handlers
	nativeErrorChannel := r.Context().Value("nativeErrorChannel").(chan error)
	nativeErrorHandled := r.Context().Value("nativeErrorHandled").(chan bool)
	//wisdomErrorChannel := r.Context().Value("wisdomErrorChannel").(chan string)
	//wisdomErrorHandled := r.Context().Value("wisdomErrorHandled").(chan bool)

	entries, err := os.ReadDir(globals.Environment["INTERNAL_ALGORITHM_LOCATION"])
	if err != nil {
		nativeErrorChannel <- err
		<-nativeErrorHandled
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
		parameters, description, err := helpers.GetAlgorithmMetadata(metaFilePath)
		if err != nil {
			nativeErrorChannel <- err
			<-nativeErrorHandled
			return
		}
		algorithmInformation.Parameter = parameters
		algorithmInformation.Description = description
		algorithms = append(algorithms, algorithmInformation)
	}

	// since now all algorithms are

	// now respond with the algorithm information
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(algorithms)
	if err != nil {
		nativeErrorChannel <- fmt.Errorf("unable encode response: %w", err)
		<-nativeErrorHandled
		return
	}

}
