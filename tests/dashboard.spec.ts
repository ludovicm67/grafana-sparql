import { test, expect } from '@grafana/plugin-e2e';

/**
 * These tests run the provisioned dashboard end to end: Grafana loads the
 * plugin, the plugin queries the SPARQL endpoint started by docker-compose, and
 * the results have to reach the panels.
 */
test.describe('provisioned dashboard', () => {
  test('should render the results of a SELECT query', async ({ readProvisionedDashboard, gotoDashboardPage }) => {
    const dashboard = await readProvisionedDashboard({ fileName: 'sparql-examples.json' });
    const dashboardPage = await gotoDashboardPage(dashboard);

    const panel = await dashboardPage.getPanelByTitle('People (SELECT)');

    await expect(panel.fieldNames).toContainText(['name', 'age']);
    await expect(panel.data).toContainText(['Alice', '34', 'Bob', '28', 'Carol']);
  });

  test('should render the result of an ASK query', async ({ readProvisionedDashboard, gotoDashboardPage }) => {
    const dashboard = await readProvisionedDashboard({ fileName: 'sparql-examples.json' });
    const dashboardPage = await gotoDashboardPage(dashboard);

    const panel = await dashboardPage.getPanelByTitle('Endpoint holds data (ASK)');

    await expect(panel.fieldNames).toContainText(['boolean']);
    await expect(panel.data).toContainText(['true']);
  });

  test('should render the triples of a CONSTRUCT query', async ({ readProvisionedDashboard, gotoDashboardPage }) => {
    const dashboard = await readProvisionedDashboard({ fileName: 'sparql-examples.json' });
    const dashboardPage = await gotoDashboardPage(dashboard);

    const panel = await dashboardPage.getPanelByTitle('Names as triples (CONSTRUCT)');

    await expect(panel.fieldNames).toContainText(['subject', 'predicate', 'object']);
    await expect(panel.data).toContainText(['http://example.org/alice', 'http://xmlns.com/foaf/0.1/name']);
  });
});
