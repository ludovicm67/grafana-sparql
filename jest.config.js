// force timezone to UTC to allow tests to work regardless of local timezone
// generally used by snapshots, but can affect specific tests
process.env.TZ = 'UTC';

const { grafanaESModules, nodeModulesToTransform } = require('./.config/jest/utils');

module.exports = {
  // Jest configuration provided by Grafana scaffolding
  ...require('./.config/jest.config'),
  // ESM-only dependencies of the Grafana 13.2 packages that the scaffolding
  // does not know about yet, so that Jest transforms them instead of failing
  // to parse them.
  transformIgnorePatterns: [nodeModulesToTransform([...grafanaESModules, '@react-hookz/web', '@ver0/deep-equal'])],
};
