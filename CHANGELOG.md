# @ludovicm67/sparql-datasource

## 0.2.1

### Patch Changes

- fee3399: Upgrade every dependency to its latest version.

  - Build against Grafana 13.2 and, as its packages now require it, React 19. React and the Grafana packages are provided by the running Grafana instance, so the plugin keeps supporting Grafana 12 and later.
  - Update the Go backend to Go 1.27 and `grafana-plugin-sdk-go` v0.296.4.
  - Update Changesets to version 3 and `changesets/action` to version 2 in the release workflow.
  - Keep ESLint on 9, `@grafana/eslint-config` on 9 and TypeScript on 5.9, which the Grafana tooling still requires.

## 0.2.0

### Minor Changes

- 305ca33: Update the plugin to the current Grafana tooling and add tests against a real SPARQL endpoint.

  - Regenerate the `@grafana/create-plugin` scaffolding (5.x → 7.9.0) and build against Grafana 13.
  - Update every frontend and backend dependency to its latest version, except where an upstream constraint blocks it: React stays on 18 (peer dependency of `@grafana/ui` 13), ESLint on 9 (`eslint-plugin-react` supports no more), TypeScript on 5.9 (`typescript-eslint` supports no more) and `@grafana/eslint-config` on 9 (version 10 drops the `flat.js` entry point the scaffolding imports).
  - Update the Go backend to `grafana-plugin-sdk-go` v0.294.0.
  - Replace the Cypress end-to-end setup with `@grafana/plugin-e2e` and Playwright.
  - Add an Oxigraph SPARQL endpoint to the development stack, seeded with `testdata/seed.ttl`, and run the backend and end-to-end tests against it in CI.
  - Fix `SELECT` queries with unbound variables producing frames with mismatched column lengths.
  - Fix the query form detection mistaking a keyword inside a comment, an IRI, a string literal or a prefixed name for the query form.
  - Report an endpoint that holds no data as healthy, and surface health check failures as a message instead of a plugin error.

## 0.1.0

### Minor Changes

- bc8ca19: Let the user configure the query timeout value

## 0.0.1

### Patch Changes

- 87b675a: Fix order of columns

## 0.0.0

### Major Changes

- d86d6b5: Initial release
