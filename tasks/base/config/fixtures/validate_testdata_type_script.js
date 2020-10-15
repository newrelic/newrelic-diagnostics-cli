"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const ServiceConfiguration_1 = require("./src/ServiceConfiguration");
exports.config = {
    app_name: [`${ServiceConfiguration_1.getConfigStore().service.name}: ${ServiceConfiguration_1.getConfigStore().service.environment}`],
    agent_enabled: ServiceConfiguration_1.getConfigStore().monitoring.newrelic && ServiceConfiguration_1.getConfigStore().monitoring.newrelic.enabled,
    license_key: ServiceConfiguration_1.getConfigStore().monitoring.newrelic.license,
    logging: {
        level: 'info'
    }
};

//# sourceMappingURL=newrelic.js.map