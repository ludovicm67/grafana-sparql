import { test, expect } from '@grafana/plugin-e2e';
import { MyDataSourceOptions, MySecureJsonData } from '../src/types';
import { LOCAL_DATASOURCE } from './utils';

test('smoke: should render the config editor', async ({
  createDataSourceConfigPage,
  readProvisionedDataSource,
  page,
}) => {
  const ds = await readProvisionedDataSource({ fileName: 'datasources.yml', name: LOCAL_DATASOURCE });
  await createDataSourceConfigPage({ type: ds.type });

  await expect(page.getByRole('textbox', { name: 'SPARQL Endpoint' })).toBeVisible();
  await expect(page.getByRole('textbox', { name: 'Username' })).toBeVisible();
  await expect(page.getByRole('spinbutton', { name: 'Query Timeout' })).toBeVisible();
});

test('"Save & test" should succeed against a reachable SPARQL endpoint', async ({
  createDataSourceConfigPage,
  readProvisionedDataSource,
  page,
}) => {
  const ds = await readProvisionedDataSource<MyDataSourceOptions, MySecureJsonData>({
    fileName: 'datasources.yml',
    name: LOCAL_DATASOURCE,
  });
  const configPage = await createDataSourceConfigPage({ type: ds.type });

  await page.getByRole('textbox', { name: 'SPARQL Endpoint' }).fill(ds.jsonData.endpoint ?? '');

  await expect(configPage.saveAndTest()).toBeOK();
  await expect(configPage).toHaveAlert('success', { hasText: 'SPARQL endpoint is healthy' });
});

test('"Save & test" should fail when no endpoint is configured', async ({
  createDataSourceConfigPage,
  readProvisionedDataSource,
}) => {
  const ds = await readProvisionedDataSource({ fileName: 'datasources.yml', name: LOCAL_DATASOURCE });
  const configPage = await createDataSourceConfigPage({ type: ds.type });

  await expect(configPage.saveAndTest()).not.toBeOK();
});

test('"Save & test" should fail when the endpoint does not answer', async ({
  createDataSourceConfigPage,
  readProvisionedDataSource,
  page,
}) => {
  const ds = await readProvisionedDataSource({ fileName: 'datasources.yml', name: LOCAL_DATASOURCE });
  const configPage = await createDataSourceConfigPage({ type: ds.type });

  await page.getByRole('textbox', { name: 'SPARQL Endpoint' }).fill('http://localhost:1/not-a-sparql-endpoint');

  await expect(configPage.saveAndTest()).not.toBeOK();
  await expect(configPage).toHaveAlert('error', { hasText: 'Failed to execute health check query' });
});
