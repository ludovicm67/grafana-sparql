package plugin

import "testing"

func TestDetectQueryForm(t *testing.T) {
	tests := []struct {
		name  string
		query string
		want  QueryForm
	}{
		{
			name:  "empty query",
			query: "",
			want:  QueryFormUnknown,
		},
		{
			name:  "plain select",
			query: "SELECT * WHERE { ?s ?p ?o }",
			want:  QueryFormSelect,
		},
		{
			name:  "lowercase select",
			query: "select * where { ?s ?p ?o }",
			want:  QueryFormSelect,
		},
		{
			name:  "select after a prologue",
			query: "PREFIX ex: <http://example.org/>\nSELECT ?s WHERE { ?s ex:p ?o }",
			want:  QueryFormSelect,
		},
		{
			name:  "construct",
			query: "CONSTRUCT { ?s ?p ?o } WHERE { ?s ?p ?o }",
			want:  QueryFormConstruct,
		},
		{
			name:  "describe",
			query: "DESCRIBE <http://example.org/thing>",
			want:  QueryFormDescribe,
		},
		{
			name:  "ask",
			query: "ASK WHERE { ?s ?p ?o }",
			want:  QueryFormAsk,
		},
		{
			name:  "keyword inside a comment is ignored",
			query: "# CONSTRUCT is not what we want here\nSELECT * WHERE { ?s ?p ?o }",
			want:  QueryFormSelect,
		},
		{
			name:  "keyword inside an IRI is ignored",
			query: "PREFIX ex: <http://example.org/construct#> ASK { ?s ex:p ?o }",
			want:  QueryFormAsk,
		},
		{
			name:  "prefix declaration on a single line with a fragment IRI",
			query: "PREFIX ex: <http://example.org/ns#> SELECT * WHERE { ?s ex:p ?o }",
			want:  QueryFormSelect,
		},
		{
			name:  "keyword inside a string literal is ignored",
			query: `SELECT * WHERE { ?s ?p "CONSTRUCT a house" }`,
			want:  QueryFormSelect,
		},
		{
			name:  "keyword inside a long string literal is ignored",
			query: "CONSTRUCT { ?s ?p ?o } WHERE { ?s ?p \"\"\"a SELECT\nover\nlines\"\"\" }",
			want:  QueryFormConstruct,
		},
		{
			name:  "escaped quote inside a string literal",
			query: `SELECT * WHERE { ?s ?p "a \" DESCRIBE" }`,
			want:  QueryFormSelect,
		},
		{
			name:  "prefixed name that looks like a keyword is ignored",
			query: "PREFIX select: <http://example.org/>\nASK { ?s select:p ?o }",
			want:  QueryFormAsk,
		},
		{
			name:  "variable that looks like a keyword is ignored",
			query: "ASK { ?select ?p ?o }",
			want:  QueryFormAsk,
		},
		{
			name:  "no query form at all",
			query: "PREFIX ex: <http://example.org/>",
			want:  QueryFormUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DetectQueryForm(tt.query); got != tt.want {
				t.Errorf("DetectQueryForm(%q) = %q, want %q", tt.query, got, tt.want)
			}
		})
	}
}

func TestQueryFormReturnsGraph(t *testing.T) {
	graph := map[QueryForm]bool{
		QueryFormSelect:    false,
		QueryFormAsk:       false,
		QueryFormConstruct: true,
		QueryFormDescribe:  true,
		QueryFormUnknown:   false,
	}

	for form, want := range graph {
		if got := form.ReturnsGraph(); got != want {
			t.Errorf("QueryForm(%q).ReturnsGraph() = %v, want %v", form, got, want)
		}
	}
}
