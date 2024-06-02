import React from 'react';
import { Button, InlineField, InlineFieldRow, CodeEditor, VerticalGroup } from '@grafana/ui';
import { QueryEditorProps } from '@grafana/data';
import { DataSource } from '../datasource';
import { MyDataSourceOptions, MyQuery } from '../types';

type Props = QueryEditorProps<DataSource, MyQuery, MyDataSourceOptions>;

export function QueryEditor({ query, onChange, onRunQuery }: Props) {
  const onQueryTextChange = (value: string) => {
    onChange({ ...query, queryText: value });
  };

  const { queryText } = query;

  return (
    <div className="gf-form">
      <VerticalGroup spacing="sm">
        <InlineFieldRow style={{ width: '100%' }}>
          <InlineField grow style={{ width: '100%' }}>
            <CodeEditor
              onBlur={onQueryTextChange}
              onChange={onQueryTextChange}
              onSave={onQueryTextChange}
              language="sparql"
              height={250}
              width={'100%'}
              value={queryText || ''}
              showLineNumbers={true}
              showMiniMap={true}
            />
          </InlineField>
        </InlineFieldRow>
        <InlineFieldRow>
          <Button variant="secondary" onClick={onRunQuery}>
            Run Query
          </Button>
        </InlineFieldRow>
      </VerticalGroup>
    </div>
  );
}
