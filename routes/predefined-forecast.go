package routes

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/go-chi/chi/v5"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	wisdomType "github.com/wisdom-oss/commonTypes/v2"
	wisdomMiddlware "github.com/wisdom-oss/microservice-middlewares/v4"
	"gopkg.in/yaml.v3"

	"github.com/wisdom-oss/service-usage-forecasts/globals"
	"github.com/wisdom-oss/service-usage-forecasts/helpers"
	"github.com/wisdom-oss/service-usage-forecasts/types"
)

// ErrNoAreaSelected is an error that occurs when the request did not specify
// the area for which the prognosis shall be executed.
var ErrNoAreaSelected = wisdomType.WISdoMError{
	Type:   "https://www.rfc-editor.org/rfc/rfc9110#section-15.5.1",
	Status: http.StatusBadRequest,
	Title:  "No Area Selected",
	Detail: "The request did not specify the area for which the prognosis shall be executed, this is not allowed",
}

// ErrNoAlgorithmSpecified is an error that occurs when the request did not
// contain an identifier for an algorithm.
var ErrNoAlgorithmSpecified = wisdomType.WISdoMError{
	Type:   "https://www.rfc-editor.org/rfc/rfc9110#section-15.5.1",
	Status: http.StatusBadRequest,
	Title:  "No Algorithm Specified",
	Detail: "The request did not contain a identifier for an algorithm",
}

// ErrUnknownAlgorithm is an error that occurs when the algorithm specified in
// the request does not exist on the server.
// Please check your request and make sure that the requested script is stored
// on the server
var ErrUnknownAlgorithm = wisdomType.WISdoMError{
	Type:   "https://www.rfc-editor.org/rfc/rfc9110#section-15.5.5",
	Status: http.StatusNotFound,
	Title:  "Unknown Algorithm",
	Detail: "The algorithm specified in the request does not exist on the server. Please check your request and make sure that the requested script is stored on the server",
}

var ErrInvalidBucketSize = wisdomType.WISdoMError{
	Type:   "https://www.rfc-editor.org/rfc/rfc9110#section-15.5.1",
	Status: http.StatusBadRequest,
	Title:  "Invalid Bucket Size",
	Detail: "The amount of seconds provided for the size of the bucket is not valid. Please check the documentation",
}

