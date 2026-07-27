import { CoreApp, DataSourceInstanceSettings, PluginType } from '@grafana/data';

import { DataSource } from './datasource';
import { DEFAULT_QUERY, MyDataSourceOptions } from './types';

const replace = jest.fn((value?: string) => value?.replace('$org', 'example.org'));

jest.mock('@grafana/runtime', () => ({
  ...jest.requireActual('@grafana/runtime'),
  getTemplateSrv: () => ({ replace }),
}));

function createDataSource(jsonData: MyDataSourceOptions = {}): DataSource {
  const settings = {
    id: 1,
    uid: 'sparql-test',
    name: 'SPARQL',
    type: 'ludovicm67-sparql-datasource',
    access: 'proxy',
    readOnly: false,
    jsonData,
    meta: { type: PluginType.datasource },
  } as DataSourceInstanceSettings<MyDataSourceOptions>;

  return new DataSource(settings);
}

describe('DataSource', () => {
  beforeEach(() => {
    replace.mockClear();
  });

  it('offers a default query', () => {
    expect(createDataSource().getDefaultQuery(CoreApp.PanelEditor)).toEqual(DEFAULT_QUERY);
  });

  it('interpolates dashboard variables in the query', () => {
    const ds = createDataSource();

    const query = ds.applyTemplateVariables({ refId: 'A', queryText: 'SELECT * WHERE { ?s <http://$org/p> ?o }' }, {});

    expect(query.queryText).toBe('SELECT * WHERE { ?s <http://example.org/p> ?o }');
  });

  it('leaves a query without a text untouched', () => {
    const ds = createDataSource();

    expect(ds.applyTemplateVariables({ refId: 'A' }, {}).queryText).toBeUndefined();
    expect(replace).not.toHaveBeenCalled();
  });

  it.each([
    ['SELECT * WHERE { ?s ?p ?o }', true],
    ['', false],
    ['   \n  ', false],
    [undefined, false],
  ])('only runs non-empty queries (%p)', (queryText, expected) => {
    expect(createDataSource().filterQuery({ refId: 'A', queryText })).toBe(expected);
  });
});
