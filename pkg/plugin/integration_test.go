package plugin

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// These tests run against a real SPARQL endpoint. Start one with
// `docker compose up -d oxigraph` and point SPARQL_TEST_ENDPOINT at it:
//
//	SPARQL_TEST_ENDPOINT=http://localhost:7878/query go test ./pkg/...
//
// They are skipped when the variable is not set, so that `go test ./...` keeps
// working without any external service.
const (
	endpointEnv       = "SPARQL_TEST_ENDPOINT"
	updateEndpointEnv = "SPARQL_TEST_UPDATE_ENDPOINT"

	// testGraph holds the data these tests insert, so that they neither depend
	// on nor disturb whatever else the endpoint already stores.
	testGraph = "http://example.org/graph/integration-test"
)

// integrationDatasource returns a datasource pointing at the SPARQL endpoint
// under test, seeded with the test data, or skips the test when no endpoint is
// configured.
func integrationDatasource(t *testing.T) *Datasource {
	t.Helper()

	endpoint := os.Getenv(endpointEnv)
	if endpoint == "" {
		t.Skipf("%s is not set, skipping the integration test", endpointEnv)
	}

	updateEndpoint := os.Getenv(updateEndpointEnv)
	if updateEndpoint == "" {
		// Oxigraph, Fuseki, GraphDB and friends all expose updates next to the
		// query endpoint.
		updateEndpoint = strings.TrimSuffix(endpoint, "/query") + "/update"
	}

	seed(t, updateEndpoint)

	ds, err := NewDatasourceFor(endpoint, "", "", 30*time.Second)
	if err != nil {
		t.Fatalf("could not create the datasource: %v", err)
	}

	return ds
}

// seed loads the test data into its own named graph and removes it again once
// the test is over.
func seed(t *testing.T, updateEndpoint string) {
	t.Helper()

	sparqlUpdate(t, updateEndpoint, fmt.Sprintf(`DROP SILENT GRAPH <%s>`, testGraph))
	sparqlUpdate(t, updateEndpoint, fmt.Sprintf(`
		PREFIX ex:   <http://example.org/>
		PREFIX foaf: <http://xmlns.com/foaf/0.1/>
		PREFIX xsd:  <http://www.w3.org/2001/XMLSchema#>
		INSERT DATA {
			GRAPH <%s> {
				ex:alice a foaf:Person ; foaf:name "Alice" ; ex:age "34"^^xsd:integer ; foaf:knows ex:bob .
				ex:bob   a foaf:Person ; foaf:name "Bob"   ; ex:age "28"^^xsd:integer .
				ex:carol a foaf:Person ; foaf:name "Carol" .
			}
		}`, testGraph))

	t.Cleanup(func() {
		sparqlUpdate(t, updateEndpoint, fmt.Sprintf(`DROP SILENT GRAPH <%s>`, testGraph))
	})
}

