package routes

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/blockloop/scan/v2"
	"github.com/go-chi/chi/v5"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/wisdom-oss/service-usage-forecasts/globals"
	"github.com/wisdom-oss/service-usage-forecasts/helpers"
	"github.com/wisdom-oss/service-usage-forecasts/types"
)

// PredefinedForecast handles requests for predefined forecasts.
// this also includes external predefined forecast algorithms loaded during the
// startup
func PredefinedForecast(w http.ResponseWriter, r *http.Request) {
	// access the error handlers
	nativeErrorChannel := r.Context().Value("nativeErrorChannel").(chan error)
	nativeErrorHandled := r.Context().Value("nativeErrorHandled").(chan bool)
	wisdomErrorChannel := r.Context().Value("wisdomErrorChannel").(chan string)
	wisdomErrorHandled := r.Context().Value("wisdomErrorHandled").(chan bool)

	// get the municipals identifying the regions from which the water usages
	// shall be taken
	municipalKeys, keysSet := r.URL.Query()["key"]
	if !keysSet {
		wisdomErrorChannel <- "NO_AREA_DEFINED"
		<-wisdomErrorHandled
		return
	}

	// now create a regular expression for the municipalites
	var keyRegEx string
	for _, key := range municipalKeys {
		keyRegEx += fmt.Sprintf(`^%s\d*$|`, key)
	}
	keyRegEx = strings.Trim(keyRegEx, "|")

	// now get the consumer groups from the query parameters
	consumerGroups, consumerGroupsSet := r.URL.Query()["consumerGroup"]
	if consumerGroupsSet {
		// since there are consumer groups set, resolve the external identifiers
		// into the uuids that are used in the usage table
		rows, err := globals.SqlQueries.Query(globals.Db, "get-consumer-groups-by-external-id", consumerGroups)
		if err != nil {
			nativeErrorChannel <- err
			<-nativeErrorHandled
			return
		}
		// now get the usage type ids
		var usageTypes []types.UsageType
		err = scan.Rows(&usageTypes, rows)
		if err != nil {
			nativeErrorChannel <- fmt.Errorf("unable to parse usage types from database: %w", err)
			<-nativeErrorHandled
			return
		}
		var consumerGroupIDs []string
		for _, usageType := range usageTypes {
			consumerGroupIDs = append(consumerGroupIDs, usageType.ID.String())
		}
		// now reassign the resolved consumer group ids to the consumer groups
		consumerGroups = consumerGroupIDs
	} else {
		rows, err := globals.SqlQueries.Query(globals.Db, "get-consumer-groups")
		if err != nil {
			nativeErrorChannel <- err
			<-nativeErrorHandled
			return
		}
		// now get the usage type ids
		var usageTypes []types.UsageType
		err = scan.Rows(&usageTypes, rows)
		if err != nil {
			nativeErrorChannel <- fmt.Errorf("unable to parse usage types from database: %w", err)
			<-nativeErrorHandled
			return
		}
		var consumerGroupIDs []string
		for _, usageType := range usageTypes {
			consumerGroupIDs = append(consumerGroupIDs, usageType.ID.String())
		}
		// now reassign the resolved consumer group ids to the consumer groups
		consumerGroups = consumerGroupIDs
	}

	// since now all required data sets are available for getting the usage data,
	// validate that the algorithm selected is even present on the service

	// get the algorithm from the url parameters
	algorithmName := strings.TrimSpace(chi.URLParam(r, "algorithm-name"))
	if algorithmName == "" {
		wisdomErrorChannel <- "ALGORITHM_NOT_SET"
		<-wisdomErrorHandled
		return
	}

	// get all entries of the directory in which the algorithms are stored
	var algorithmFileName string
	entries, err := os.ReadDir(globals.Environment["INTERNAL_ALGORITHM_LOCATION"])
	if err != nil {
		nativeErrorChannel <- err
		<-nativeErrorHandled
		return
	}
	// now iterate over the entries
	for _, entry := range entries {
		// skip every entry that is a directory
		if entry.IsDir() {
			continue
		}

		// now check if the file name starts with the algorithm name and skip
		// every entry that does not start with the algorithm name
		if !strings.HasPrefix(entry.Name(), algorithmName) {
			continue
		}

		// now set the algorithm file name and exit the loop
		algorithmFileName = fmt.Sprintf("%s/%s", globals.Environment["INTERNAL_ALGORITHM_LOCATION"], entry.Name())
		break
	}

	// now check if the algorithm file name is still empty
	if strings.TrimSpace(algorithmFileName) == "" {
		wisdomErrorChannel <- "UNKNOWN_ALGORITHM"
		<-wisdomErrorHandled
		return
	}

	log.Debug().Msg("pulling usage data from the database")
	var rows *sql.Rows
	if !consumerGroupsSet {
		rows, err = globals.SqlQueries.Query(globals.Db, "get-usages-by-municipality", keyRegEx)
	} else {
		rows, err = globals.SqlQueries.Query(globals.Db, "get-usages-by-municipality-consumer-groups", keyRegEx, pq.Array(consumerGroups))
	}

	if err != nil {
		nativeErrorChannel <- err
		<-nativeErrorHandled
		return
	}

	// now parse the rows into an array of UsageDataPoint
	var usageDataPoints []types.UsageDataPoint
	err = scan.Rows(&usageDataPoints, rows)
	if err != nil {
		nativeErrorChannel <- fmt.Errorf("unable to parse usage data into structs: %w", err)
		<-nativeErrorHandled
		return
	}
	log.Debug().Msg("pulled usage data from the database")
	// now write the usage data points into a temporary json file
	log.Debug().Msg("writing usage data to file")
	tempDataFile, err := os.CreateTemp("", "forecast.*.input")
	defer tempDataFile.Close()
	if err != nil {
		nativeErrorChannel <- fmt.Errorf("unable to create temporary data file: %w", err)
		<-nativeErrorHandled
		return
	}
	err = json.NewEncoder(tempDataFile).Encode(usageDataPoints)
	if err != nil {
		nativeErrorChannel <- fmt.Errorf("unable to write usage data to file: %w", err)
		<-nativeErrorHandled
		return
	}
	log.Debug().Msg("wrote data to temporary file")

	// now create a temporary file which will contain the result of the
	// forecasting algorithm
	outputFileName := strings.ReplaceAll(tempDataFile.Name(), "input", "output")
	outputFile, err := os.Create(outputFileName)
	defer outputFile.Close()
	if err != nil {
		nativeErrorChannel <- fmt.Errorf("unable to open temporaray output file: %w", err)
		<-nativeErrorHandled
		return
	}

	// now create a temporary file for the parameters
	parameterFileName := strings.ReplaceAll(tempDataFile.Name(), "input", "parameter")
	parameterFile, err := os.Create(parameterFileName)
	defer parameterFile.Close()
	defer os.Remove(parameterFileName)
	if err != nil {
		nativeErrorChannel <- fmt.Errorf("unable to create parameter file: %w", err)
		<-nativeErrorHandled
		return
	}

	err = r.ParseMultipartForm(5242880)
	if err != nil && !errors.Is(err, http.ErrNotMultipart) {
		nativeErrorChannel <- fmt.Errorf("unable to parse form body: %w", err)
		<-nativeErrorHandled
		return
	}

	if r.MultipartForm.Value["parameter"] != nil {
		c, err := parameterFile.Write([]byte(r.MultipartForm.Value["parameter"][0]))
		log.Debug().Int("bytes", c).Msg("wrote parameter")
		if err != nil {
			nativeErrorChannel <- fmt.Errorf("write to parameter file: %w", err)
			<-nativeErrorHandled
			return
		}
	}

	// now call the algorithm
	log.Debug().Msg("calling algorithm")
	err = helpers.CallAlgorithm(algorithmFileName, tempDataFile.Name(), outputFile.Name(), parameterFile.Name())
	if err != nil {
		nativeErrorChannel <- fmt.Errorf("unable to run algorithm: %w", err)
		<-nativeErrorHandled
		return
	}
	log.Debug().Msg("algorithm finished")

	// now read the contents from the output file and send them directly back to the client
	w.Header().Set("Content-Type", "application/json")
	_, err = io.Copy(w, outputFile)
	if err != nil {
		nativeErrorChannel <- fmt.Errorf("unable to read results: %w", err)
		<-nativeErrorHandled
		return
	}
}
