package plugin

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

// sparqlServer starts a fake SPARQL endpoint that answers every request with
// the given content type and body, and returns a datasource pointing at it.
func sparqlServer(t *testing.T, contentType, body string) *Datasource {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Errorf("could not parse the request form: %v", err)
		}
		if r.PostForm.Get("query") == "" {
			t.Error("the request carries no SPARQL query")
		}
		w.Header().Set("Content-Type", contentType)
		if _, err := w.Write([]byte(body)); err != nil {
			t.Errorf("could not write the response: %v", err)
		}
	}))
	t.Cleanup(server.Close)

	ds, err := NewDatasourceFor(server.URL, "", "", 5*time.Second)
	if err != nil {
		t.Fatalf("could not create the datasource: %v", err)
	}

	return ds
}

// runQuery executes a single query and returns its response.
func runQuery(t *testing.T, ds *Datasource, query string) backend.DataResponse {
	t.Helper()

	body, err := json.Marshal(queryModel{QueryText: query})
	if err != nil {
		t.Fatalf("could not marshal the query: %v", err)
	}

	resp, err := ds.QueryData(context.Background(), &backend.QueryDataRequest{
		Queries: []backend.DataQuery{{RefID: "A", JSON: body}},
	})
	if err != nil {
		t.Fatalf("QueryData returned an error: %v", err)
	}

	if len(resp.Responses) != 1 {
		t.Fatalf("QueryData returned %d responses, want 1", len(resp.Responses))
	}

	return resp.Responses["A"]
}

// fieldValues returns the string values of a frame field, using "<nil>" for
// null values so that they can be compared.
func fieldValues(t *testing.T, field *data.Field) []string {
	t.Helper()

	values := make([]string, field.Len())
	for i := range values {
		value, ok := field.ConcreteAt(i)
		if !ok {
			values[i] = "<nil>"
			continue
		}
		text, ok := value.(string)
		if !ok {
			t.Fatalf("field %q holds a %T, want a string", field.Name, value)
		}
		values[i] = text
	}

	return values
}

func TestQueryDataSelect(t *testing.T) {
	ds := sparqlServer(t, "application/sparql-results+json", `{
		"head": { "vars": [ "s", "p", "o" ] },
		"results": { "bindings": [
			{
				"s": { "type": "uri", "value": "http://example.org/a" },
				"p": { "type": "uri", "value": "http://example.org/name" },
				"o": { "type": "literal", "value": "Alice" }
			},
			{
				"s": { "type": "uri", "value": "http://example.org/b" },
				"p": { "type": "uri", "value": "http://example.org/name" },
				"o": { "type": "literal", "value": "Bob" }
			}
		] }
	}`)

	res := runQuery(t, ds, "SELECT * WHERE { ?s ?p ?o }")
	if res.Error != nil {
		t.Fatalf("query returned an error: %v", res.Error)
	}

	if len(res.Frames) != 1 {
		t.Fatalf("got %d frames, want 1", len(res.Frames))
	}

	frame := res.Frames[0]
	if len(frame.Fields) != 3 {
		t.Fatalf("got %d fields, want 3", len(frame.Fields))
	}

	for i, want := range []string{"s", "p", "o"} {
		if frame.Fields[i].Name != want {
			t.Errorf("field %d is named %q, want %q", i, frame.Fields[i].Name, want)
		}
	}

	if got := fieldValues(t, frame.Fields[2]); got[0] != "Alice" || got[1] != "Bob" {
		t.Errorf("object column = %v, want [Alice Bob]", got)
	}
	if got := fieldValues(t, frame.Fields[0]); got[0] != "http://example.org/a" {
		t.Errorf("subject column = %v, want it to start with http://example.org/a", got)
	}
}

// Unbound variables must become null values so that every column of the frame
// keeps the same length: Grafana rejects frames with mismatched field lengths.
func TestQueryDataSelectWithUnboundVariables(t *testing.T) {
	ds := sparqlServer(t, "application/sparql-results+json", `{
		"head": { "vars": [ "s", "label" ] },
		"results": { "bindings": [
			{
				"s": { "type": "uri", "value": "http://example.org/a" },
				"label": { "type": "literal", "value": "Alice" }
			},
			{
				"s": { "type": "uri", "value": "http://example.org/b" }
			},
			{
				"s": { "type": "uri", "value": "http://example.org/c" },
				"label": { "type": "literal", "value": "Carol" }
			}
		] }
	}`)

	res := runQuery(t, ds, "SELECT ?s ?label WHERE { ?s ?p ?o OPTIONAL { ?s rdfs:label ?label } }")
	if res.Error != nil {
		t.Fatalf("query returned an error: %v", res.Error)
	}

	frame := res.Frames[0]
	if _, err := frame.RowLen(); err != nil {
		t.Fatalf("the frame has mismatched field lengths: %v", err)
	}

	got := fieldValues(t, frame.Fields[1])
	want := []string{"Alice", "<nil>", "Carol"}
	if len(got) != len(want) {
		t.Fatalf("label column = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("label column = %v, want %v", got, want)
			break
		}
	}
}