func sparqlUpdate(t *testing.T, updateEndpoint, update string) {
	t.Helper()

	form := url.Values{"update": {update}}
	resp, err := http.Post(updateEndpoint, "application/x-www-form-urlencoded", strings.NewReader(form.Encode()))
	if err != nil {
		t.Fatalf("SPARQL update request to %s failed: %v", updateEndpoint, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 300 {
		t.Fatalf("SPARQL update request to %s returned %s", updateEndpoint, resp.Status)
	}
}

func TestIntegrationCheckHealth(t *testing.T) {
	ds := integrationDatasource(t)

	res, err := ds.CheckHealth(context.Background(), &backend.CheckHealthRequest{})
	if err != nil {
		t.Fatalf("CheckHealth returned an error: %v", err)
	}
	if res.Status != backend.HealthStatusOk {
		t.Fatalf("status = %v (%q), want %v", res.Status, res.Message, backend.HealthStatusOk)
	}
}

func TestIntegrationSelect(t *testing.T) {
	ds := integrationDatasource(t)

	res := runQuery(t, ds, fmt.Sprintf(`
		PREFIX foaf: <http://xmlns.com/foaf/0.1/>
		SELECT ?name WHERE {
			GRAPH <%s> { ?person foaf:name ?name }
		} ORDER BY ?name`, testGraph))
	if res.Error != nil {
		t.Fatalf("query returned an error: %v", res.Error)
	}

	frame := res.Frames[0]
	if len(frame.Fields) != 1 || frame.Fields[0].Name != "name" {
		t.Fatalf("got %d fields, want a single 'name' field", len(frame.Fields))
	}

	got := fieldValues(t, frame.Fields[0])
	want := []string{"Alice", "Bob", "Carol"}
	if len(got) != len(want) {
		t.Fatalf("names = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("names = %v, want %v", got, want)
		}
	}
}

// Carol has no age, so the ?age column must hold a null value for her rather
// than shifting the remaining values up by one row.
func TestIntegrationSelectWithOptional(t *testing.T) {
	ds := integrationDatasource(t)

	res := runQuery(t, ds, fmt.Sprintf(`
		PREFIX ex:   <http://example.org/>
		PREFIX foaf: <http://xmlns.com/foaf/0.1/>
		SELECT ?name ?age WHERE {
			GRAPH <%s> {
				?person foaf:name ?name .
				OPTIONAL { ?person ex:age ?age }
			}
		} ORDER BY ?name`, testGraph))
	if res.Error != nil {
		t.Fatalf("query returned an error: %v", res.Error)
	}

	frame := res.Frames[0]
	rows, err := frame.RowLen()
	if err != nil {
		t.Fatalf("the frame has mismatched field lengths: %v", err)
	}
	if rows != 3 {
		t.Fatalf("got %d rows, want 3", rows)
	}

	ages := fieldValues(t, frame.Fields[1])
	want := []string{"34", "28", "<nil>"}
	for i := range want {
		if ages[i] != want[i] {
			t.Fatalf("ages = %v, want %v", ages, want)
		}
	}
}

func TestIntegrationAsk(t *testing.T) {
	ds := integrationDatasource(t)

	res := runQuery(t, ds, fmt.Sprintf(`
		PREFIX foaf: <http://xmlns.com/foaf/0.1/>
		ASK { GRAPH <%s> { ?person foaf:name "Alice" } }`, testGraph))
	if res.Error != nil {
		t.Fatalf("query returned an error: %v", res.Error)
	}

	frame := res.Frames[0]
	if len(frame.Fields) != 1 || frame.Fields[0].Name != "boolean" {
		t.Fatalf("got %d fields, want a single 'boolean' field", len(frame.Fields))
	}
	if value, ok := frame.Fields[0].ConcreteAt(0); !ok || value != true {
		t.Errorf("boolean field = %v, want true", value)
	}
}

func TestIntegrationConstruct(t *testing.T) {
	ds := integrationDatasource(t)

	res := runQuery(t, ds, fmt.Sprintf(`
		PREFIX foaf: <http://xmlns.com/foaf/0.1/>
		CONSTRUCT { ?person foaf:name ?name }
		WHERE { GRAPH <%s> { ?person foaf:name ?name } }`, testGraph))
	if res.Error != nil {
		t.Fatalf("query returned an error: %v", res.Error)
	}

	frame := res.Frames[0]
	rows, err := frame.RowLen()
	if err != nil {
		t.Fatalf("the frame has mismatched field lengths: %v", err)
	}
	if rows != 3 {
		t.Fatalf("got %d triples, want 3", rows)
	}

	for i, want := range []string{"subject", "predicate", "object"} {
		if frame.Fields[i].Name != want {
			t.Errorf("field %d is named %q, want %q", i, frame.Fields[i].Name, want)
		}
	}
}

func TestIntegrationMalformedQuery(t *testing.T) {
	ds := integrationDatasource(t)

	res := runQuery(t, ds, "SELECT * WHERE { this is not SPARQL")
	if res.Error == nil {
		t.Fatal("a malformed query must produce an error response")
	}
}
