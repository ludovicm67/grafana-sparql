# SPARQL Datasource for Grafana

![CI](https://github.com/ludovicm67/grafana-sparql/workflows/CI/badge.svg)

## Overview

This plugin allows you to connect to a [SPARQL](https://www.w3.org/TR/sparql11-protocol/) endpoint and visualize the data in [Grafana](https://grafana.com/).

[SPARQL](https://www.w3.org/TR/sparql11-protocol/) is a query language and protocol for querying RDF data, which is a standard model for data interchange on the Web.

## Screenshots

### Data source configuration

The data source configuration is quite simple:

- **Name**: The name of the data source.
- **SPARQL Endpoint**: The URL of the SPARQL endpoint.
- **Username**: The username to use for authentication (optional).
- **Password**: The password to use for authentication (optional).

![Data source configuration](https://raw.githubusercontent.com/ludovicm67/grafana-sparql/main/src/img/screenshots/datasource-configuration.png)

### Query editor

The query editor allows you to write SPARQL queries and visualize the results.

Here is an example of a basic query run against the [DBpedia](https://dbpedia.org/sparql) SPARQL endpoint:

![Query editor](https://raw.githubusercontent.com/ludovicm67/grafana-sparql/main/src/img/screenshots/query-editor.png)

## Contributing

The source code of this plugin is hosted on [GitHub](https://github.com/ludovicm67/grafana-sparql).

Pull requests are welcome!
