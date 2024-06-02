package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"

	"github.com/knakk/sparql"
)

// Make sure Datasource implements required interfaces. This is important to do
// since otherwise we will only get a not implemented error response from plugin in
// runtime. In this example datasource instance implements backend.QueryDataHandler,
// backend.CheckHealthHandler interfaces. Plugin should not implement all these
// interfaces - only those which are required for a particular task.
var (
	_ backend.QueryDataHandler      = (*Datasource)(nil)
	_ backend.CheckHealthHandler    = (*Datasource)(nil)
	_ instancemgmt.InstanceDisposer = (*Datasource)(nil)
)

type JSONDataStruct struct {
	Username string `json:"username"`
	Endpint  string `json:"endpoint"`
}

// NewDatasource creates a new datasource instance.
func NewDatasource(_ context.Context, settings backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	// Variable to hold the unmarshaled data
	var jsonData JSONDataStruct

	// Unmarshal the JSON data into the struct
	err := json.Unmarshal([]byte(settings.JSONData), &jsonData)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling JSON data: %w", err)
	}

	// Those are the configured fields from the datasource options
	endpoint := jsonData.Endpint
	username := jsonData.Username
	password := settings.DecryptedSecureJSONData["password"]

	// Create a new SPARQL repo
	repo, err := sparql.NewRepo(endpoint,
		sparql.DigestAuth(username, password),
		sparql.Timeout(time.Millisecond*1500),
	)
	if err != nil {
		return nil, fmt.Errorf("error initializing SPARQL repo: %w", err)
	}

	return &Datasource{
		Repo: repo,
	}, nil
}

// Datasource is an example datasource which can respond to data queries, reports
// its health and has streaming skills.
type Datasource struct {
	Repo *sparql.Repo
}

// Dispose here tells plugin SDK that plugin wants to clean up resources when a new instance
// created. As soon as datasource settings change detected by SDK old datasource instance will
// be disposed and a new one will be created using NewSampleDatasource factory function.
func (d *Datasource) Dispose() {
	// Clean up datasource instance resources.
}

// QueryData handles multiple queries and returns multiple responses.
// req contains the queries []DataQuery (where each query contains RefID as a unique identifier).
// The QueryDataResponse contains a map of RefID to the response for each query, and each response
// contains Frames ([]*Frame).
func (d *Datasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	// create response struct
	response := backend.NewQueryDataResponse()

	// loop over queries and execute them individually.
	for _, q := range req.Queries {
		res := d.query(ctx, req.PluginContext, q)

		// save the response in a hashmap
		// based on with RefID as identifier
		response.Responses[q.RefID] = res
	}

	return response, nil
}

type queryModel struct {
	QueryText string `json:"queryText"`
}

// removeComments removes comments from a SPARQL query.
func removeComments(query string) string {
	re := regexp.MustCompile(`(?m)#.*$`)
	return re.ReplaceAllString(query, "")
}

// isConstructQuery checks if the given SPARQL query is a CONSTRUCT or a DESCRIBE query.
func isConstructQuery(query string) bool {
	// Convert the query to uppercase
	upperQuery := strings.ToUpper(query)
	// Remove comments from the query
	cleanQuery := removeComments(upperQuery)

	// Find the positions of keywords
	constructPos := strings.Index(cleanQuery, "CONSTRUCT")
	selectPos := strings.Index(cleanQuery, "SELECT")
	askPos := strings.Index(cleanQuery, "ASK")
	describePos := strings.Index(cleanQuery, "DESCRIBE")

	// Find the first occurrence of any keyword
	firstKeywordPos := -1
	if constructPos != -1 {
		firstKeywordPos = constructPos
	}
	if selectPos != -1 && (firstKeywordPos == -1 || selectPos < firstKeywordPos) {
		firstKeywordPos = selectPos
	}
	if askPos != -1 && (firstKeywordPos == -1 || askPos < firstKeywordPos) {
		firstKeywordPos = askPos
	}
	if describePos != -1 && (firstKeywordPos == -1 || describePos < firstKeywordPos) {
		firstKeywordPos = describePos
	}

	// Check if the first keyword is CONSTRUCT or DESCRIBE
	return firstKeywordPos == constructPos || firstKeywordPos == describePos
}

