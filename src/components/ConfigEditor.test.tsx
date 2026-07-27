import React from 'react';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { DataSourcePluginOptionsEditorProps, DataSourceSettings, PluginType } from '@grafana/data';

import { ConfigEditor } from './ConfigEditor';
import { MyDataSourceOptions, MySecureJsonData } from '../types';

type Props = DataSourcePluginOptionsEditorProps<MyDataSourceOptions, MySecureJsonData>;

function setup(overrides: Partial<DataSourceSettings<MyDataSourceOptions, MySecureJsonData>> = {}) {
  const onOptionsChange = jest.fn();
  const options = {
    id: 1,
    uid: 'sparql-test',
    name: 'SPARQL',
    type: 'ludovicm67-sparql-datasource',
    access: 'proxy',
    jsonData: {},
    secureJsonData: {},
    secureJsonFields: {},
    typeLogoUrl: '',
    typeName: 'SPARQL',
    url: '',
    user: '',
    database: '',
    basicAuth: false,
    basicAuthUser: '',
    isDefault: false,
    withCredentials: false,
    readOnly: false,
    orgId: 1,
    version: 1,
    meta: { type: PluginType.datasource },
    ...overrides,
  } as DataSourceSettings<MyDataSourceOptions, MySecureJsonData>;

  render(<ConfigEditor {...({ options, onOptionsChange } as unknown as Props)} />);

  return { onOptionsChange };
}

describe('ConfigEditor', () => {
  it('renders every option', () => {
    setup();

    expect(screen.getByRole('textbox', { name: /SPARQL Endpoint/ })).toBeInTheDocument();
    expect(screen.getByRole('textbox', { name: /Username/ })).toBeInTheDocument();
    expect(screen.getByRole('spinbutton', { name: /Query Timeout/ })).toBeInTheDocument();
  });

  it('reports the endpoint through jsonData', async () => {
    const { onOptionsChange } = setup();

    await userEvent.type(screen.getByRole('textbox', { name: /SPARQL Endpoint/ }), 'h');

    expect(onOptionsChange).toHaveBeenCalledWith(
      expect.objectContaining({ jsonData: expect.objectContaining({ endpoint: 'h' }) })
    );
  });

  it('reports the password through secureJsonData, never through jsonData', async () => {
    const { onOptionsChange } = setup();

    await userEvent.type(screen.getByLabelText(/Password/), 's');

    const [update] = onOptionsChange.mock.calls[0];
    expect(update.secureJsonData).toEqual({ password: 's' });
    expect(update.jsonData).not.toHaveProperty('password');
  });

  it('hides a configured password behind a reset button', () => {
    setup({ secureJsonFields: { password: true } });

    expect(screen.getByRole('button', { name: /Reset/ })).toBeInTheDocument();
  });
});
