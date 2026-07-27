package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
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

// DefaultTimeout is the query timeout used when the datasource does not
// configure one.
const DefaultTimeout = 30 * time.Second

// JSONDataStruct holds the datasource options configured in the UI, as they are
// stored in the datasource `jsonData`.
type JSONDataStruct struct {
	Username string `json:"username"`
	Endpoint string `json:"endpoint"`
	Timeout  string `json:"timeout"`
}

// NewDatasource creates a new datasource instance.
func NewDatasource(_ context.Context, settings backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	// Variable to hold the unmarshaled data
	var jsonData JSONDataStruct

	// Unmarshal the JSON data into the struct
	if err := json.Unmarshal(settings.JSONData, &jsonData); err != nil {
		return nil, fmt.Errorf("error unmarshalling JSON data: %w", err)
	}

	// If the timeout is not set, use the default one
	timeout := DefaultTimeout
	if jsonData.Timeout != "" {
		ms, err := strconv.ParseInt(jsonData.Timeout, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("error parsing query timeout value: %w", err)
		}
		timeout = time.Duration(ms) * time.Millisecond
	}

	return NewDatasourceFor(jsonData.Endpoint, jsonData.Username, settings.DecryptedSecureJSONData["password"], timeout)
}

// NewDatasourceFor creates a datasource instance talking to the given SPARQL
// endpoint. Credentials are only sent when a username is configured.
func NewDatasourceFor(endpoint, username, password string, timeout time.Duration) (*Datasource, error) {
	if endpoint == "" {
		return nil, fmt.Errorf("no SPARQL endpoint configured")
	}

	options := []func(*sparql.Repo) error{
		sparql.Timeout(timeout),
	}
	if username != "" {
		options = append(options, sparql.DigestAuth(username, password))
	}

	repo, err := sparql.NewRepo(endpoint, options...)
	if err != nil {
		return nil, fmt.Errorf("error initializing SPARQL repo: %w", err)
	}

	return &Datasource{Repo: repo}, nil
}

// Datasource queries a SPARQL endpoint and turns the results into data frames.
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

// handleGraphQuery handles a CONSTRUCT or DESCRIBE query, whose result is an
// RDF graph rendered as subject/predicate/object columns.
func (d *Datasource) handleGraphQuery(query string) backend.DataResponse {
	var response backend.DataResponse

	// Execute the SPARQL query
	res, err := d.Repo.Construct(query)
	if err != nil {
		return backend.ErrDataResponse(backend.StatusBadRequest, fmt.Sprintf("SPARQL query execution: %v", err.Error()))
	}

	// Prepare the data frame for the results
	frame := data.NewFrame("response")

	// Create slices to hold the results
	subjects := make([]string, 0, len(res))
	predicates := make([]string, 0, len(res))
	objects := make([]string, 0, len(res))

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
	frame.Fields = append(frame.Fields,
		data.NewField("subject", nil, subjects),
		data.NewField("predicate", nil, predicates),
		data.NewField("object", nil, objects),
	)

	// Add the frame to the response
	response.Frames = append(response.Frames, frame)

	return response
}

// handleSolutionQuery handles a SELECT or an ASK query.
func (d *Datasource) handleSolutionQuery(query string) backend.DataResponse {
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
		// This is a SELECT query. Iterate over the solutions rather than over
		// `Bindings()`, which drops unbound values: a data frame requires every
		// field to have the same length, so an unbound variable has to become a
		// null value and not a missing row.
		solutions := res.Solutions()

		columns := make([][]*string, len(vars))
		for i := range columns {
			columns[i] = make([]*string, 0, len(solutions))
		}

		for _, solution := range solutions {
			for i, varName := range vars {
				term, bound := solution[varName]
				if !bound {
					columns[i] = append(columns[i], nil)
					continue
				}
				value := term.String()
				columns[i] = append(columns[i], &value)
			}
		}

		for i, varName := range vars {
			frame.Fields = append(frame.Fields, data.NewField(varName, nil, columns[i]))
		}
	}

	// Add the frame to the response
	response.Frames = append(response.Frames, frame)

	return response
}

func (d *Datasource) query(_ context.Context, _ backend.PluginContext, query backend.DataQuery) (response backend.DataResponse) {
	// Recover from a panic and report it as a query error, instead of taking
	// the whole plugin process down.
	defer func() {
		if r := recover(); r != nil {
			log.DefaultLogger.Error("panic while executing SPARQL query", "error", r)
			response = backend.ErrDataResponse(backend.StatusInternal, fmt.Sprintf("panic while executing SPARQL query: %v", r))
		}
	}()

	// Unmarshal the JSON into our queryModel.
	var qm queryModel
	if err := json.Unmarshal(query.JSON, &qm); err != nil {
		return backend.ErrDataResponse(backend.StatusBadRequest, fmt.Sprintf("json unmarshal: %v", err.Error()))
	}

	if d.Repo == nil {
		return backend.ErrDataResponse(backend.StatusValidationFailed, "the datasource has no SPARQL endpoint configured")
	}

	if DetectQueryForm(qm.QueryText).ReturnsGraph() {
		return d.handleGraphQuery(qm.QueryText)
	}

	return d.handleSolutionQuery(qm.QueryText)
}

// CheckHealth handles health checks sent from Grafana to the plugin.
// The main use case for these health checks is the test button on the
// datasource configuration page which allows users to verify that
// a datasource is working as expected.
func (d *Datasource) CheckHealth(_ context.Context, _ *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	if d.Repo == nil {
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: "No SPARQL endpoint is configured",
		}, nil
	}

	// Any SPARQL 1.1 endpoint can answer this query: being able to run it is
	// what proves the endpoint is reachable and speaks SPARQL. Note that the
	// boolean it returns is not checked, because an endpoint that holds no data
	// yet is still a healthy endpoint.
	if _, err := d.Repo.Query(`ASK WHERE { ?s ?p ?o }`); err != nil {
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: fmt.Sprintf("Failed to execute health check query: %v", err),
		}, nil
	}

	return &backend.CheckHealthResult{
		Status:  backend.HealthStatusOk,
		Message: "SPARQL endpoint is healthy",
	}, nil
}
