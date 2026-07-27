import { CoreApp, DataSourceInstanceSettings, ScopedVars } from '@grafana/data';
import { DataSourceWithBackend, getTemplateSrv } from '@grafana/runtime';

import { MyQuery, MyDataSourceOptions, DEFAULT_QUERY } from './types';

export class DataSource extends DataSourceWithBackend<MyQuery, MyDataSourceOptions> {
  constructor(instanceSettings: DataSourceInstanceSettings<MyDataSourceOptions>) {
    super(instanceSettings);
  }

  getDefaultQuery(_: CoreApp): Partial<MyQuery> {
    return DEFAULT_QUERY;
  }

  /**
   * Interpolate dashboard variables inside the SPARQL query before it is sent
   * to the backend.
   */
  applyTemplateVariables(query: MyQuery, scopedVars: ScopedVars): MyQuery {
    return {
      ...query,
      queryText: query.queryText ? getTemplateSrv().replace(query.queryText, scopedVars) : query.queryText,
    };
  }

  /**
   * Skip empty queries instead of sending them to the SPARQL endpoint.
   */
  filterQuery(query: MyQuery): boolean {
    return Boolean(query.queryText?.trim());
  }
}