func TestQueryDataAsk(t *testing.T) {
	ds := sparqlServer(t, "application/sparql-results+json",
		`{ "head": {}, "boolean": true }`)

	res := runQuery(t, ds, "ASK WHERE { ?s ?p ?o }")
	if res.Error != nil {
		t.Fatalf("query returned an error: %v", res.Error)
	}

	frame := res.Frames[0]
	if len(frame.Fields) != 1 || frame.Fields[0].Name != "boolean" {
		t.Fatalf("got %d fields, want a single 'boolean' field", len(frame.Fields))
	}

	value, ok := frame.Fields[0].ConcreteAt(0)
	if !ok || value != true {
		t.Errorf("boolean field = %v, want true", value)
	}
}

func TestQueryDataConstruct(t *testing.T) {
	ds := sparqlServer(t, "text/turtle",
		`<http://example.org/a> <http://example.org/name> "Alice" .`+"\n"+
			`<http://example.org/b> <http://example.org/name> "Bob" .`+"\n")

	res := runQuery(t, ds, "CONSTRUCT { ?s ?p ?o } WHERE { ?s ?p ?o }")
	if res.Error != nil {
		t.Fatalf("query returned an error: %v", res.Error)
	}

	frame := res.Frames[0]
	if len(frame.Fields) != 3 {
		t.Fatalf("got %d fields, want 3", len(frame.Fields))
	}

	for i, want := range []string{"subject", "predicate", "object"} {
		if frame.Fields[i].Name != want {
			t.Errorf("field %d is named %q, want %q", i, frame.Fields[i].Name, want)
		}
	}

	if rows, err := frame.RowLen(); err != nil || rows != 2 {
		t.Errorf("got %d rows (err: %v), want 2", rows, err)
	}
}

func TestQueryDataEndpointError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "malformed query", http.StatusBadRequest)
	}))
	t.Cleanup(server.Close)

	ds, err := NewDatasourceFor(server.URL, "", "", 5*time.Second)
	if err != nil {
		t.Fatalf("could not create the datasource: %v", err)
	}

	res := runQuery(t, ds, "SELECT * WHERE { this is not SPARQL")
	if res.Error == nil {
		t.Fatal("a failing endpoint must produce an error response")
	}
	if res.Status != backend.StatusBadRequest {
		t.Errorf("status = %v, want %v", res.Status, backend.StatusBadRequest)
	}
}

func TestQueryDataWithoutEndpoint(t *testing.T) {
	ds := &Datasource{}

	res := runQuery(t, ds, "SELECT * WHERE { ?s ?p ?o }")
	if res.Error == nil {
		t.Fatal("an unconfigured datasource must produce an error response")
	}
}

func TestCheckHealth(t *testing.T) {
	ds := sparqlServer(t, "application/sparql-results+json",
		`{ "head": {}, "boolean": false }`)

	res, err := ds.CheckHealth(context.Background(), &backend.CheckHealthRequest{})
	if err != nil {
		t.Fatalf("CheckHealth returned an error: %v", err)
	}

	// An endpoint that holds no data is still a healthy endpoint.
	if res.Status != backend.HealthStatusOk {
		t.Errorf("status = %v (%q), want %v", res.Status, res.Message, backend.HealthStatusOk)
	}
}

func TestCheckHealthUnreachableEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	t.Cleanup(server.Close)

	ds, err := NewDatasourceFor(server.URL, "", "", 5*time.Second)
	if err != nil {
		t.Fatalf("could not create the datasource: %v", err)
	}

	res, err := ds.CheckHealth(context.Background(), &backend.CheckHealthRequest{})
	if err != nil {
		t.Fatalf("CheckHealth must report the failure through the result, not as an error: %v", err)
	}
	if res.Status != backend.HealthStatusError {
		t.Errorf("status = %v, want %v", res.Status, backend.HealthStatusError)
	}
}

func TestCheckHealthWithoutEndpoint(t *testing.T) {
	ds := &Datasource{}

	res, err := ds.CheckHealth(context.Background(), &backend.CheckHealthRequest{})
	if err != nil {
		t.Fatalf("CheckHealth returned an error: %v", err)
	}
	if res.Status != backend.HealthStatusError {
		t.Errorf("status = %v, want %v", res.Status, backend.HealthStatusError)
	}
}

func TestNewDatasource(t *testing.T) {
	tests := []struct {
		name     string
		jsonData string
		wantErr  bool
	}{
		{name: "endpoint only", jsonData: `{"endpoint": "http://localhost:7878/query"}`},
		{name: "with a timeout", jsonData: `{"endpoint": "http://localhost:7878/query", "timeout": "5000"}`},
		{name: "with credentials", jsonData: `{"endpoint": "http://localhost:7878/query", "username": "admin"}`},
		{name: "missing endpoint", jsonData: `{}`, wantErr: true},
		{name: "invalid timeout", jsonData: `{"endpoint": "http://localhost:7878/query", "timeout": "soon"}`, wantErr: true},
		{name: "invalid json data", jsonData: `not json`, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			instance, err := NewDatasource(context.Background(), backend.DataSourceInstanceSettings{
				JSONData:                []byte(tt.jsonData),
				DecryptedSecureJSONData: map[string]string{"password": "hunter2"},
			})

			if tt.wantErr {
				if err == nil {
					t.Fatalf("NewDatasource(%s) succeeded, want an error", tt.jsonData)
				}
				return
			}

			if err != nil {
				t.Fatalf("NewDatasource(%s) returned an error: %v", tt.jsonData, err)
			}
			if _, ok := instance.(*Datasource); !ok {
				t.Fatalf("NewDatasource returned a %T, want a *Datasource", instance)
			}
		})
	}
}
