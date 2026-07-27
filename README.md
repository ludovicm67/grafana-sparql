# Grafana SPARQL Data Source plugin

This plugin allows you to connect to a SPARQL endpoint and visualize the data in Grafana.

## Installation

If you are using the official Grafana Docker image, you can install this plugin by configuring the following environment variables:

- `GF_PLUGINS_ALLOW_LOADING_UNSIGNED_PLUGINS`: `ludovicm67-sparql-datasource`
- `GF_INSTALL_PLUGINS`: `${PATH_TO_ZIP_ARCHIVE};ludovicm67-sparql-datasource`

where `${PATH_TO_ZIP_ARCHIVE}` is the path to the zip archive of the plugin.

You can browse the [latest releases of the plugin](https://github.com/ludovicm67/grafana-sparql/releases) to find the link to the zip archive.

## Development

Requirements: Node.js (see [`.nvmrc`](.nvmrc)), Go (see the `go` directive in [`go.mod`](go.mod)), [Mage](https://magefile.org/) and Docker.

Run the following commands to get started:

- `npm install` to install frontend dependencies.
- `npm run dev` to build (and watch) the plugin frontend code.
- `mage -v` to build the plugin backend code. Rerun this command every time you edit your backend files.
- `npm run server` to start a Grafana development server. Restart this command every time you run mage, to run your new backend code.
- Open <http://localhost:3000> in your browser to create a dashboard and begin developing the plugin.

`npm run server` starts two containers:

- **Grafana**, on <http://localhost:3000>, with this plugin installed.
- **[Oxigraph](https://github.com/oxigraph/oxigraph)**, a local SPARQL endpoint on <http://localhost:7878/query>, seeded with
  [`testdata/seed.ttl`](testdata/seed.ttl).

Several data sources are provisioned, including `SPARQL - Local` pointing at the
local Oxigraph endpoint, plus the public Wikidata and DBpedia endpoints. A
`SPARQL examples` dashboard is provisioned as well. Take a look at the files in
the [`provisioning`](provisioning) directory to help you get started.

Run `npm run server:down` to stop the stack and drop its data.

## Testing

| Command                                      | What it runs                                                          |
| -------------------------------------------- | --------------------------------------------------------------------- |
| `npm run test:ci`                            | Frontend unit tests (Jest).                                           |
| `npm run typecheck` / `npm run lint`         | TypeScript and ESLint checks.                                         |
| `go test ./pkg/...`                          | Backend unit tests, against a stubbed SPARQL endpoint.                |
| `SPARQL_TEST_ENDPOINT=... go test ./pkg/...` | Backend tests, additionally against a real SPARQL endpoint.           |
| `npm run e2e`                                | End-to-end tests (Playwright), against the running development stack. |

### Backend tests against a real SPARQL endpoint

The backend integration tests are skipped unless `SPARQL_TEST_ENDPOINT` is set.
They insert their own data into a dedicated named graph and drop it afterwards,
so they can run against any writable SPARQL 1.1 endpoint:

```sh
docker compose up -d oxigraph
SPARQL_TEST_ENDPOINT=http://localhost:7878/query go test ./pkg/...
```

The update endpoint is derived from `SPARQL_TEST_ENDPOINT` by replacing a
trailing `/query` with `/update`; set `SPARQL_TEST_UPDATE_ENDPOINT` when your
endpoint does not follow that convention.

### End-to-end tests

The end-to-end tests drive a real Grafana that queries the local Oxigraph
endpoint, so they need both the plugin build and the development stack:

```sh
npm run build
mage -v buildAll
ANONYMOUS_AUTH_ENABLED=false DEVELOPMENT=false docker compose up -d --build
npm exec playwright install chromium
npm run e2e
```

`ANONYMOUS_AUTH_ENABLED=false` is required, because the tests log in as the
`admin` user.

To run them against another Grafana version, set `GRAFANA_VERSION` (and
optionally `GRAFANA_IMAGE`) before starting the stack. CI runs them against
every Grafana version supported by the `grafanaDependency` declared in
[`src/plugin.json`](src/plugin.json).
