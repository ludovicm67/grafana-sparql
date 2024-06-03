import React, { ChangeEvent } from 'react';
import { InlineField, Input, SecretInput } from '@grafana/ui';
import { DataSourcePluginOptionsEditorProps } from '@grafana/data';
import { MyDataSourceOptions, MySecureJsonData } from '../types';

interface Props extends DataSourcePluginOptionsEditorProps<MyDataSourceOptions> {}

export function ConfigEditor(props: Props) {
  const { onOptionsChange, options } = props;

  const onEndpointChange = (event: ChangeEvent<HTMLInputElement>) => {
    const jsonData = {
      ...options.jsonData,
      endpoint: event.target.value,
    };
    onOptionsChange({ ...options, jsonData });
  };

  const onUsernameChange = (event: ChangeEvent<HTMLInputElement>) => {
    const jsonData = {
      ...options.jsonData,
      username: event.target.value,
    };
    onOptionsChange({ ...options, jsonData });
  };

  const onTimeoutChange = (event: ChangeEvent<HTMLInputElement>) => {
    const jsonData = {
      ...options.jsonData,
      timeout: event.target.value,
    };
    onOptionsChange({ ...options, jsonData });
  };

  // Secure field (only sent to the backend)
  const onPasswordChange = (event: ChangeEvent<HTMLInputElement>) => {
    onOptionsChange({
      ...options,
      secureJsonData: {
        password: event.target.value,
      },
    });
  };

  const onResetPassword = () => {
    onOptionsChange({
      ...options,
      secureJsonFields: {
        ...options.secureJsonFields,
        password: false,
      },
      secureJsonData: {
        ...options.secureJsonData,
        password: '',
      },
    });
  };

  const { jsonData, secureJsonFields } = options;
  const secureJsonData = (options.secureJsonData || {}) as MySecureJsonData;

  return (
    <div className="gf-form-group">
      <InlineField label="SPARQL Endpoint" labelWidth={16}>
        <Input
          onChange={onEndpointChange}
          value={jsonData.endpoint || ''}
          placeholder="https://example.com/query"
          width={40}
        />
      </InlineField>
      <InlineField label="Username" labelWidth={16}>
        <Input onChange={onUsernameChange} value={jsonData.username || ''} placeholder="admin" width={40} />
      </InlineField>
      <InlineField label="Password" labelWidth={16}>
        <SecretInput
          isConfigured={(secureJsonFields && secureJsonFields.password) as boolean}
          onReset={onResetPassword}
          onChange={onPasswordChange}
          value={secureJsonData.password || ''}
          placeholder="super-secret-password"
          width={40}
          autoComplete="new-password"
        />
      </InlineField>
      <InlineField label="Query Timeout" labelWidth={16}>
        <Input
          type="number"
          onChange={onTimeoutChange}
          value={jsonData.timeout || ''}
          placeholder="30000"
          width={40}
          suffix="ms"
        />
      </InlineField>
    </div>
  );
}
