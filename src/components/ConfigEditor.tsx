import React, { ChangeEvent } from 'react';
import { InlineField, Input, SecretInput } from '@grafana/ui';
import { DataSourcePluginOptionsEditorProps } from '@grafana/data';
import { MyDataSourceOptions, MySecureJsonData } from '../types';

interface Props extends DataSourcePluginOptionsEditorProps<MyDataSourceOptions, MySecureJsonData> {}

export function ConfigEditor(props: Props) {
  const { onOptionsChange, options } = props;
  const { jsonData, secureJsonFields, secureJsonData } = options;

  const onEndpointChange = (event: ChangeEvent<HTMLInputElement>) => {
    onOptionsChange({
      ...options,
      jsonData: { ...jsonData, endpoint: event.target.value },
    });
  };

  const onUsernameChange = (event: ChangeEvent<HTMLInputElement>) => {
    onOptionsChange({
      ...options,
      jsonData: { ...jsonData, username: event.target.value },
    });
  };

  const onTimeoutChange = (event: ChangeEvent<HTMLInputElement>) => {
    onOptionsChange({
      ...options,
      jsonData: { ...jsonData, timeout: event.target.value },
    });
  };

  // Secure field (only sent to the backend)
  const onPasswordChange = (event: ChangeEvent<HTMLInputElement>) => {
    onOptionsChange({
      ...options,
      secureJsonData: { ...secureJsonData, password: event.target.value },
    });
  };

  const onResetPassword = () => {
    onOptionsChange({
      ...options,
      secureJsonFields: { ...secureJsonFields, password: false },
      secureJsonData: { ...secureJsonData, password: '' },
    });
  };

  return (
    <>
      <InlineField
        label="SPARQL Endpoint"
        labelWidth={20}
        interactive
        tooltip="URL of the SPARQL query endpoint, for example https://query.wikidata.org/sparql"
      >
        <Input
          id="config-editor-endpoint"
          required
          onChange={onEndpointChange}
          value={jsonData.endpoint || ''}
          placeholder="https://example.com/query"
          width={40}
        />
      </InlineField>
      <InlineField label="Username" labelWidth={20} interactive tooltip="Optional user name for HTTP digest auth">
        <Input
          id="config-editor-username"
          onChange={onUsernameChange}
          value={jsonData.username || ''}
          placeholder="admin"
          width={40}
        />
      </InlineField>
      <InlineField label="Password" labelWidth={20} interactive tooltip="Optional password for HTTP digest auth">
        <SecretInput
          id="config-editor-password"
          isConfigured={Boolean(secureJsonFields?.password)}
          onReset={onResetPassword}
          onChange={onPasswordChange}
          value={secureJsonData?.password || ''}
          placeholder="super-secret-password"
          width={40}
          autoComplete="new-password"
        />
      </InlineField>
      <InlineField label="Query Timeout" labelWidth={20} interactive tooltip="Query timeout in milliseconds">
        <Input
          id="config-editor-timeout"
          type="number"
          onChange={onTimeoutChange}
          value={jsonData.timeout || ''}
          placeholder="30000"
          width={40}
          suffix="ms"
        />
      </InlineField>
    </>
  );
}
