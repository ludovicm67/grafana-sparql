package plugin

import "strings"

// QueryForm is the form of a SPARQL query, as defined by the SPARQL 1.1 Query
// Language specification.
type QueryForm string

const (
	QueryFormSelect    QueryForm = "SELECT"
	QueryFormConstruct QueryForm = "CONSTRUCT"
	QueryFormAsk       QueryForm = "ASK"
	QueryFormDescribe  QueryForm = "DESCRIBE"
	// QueryFormUnknown is returned when no query form keyword could be found.
	QueryFormUnknown QueryForm = ""
)

var queryForms = []QueryForm{QueryFormSelect, QueryFormConstruct, QueryFormAsk, QueryFormDescribe}

// DetectQueryForm returns the form of the given SPARQL query.
//
// The query is scanned token by token so that occurrences of a query form
// keyword inside comments, IRIs, string literals or prefixed names are not
// mistaken for the actual query form. For example the `SELECT` in
// `PREFIX ex: <http://example.org/select#> ASK { ?s ex:p ?o }` is part of an
// IRI, and the query form is `ASK`.
func DetectQueryForm(query string) QueryForm {
	runes := []rune(query)

	for i := 0; i < len(runes); i++ {
		switch runes[i] {
		case '#': // comment, runs until the end of the line
			for i < len(runes) && runes[i] != '\n' {
				i++
			}
		case '<': // IRI reference, runs until the closing '>'
			for i < len(runes) && runes[i] != '>' && runes[i] != '\n' {
				i++
			}
		case '"', '\'': // string literal
			i = skipLiteral(runes, i)
		default:
			if !isWordRune(runes[i]) {
				continue
			}

			end := i
			for end < len(runes) && isWordRune(runes[end]) {
				end++
			}

			word := strings.ToUpper(string(runes[i:end]))
			for _, form := range queryForms {
				if word == string(form) {
					return form
				}
			}

			i = end - 1
		}
	}

	return QueryFormUnknown
}

// ReturnsGraph reports whether the query form produces an RDF graph
// (a set of triples) rather than a solution sequence or a boolean.
func (f QueryForm) ReturnsGraph() bool {
	return f == QueryFormConstruct || f == QueryFormDescribe
}

// skipLiteral returns the index of the closing quote of the string literal
// starting at index start, or the index of the last rune when the literal is
// not terminated.
func skipLiteral(runes []rune, start int) int {
	quote := runes[start]

	// Long string literals are delimited by three quote characters.
	quoteLen := 1
	if start+2 < len(runes) && runes[start+1] == quote && runes[start+2] == quote {
		quoteLen = 3
	}

	for i := start + quoteLen; i < len(runes); i++ {
		if runes[i] == '\\' {
			i++ // the next rune is escaped, skip it
			continue
		}
		if runes[i] != quote {
			continue
		}
		if quoteLen == 1 {
			return i
		}
		if i+2 < len(runes) && runes[i+1] == quote && runes[i+2] == quote {
			return i + 2
		}
	}

	return len(runes) - 1
}

// isWordRune reports whether r can be part of a SPARQL keyword, variable or
// prefixed name. ':' is included on purpose so that a prefixed name such as
// `ex:select` is read as a single token and never matches a query form keyword.
func isWordRune(r rune) bool {
	switch {
	case r >= 'a' && r <= 'z':
		return true
	case r >= 'A' && r <= 'Z':
		return true
	case r >= '0' && r <= '9':
		return true
	case r == '_' || r == ':' || r == '-' || r == '?' || r == '$':
		return true
	default:
		return false
	}
}
