import React from 'react';
import { Button, CodeEditor, Stack } from '@grafana/ui';
import { QueryEditorProps } from '@grafana/data';
import { DataSource } from '../datasource';
import { MyDataSourceOptions, MyQuery } from '../types';

type Props = QueryEditorProps<DataSource, MyQuery, MyDataSourceOptions>;

export const QUERY_EDITOR_ARIA_LABEL = 'SPARQL query';

export function QueryEditor({ query, onChange, onRunQuery }: Props) {
  const onQueryTextChange = (value: string) => {
    onChange({ ...query, queryText: value });
  };

  const { queryText } = query;

  return (
    <Stack direction="column" gap={1}>
      <CodeEditor
        onBlur={onQueryTextChange}
        onChange={onQueryTextChange}
        onSave={onQueryTextChange}
        language="sparql"
        height={250}
        width="100%"
        value={queryText || ''}
        showLineNumbers={true}
        showMiniMap={true}
        monacoOptions={{ ariaLabel: QUERY_EDITOR_ARIA_LABEL }}
      />
      <Stack direction="row">
        <Button variant="secondary" onClick={onRunQuery}>
          Run Query
        </Button>
      </Stack>
    </Stack>
  );
}
