---
'@ludovicm67/sparql-datasource': patch
---

Upgrade every dependency to its latest version.

- Build against Grafana 13.2 and, as its packages now require it, React 19. React and the Grafana packages are provided by the running Grafana instance, so the plugin keeps supporting Grafana 12 and later.
- Update the Go backend to Go 1.27 and `grafana-plugin-sdk-go` v0.296.4.
- Update Changesets to version 3 and `changesets/action` to version 2 in the release workflow.
- Keep ESLint on 9, `@grafana/eslint-config` on 9 and TypeScript on 5.9, which the Grafana tooling still requires.