// PredefinedForecast handles requests for predefined forecasts.
// this also includes external predefined forecast algorithms loaded during the
// startup
func PredefinedForecast(w http.ResponseWriter, r *http.Request) {
	// access the error handlers
	errorHandler := r.Context().Value(wisdomMiddlware.ErrorChannelName).(chan<- interface{})
	statusChannel := r.Context().Value(wisdomMiddlware.StatusChannelName).(<-chan bool)

	// get the municipals identifying the regions from which the water usages
	// shall be taken
	municipalKeys, keysSet := r.URL.Query()["key"]
	if !keysSet {
		errorHandler <- ErrNoAreaSelected
		<-statusChannel
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
		query, err := globals.SqlQueries.Raw("get-consumer-groups-by-external-id")
		if err != nil {
			errorHandler <- err
			<-statusChannel
			return
		}
		// now get the usage type ids
		var usageTypes []types.UsageType
		err = pgxscan.Select(r.Context(), globals.Db, &usageTypes, query, consumerGroups)
		if err != nil {
			errorHandler <- fmt.Errorf("unable to query usage types from database: %w", err)
			<-statusChannel
			return
		}
		var consumerGroupIDs []string
		for _, usageType := range usageTypes {
			uuid, _ := usageType.ID.Value()
			consumerGroupIDs = append(consumerGroupIDs, uuid.(string))
		}
		// now reassign the resolved consumer group ids to the consumer groups
		consumerGroups = consumerGroupIDs
	} else {
		query, err := globals.SqlQueries.Raw("get-consumer-groups")
		if err != nil {
			errorHandler <- err
			<-statusChannel
			return
		}
		var usageTypes []types.UsageType
		err = pgxscan.Select(r.Context(), globals.Db, &usageTypes, query)
		if err != nil {
			errorHandler <- fmt.Errorf("unable to query usage types from database: %w", err)
			<-statusChannel
			return
		}
		var consumerGroupIDs []string
		for _, usageType := range usageTypes {
			uuid, _ := usageType.ID.Value()
			consumerGroupIDs = append(consumerGroupIDs, uuid.(string))
		}
		// now reassign the resolved consumer group ids to the consumer groups
		consumerGroups = consumerGroupIDs
	}

	// since now all required data sets are available for getting the usage data,
	// validate that the algorithm selected is even present on the service

	// get the algorithm from the url parameters
	algorithmName := strings.TrimSpace(chi.URLParam(r, "algorithm-name"))
	if algorithmName == "" {
		errorHandler <- ErrNoAlgorithmSpecified
		<-statusChannel
		return
	}

	// get all entries of the directory in which the algorithms are stored
	var algorithmFileName string
	entries, err := os.ReadDir(globals.Environment["INTERNAL_ALGORITHM_LOCATION"])
	if err != nil {
		errorHandler <- err
		<-statusChannel
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
		errorHandler <- ErrUnknownAlgorithm
		<-statusChannel
		return
	}

	metaFilePath := fmt.Sprintf("%s/%s.yaml", globals.Environment["INTERNAL_ALGORITHM_LOCATION"], algorithmName)
	var metadata types.AlgorithmMetadata
	file, err := os.Open(metaFilePath)
	if err != nil {
		errorHandler <- fmt.Errorf("unable to open script metadata: %w", err)
		<-statusChannel
		return
	}
	err = yaml.NewDecoder(file).Decode(&metadata)
	if err != nil {
		errorHandler <- fmt.Errorf("unable to parse script metadata: %w", err)
		<-statusChannel
		return
	}

	log.Debug().Msg("pulling usage data from the database")
	var query string
	var args []interface{}

	switch {
	case metadata.UseBuckets && consumerGroupsSet:
		query, err = globals.SqlQueries.Raw("get-bucketed-usages-by-municipality-consumer-groups")
		args = []interface{}{metadata.BucketSize, keyRegEx, consumerGroupsSet}
		break
	case metadata.UseBuckets && !consumerGroupsSet:
		query, err = globals.SqlQueries.Raw("get-bucketed-usages-by-municipality")
		args = []interface{}{metadata.BucketSize, keyRegEx}
		break
	case !metadata.UseBuckets && consumerGroupsSet:
		query, err = globals.SqlQueries.Raw("get-usages-by-municipality-consumer-groups")
		args = []interface{}{keyRegEx, consumerGroups}
		break
	case !metadata.UseBuckets && !consumerGroupsSet:
		query, err = globals.SqlQueries.Raw("get-usages-by-municipality")
		args = []interface{}{keyRegEx}
	}

	if err != nil {
		errorHandler <- fmt.Errorf("unable to prepare query for usage data: %w", err)
		<-statusChannel
		return
	}

	var usageDataPoints []types.UsageDataPoint
	err = pgxscan.Select(r.Context(), globals.Db, &usageDataPoints, query, args...)

	if err != nil {
		errorHandler <- err
		<-statusChannel
		return
	}

	log.Debug().Msg("pulled usage data from the database")
	// now write the usage data points into a temporary json file
	log.Debug().Msg("writing usage data to file")
	tempDataFile, err := os.CreateTemp("", "forecast.*.input")
	defer tempDataFile.Close()
	if err != nil {
		errorHandler <- fmt.Errorf("unable to create temporary data file: %w", err)
		<-statusChannel
		return
	}
	err = json.NewEncoder(tempDataFile).Encode(usageDataPoints)
	if err != nil {
		errorHandler <- fmt.Errorf("unable to write usage data to file: %w", err)
		<-statusChannel
		return
	}
	log.Debug().Msg("wrote data to temporary file")

	// now create a temporary file which will contain the result of the
	// forecasting algorithm
	outputFileName := strings.ReplaceAll(tempDataFile.Name(), "input", "output")
	outputFile, err := os.Create(outputFileName)
	defer outputFile.Close()
	if err != nil {
		errorHandler <- fmt.Errorf("unable to open temporaray output file: %w", err)
		<-statusChannel
		return
	}

	// now create a temporary file for the parameters
	parameterFileName := strings.ReplaceAll(tempDataFile.Name(), "input", "parameter")
	parameterFile, err := os.Create(parameterFileName)
	defer parameterFile.Close()
	defer os.Remove(parameterFileName)
	if err != nil {
		errorHandler <- fmt.Errorf("unable to create parameter file: %w", err)
		<-statusChannel
		return
	}

	err = r.ParseMultipartForm(5242880)
	if err != nil && !errors.Is(err, http.ErrNotMultipart) {
		errorHandler <- fmt.Errorf("unable to parse form body: %w", err)
		<-statusChannel
		return
	}

	if r.Method == "POST" && r.MultipartForm != nil && r.MultipartForm.Value["parameter"] != nil {
		c, err := parameterFile.Write([]byte(r.MultipartForm.Value["parameter"][0]))
		log.Debug().Int("bytes", c).Msg("wrote parameter")
		if err != nil {
			errorHandler <- fmt.Errorf("write to parameter file: %w", err)
			<-statusChannel
			return
		}
	}

	// now call the algorithm
	log.Debug().Msg("calling algorithm")
	err = helpers.CallAlgorithm(algorithmFileName, tempDataFile.Name(), outputFile.Name(), parameterFile.Name())
	if err != nil {
		errorHandler <- fmt.Errorf("unable to run algorithm: %w", err)
		<-statusChannel
		return
	}
	log.Debug().Msg("algorithm finished")

	// now read the contents from the output file and send them directly back to the client
	w.Header().Set("Content-Type", "application/json")
	_, err = io.Copy(w, outputFile)
	if err != nil {
		errorHandler <- fmt.Errorf("unable to read results: %w", err)
		<-statusChannel
		return
	}
}
