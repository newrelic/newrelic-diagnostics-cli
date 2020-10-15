package requirements

import (
	"fmt"
	"strings"

	log "github.com/newrelic/NrDiag/logger"
	"github.com/newrelic/NrDiag/tasks"
	dependencies "github.com/newrelic/NrDiag/tasks/node/env"
)

var modules = map[string][]string{
	"supported":   []string{"express", "hapi", "restify", "connect", "koa"},
	"unsupported": []string{"mongoose", "typescript", "@types/node", "@babel/core", "@babel/node", "@babel/generator", "webpack", "graphql-server-express", "apollo-server", "graphql", "webpack-node-externals"},
	"optional":    []string{"@newrelic/native-metrics"},
	"frontend":    []string{"react", "react-dom", "angular", "@angular/core", "@angular/common"},
}

type NodeRequirementsProblematicModules struct {
}

func (p NodeRequirementsProblematicModules) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Node/Requirements/ProblematicModules")
}

func (p NodeRequirementsProblematicModules) Explain() string {
	return "This task declares unsupported Node Agent technologies"
}
func (p NodeRequirementsProblematicModules) Dependencies() []string {
	return []string{
		"Node/Env/Dependencies",
	}
}

func (p NodeRequirementsProblematicModules) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {

	foundNodeDependencies := initializeTaskDependencies(upstream)

	if len(foundNodeDependencies) == 0 {
		return tasks.Result{
			Status:  tasks.None,
			Summary: "A list of Node modules was not found. This task did not run",
		}
	}

	foundMissingDataIssues := p.checkForMissingDataIssues(foundNodeDependencies)

	if len(foundMissingDataIssues) > 0 {
		return tasks.Result{
			Status:  tasks.Warning,
			Summary: foundMissingDataIssues,
			URL:     "https://docs.newrelic.com/docs/agents/nodejs-agent/getting-started/compatibility-requirements-nodejs-agent",
		}
	}

	return tasks.Result{
		Status:  tasks.Success,
		Summary: "Your Node modules are compatible with the Node Agent requirements",
	}

}

func initializeTaskDependencies(upstream map[string]tasks.Result) []dependencies.NodeModuleVersion {

	if upstream["Node/Env/Dependencies"].Status != tasks.Info {
		return []dependencies.NodeModuleVersion{}
	}

	modulesList, ok := upstream["Node/Env/Dependencies"].Payload.([]dependencies.NodeModuleVersion)
	if !ok {
		log.Debug("Type assertion failure")
		return []dependencies.NodeModuleVersion{}
	}

	return modulesList
}

func (p NodeRequirementsProblematicModules) isUsingSupportedFramework(foundDependencies []dependencies.NodeModuleVersion) bool {

	for _, dependency := range foundDependencies {
		moduleName := dependency.Module
		for _, val := range modules["supported"] {
			if val == moduleName {
				return true
			}
		}
	}
	return false
}

func checkForFrontendFrameworks(foundNodeDependencies []dependencies.NodeModuleVersion) []string {
	var frontendFrameworks []string
	for _, dependency := range foundNodeDependencies {
		moduleName := dependency.Module
		for _, val := range modules["frontend"] {
			if val == moduleName {
				frontendFrameworks = append(frontendFrameworks, val)
			}
		}
	}
	return frontendFrameworks
}

func isUsingNativeMetricsModule(foundDependencies []dependencies.NodeModuleVersion) bool {
	for _, dependency := range foundDependencies {
		moduleName := dependency.Module
		for _, val := range modules["optional"] {
			if val == moduleName {
				return true
			}
		}
	}
	return false
}

func checkForConflictiveModules(foundDependencies []dependencies.NodeModuleVersion) []string {
	var conflictiveModules []string
	for _, dependency := range foundDependencies {
		moduleName := dependency.Module
		for _, val := range modules["unsupported"] {
			if val == moduleName {
				conflictiveModules = append(conflictiveModules, val)
			}
		}
	}
	return conflictiveModules
}

func (p NodeRequirementsProblematicModules) checkForMissingDataIssues(foundDependencies []dependencies.NodeModuleVersion) string {

	var warningSummary string

	if !(p.isUsingSupportedFramework(foundDependencies)) {
		warningSummary += "- You are not using a supported framework by the Node Agent. In order to get monitoring data, you'll have to apply manual instrumentation using our APIs. For more information: https://docs.newrelic.com/docs/agents/nodejs-agent/supported-features/nodejs-custom-instrumentation\n"
	}

	foundFrontendFrameworks := checkForFrontendFrameworks(foundDependencies)
	if len(foundFrontendFrameworks) > 0 {
		warningSummary += fmt.Sprintf("- We noticed that you are using: %s. If you are looking to monitor a client side app, beware that the Node Agent only monitors server side frameworks. To get metrics for front-end libraries/frameworks use the Browser Agent instead: https://docs.newrelic.com/docs/browser/new-relic-browser/getting-started/compatibility-requirements-new-relic-browser\n", strings.Join(foundFrontendFrameworks, ", "))
	}

	foundConflictiveModules := checkForConflictiveModules(foundDependencies)
	if len(foundConflictiveModules) > 0 {
		warningSummary += fmt.Sprintf("- We have detected the following unsupported module(s) in your application: %s. This may cause instrumentation issues and inconsistency of data for the Node Agent.\n", strings.Join(foundConflictiveModules, ", "))
	}

	if !(isUsingNativeMetricsModule(foundDependencies)) {
		warningSummary += "- Keep in mind that if you are looking for additional Node.js runtime level statistics, you'll need to install our optional module: @newrelic/native-metrics. For more information: https://docs.newrelic.com/docs/agents/nodejs-agent/supported-features/nodejs-vm-measurements\n"
	}

	return warningSummary

}