// handleConstructQuery handles a CONSTRUCT or DESCRIBE query.
func handleConstructQuery(d *Datasource, query string) backend.DataResponse {
	var response backend.DataResponse

	// Execute the SPARQL query
	res, err := d.Repo.Construct(query)
	if err != nil {
		return backend.ErrDataResponse(backend.StatusBadRequest, fmt.Sprintf("SPARQL query execution: %v", err.Error()))
	}

	// Prepare the data frame for the results
	frame := data.NewFrame("response")

	// Get the number of results
	// nbResults := len(res)

	// Create slices to hold the results
	subjects := make([]string, 0)
	predicates := make([]string, 0)
	objects := make([]string, 0)

	// Add each triple one by one to each slice
	for _, triple := range res {
		s := triple.Subj.String()
		p := triple.Pred.String()
		o := triple.Obj.String()

		// Skip empty triples
		if s == "" && p == "" && o == "" {
			continue
		}

		subjects = append(subjects, s)
		predicates = append(predicates, p)
		objects = append(objects, o)
	}

	// Add the slices to the frame
	frame.Fields = append(frame.Fields, data.NewField("subject", nil, subjects))
	frame.Fields = append(frame.Fields, data.NewField("predicate", nil, predicates))
	frame.Fields = append(frame.Fields, data.NewField("object", nil, objects))

	// Add the frame to the response
	response.Frames = append(response.Frames, frame)

	return response
}

// handleGenericQuery handles a generic SPARQL query (ASK or SELECT).
func handleGenericQuery(d *Datasource, query string) backend.DataResponse {
	var response backend.DataResponse

	// Execute the SPARQL query
	res, err := d.Repo.Query(query)
	if err != nil {
		return backend.ErrDataResponse(backend.StatusBadRequest, fmt.Sprintf("SPARQL query execution: %v", err.Error()))
	}

	// Prepare the data frame for the results
	frame := data.NewFrame("response")

	vars := res.Head.Vars

	if len(vars) == 0 {
		// This is a boolean result (ASK query)
		frame.Fields = append(frame.Fields, data.NewField("boolean", nil, []bool{res.Boolean}))
	} else {
		// This is a SELECT query
		bindings := res.Bindings()

		for _, varName := range vars {
			// Get the values for the variable
			values := bindings[varName]

			// Create a slice to hold the results
			results := make([]string, len(values))

			// Trasform the values to strings
			for i, value := range values {
				results[i] = value.String()
			}

			frame.Fields = append(frame.Fields, data.NewField(varName, nil, results))
		}
	}

	// Add the frame to the response
	response.Frames = append(response.Frames, frame)

	return response
}

func (d *Datasource) query(_ context.Context, _ backend.PluginContext, query backend.DataQuery) backend.DataResponse {
	// Recover from panic, and log the error
	defer func() {
		if r := recover(); r != nil {
			log.DefaultLogger.Error(fmt.Sprintf(">>>>>>>> PANIC!!!: %v", r))
		}
	}()

	// Unmarshal the JSON into our queryModel.
	var qm queryModel
	err := json.Unmarshal(query.JSON, &qm)
	if err != nil {
		return backend.ErrDataResponse(backend.StatusBadRequest, fmt.Sprintf("json unmarshal: %v", err.Error()))
	}

	sparqlQuery := qm.QueryText

	if isConstructQuery(sparqlQuery) {
		return handleConstructQuery(d, sparqlQuery)
	}

	return handleGenericQuery(d, sparqlQuery)
}

// CheckHealth handles health checks sent from Grafana to the plugin.
// The main use case for these health checks is the test button on the
// datasource configuration page which allows users to verify that
// a datasource is working as expected.
func (d *Datasource) CheckHealth(ctx context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	// Define a simple SPARQL query to check the health of the endpoint
	query := `ASK WHERE { ?s ?p ?o }`

	// Execute the query using the SPARQL client
	res, err := d.Repo.Query(query)
	if err != nil {
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: fmt.Sprintf("Failed to execute health check query: %v", err),
		}, err
	}

	// Check if the endpoint returned a valid response
	if !res.Boolean {
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: "SPARQL endpoint did not return a valid response",
		}, fmt.Errorf("SPARQL endpoint did not return a valid response")
	}

	// If everything is working as expected, return a healthy status
	return &backend.CheckHealthResult{
		Status:  backend.HealthStatusOk,
		Message: "SPARQL endpoint is healthy",
	}, nil
}
