import { test, expect } from '@grafana/plugin-e2e';
import { LOCAL_DATASOURCE, setSparqlQuery } from './utils';

test('smoke: should render the query editor', async ({ panelEditPage, readProvisionedDataSource, selectors }) => {
  const ds = await readProvisionedDataSource({ fileName: 'datasources.yml', name: LOCAL_DATASOURCE });
  await panelEditPage.datasource.set(ds.name);

  await expect(panelEditPage.getByGrafanaSelector(selectors.components.CodeEditor.container)).toBeVisible();
  await expect(panelEditPage.getQueryEditorRow('A').getByRole('button', { name: 'Run Query' })).toBeVisible();
});

test('a SELECT query should return the data of the SPARQL endpoint', async ({
  panelEditPage,
  page,
  readProvisionedDataSource,
  selectors,
}) => {
  const ds = await readProvisionedDataSource({ fileName: 'datasources.yml', name: LOCAL_DATASOURCE });
  await panelEditPage.datasource.set(ds.name);
  await panelEditPage.setVisualization('Table');

  await setSparqlQuery(
    page,
    panelEditPage.getByGrafanaSelector(selectors.components.CodeEditor.container),
    'SELECT ?name WHERE { ?person <http://xmlns.com/foaf/0.1/name> ?name } ORDER BY ?name'
  );

  await expect(panelEditPage.refreshPanel()).toBeOK();
  await expect(panelEditPage.panel.fieldNames).toContainText(['name']);
  await expect(panelEditPage.panel.data).toContainText(['Alice', 'Bob', 'Carol']);
});

test('an ASK query should return a boolean', async ({ panelEditPage, page, readProvisionedDataSource, selectors }) => {
  const ds = await readProvisionedDataSource({ fileName: 'datasources.yml', name: LOCAL_DATASOURCE });
  await panelEditPage.datasource.set(ds.name);
  await panelEditPage.setVisualization('Table');

  await setSparqlQuery(
    page,
    panelEditPage.getByGrafanaSelector(selectors.components.CodeEditor.container),
    'ASK WHERE { ?s ?p ?o }'
  );

  await expect(panelEditPage.refreshPanel()).toBeOK();
  await expect(panelEditPage.panel.fieldNames).toContainText(['boolean']);
  await expect(panelEditPage.panel.data).toContainText(['true']);
});

test('a CONSTRUCT query should return triples', async ({
  panelEditPage,
  page,
  readProvisionedDataSource,
  selectors,
}) => {
  const ds = await readProvisionedDataSource({ fileName: 'datasources.yml', name: LOCAL_DATASOURCE });
  await panelEditPage.datasource.set(ds.name);
  await panelEditPage.setVisualization('Table');

  await setSparqlQuery(
    page,
    panelEditPage.getByGrafanaSelector(selectors.components.CodeEditor.container),
    'CONSTRUCT { ?person <http://xmlns.com/foaf/0.1/name> ?name } WHERE { ?person <http://xmlns.com/foaf/0.1/name> ?name }'
  );

  await expect(panelEditPage.refreshPanel()).toBeOK();
  await expect(panelEditPage.panel.fieldNames).toContainText(['subject', 'predicate', 'object']);
  await expect(panelEditPage.panel.data).toContainText(['http://example.org/alice']);
});

test('a malformed query should surface the error of the endpoint', async ({
  panelEditPage,
  page,
  readProvisionedDataSource,
  selectors,
}) => {
  const ds = await readProvisionedDataSource({ fileName: 'datasources.yml', name: LOCAL_DATASOURCE });
  await panelEditPage.datasource.set(ds.name);
  await panelEditPage.setVisualization('Table');

  await setSparqlQuery(
    page,
    panelEditPage.getByGrafanaSelector(selectors.components.CodeEditor.container),
    'SELECT * WHERE { this is not SPARQL'
  );

  // Assert on the state the panel settles in rather than on a single response,
  // so that a query request still in flight cannot decide the outcome.
  await panelEditPage.refreshPanel();
  await expect(panelEditPage.panel.getErrorIcon()).toBeVisible();
});
