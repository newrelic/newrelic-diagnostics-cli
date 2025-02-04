const nconf = require('nconf');
const path = require('path');

let configFile = path.join(__dirname, '/config/config.json');
configuration = nconf
  .env({ separator: '__' })
  .file(configFile);

exports.config = {
    app_name: [`${configuration.get('service').name}: ${configuration.get('service').environment}`],
    agent_enabled: configuration.get('monitoring').newrelic && configuration.get('monitoring').newrelic.enabled,
    license_key: configuration.get('monitoring').newrelic.license,
    logging: {
        level: 'info'
    }
};