---
'@ludovicm67/sparql-datasource': minor
---

Update the plugin to the current Grafana tooling and add tests against a real SPARQL endpoint.

- Regenerate the `@grafana/create-plugin` scaffolding (5.x → 7.9.0) and build against Grafana 13.
- Update the Go backend to `grafana-plugin-sdk-go` v0.294.0.
- Replace the Cypress end-to-end setup with `@grafana/plugin-e2e` and Playwright.
- Add an Oxigraph SPARQL endpoint to the development stack, seeded with `testdata/seed.ttl`, and run the backend and end-to-end tests against it in CI.
- Fix `SELECT` queries with unbound variables producing frames with mismatched column lengths.
- Fix the query form detection mistaking a keyword inside a comment, an IRI, a string literal or a prefixed name for the query form.
- Report an endpoint that holds no data as healthy, and surface health check failures as a message instead of a plugin error.
